package summarizer

import (
	"context"
	"net/http"
	"testing"

	"github.com/cockroachdb/errors"
	"github.com/mopemope/quicknews/scraper"
	openai "github.com/openai/openai-go/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/genai"
)

func TestRetryableError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil", err: nil, want: false},
		{name: "unknown error", err: errors.New("boom"), want: true},
		{name: "context canceled", err: context.Canceled, want: false},
		{name: "context deadline", err: context.DeadlineExceeded, want: false},
		{name: "openai 429", err: &openai.Error{StatusCode: http.StatusTooManyRequests}, want: true},
		{name: "openai 500", err: &openai.Error{StatusCode: http.StatusInternalServerError}, want: true},
		{name: "openai 401", err: &openai.Error{StatusCode: http.StatusUnauthorized}, want: false},
		{name: "openai 404", err: &openai.Error{StatusCode: http.StatusNotFound}, want: false},
		{name: "openai 408", err: &openai.Error{StatusCode: http.StatusRequestTimeout}, want: true},
		{name: "genai 429", err: genai.APIError{Code: http.StatusTooManyRequests}, want: true},
		{name: "genai 503", err: genai.APIError{Code: http.StatusServiceUnavailable}, want: true},
		{name: "genai 403", err: genai.APIError{Code: http.StatusForbidden}, want: false},
		{name: "scrape 404", err: &scraper.HTTPStatusError{StatusCode: http.StatusNotFound, URL: "https://example.com"}, want: false},
		{name: "scrape 403", err: &scraper.HTTPStatusError{StatusCode: http.StatusForbidden, URL: "https://example.com"}, want: false},
		{name: "scrape 429", err: &scraper.HTTPStatusError{StatusCode: http.StatusTooManyRequests, URL: "https://example.com"}, want: true},
		{name: "scrape 503", err: &scraper.HTTPStatusError{StatusCode: http.StatusServiceUnavailable, URL: "https://example.com"}, want: true},
		{name: "scrape permanent", err: &scraper.PermanentError{Err: errors.New("unsupported content type")}, want: false},
		{name: "wrapped openai 429", err: errors.Wrap(&openai.Error{StatusCode: 429}, "wrapped"), want: true},
		{name: "wrapped genai 500", err: errors.Wrap(genai.APIError{Code: 500}, "wrapped"), want: true},
		{name: "wrapped scrape 404", err: errors.Wrap(&scraper.HTTPStatusError{StatusCode: 404}, "wrapped"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, retryableError(tt.err))
		})
	}
}

func TestSummarizeWithRetry_NonRetryableAbortsImmediately(t *testing.T) {
	s := &retryTestSummarizer{errs: []error{&openai.Error{StatusCode: http.StatusUnauthorized}}}

	_, err := SummarizeWithRetry(context.Background(), s, "https://example.com", noWait)
	require.Error(t, err)
	assert.Equal(t, 1, s.calls)
	assert.Contains(t, err.Error(), "non-retryable")
}
