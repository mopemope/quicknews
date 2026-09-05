package summarizer

import (
	"context"
	"strings"
	"testing"

	"github.com/cockroachdb/errors"
	"github.com/mopemope/quicknews/scraper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidatePageContent(t *testing.T) {
	t.Run("nil page", func(t *testing.T) {
		err := validatePageContent(nil)
		require.Error(t, err)
		assert.True(t, retryableError(err) == false)
	})
	t.Run("too short", func(t *testing.T) {
		err := validatePageContent(&scraper.PageContent{Content: "短すぎる"})
		require.Error(t, err)
		assert.False(t, retryableError(err))
		assert.Contains(t, err.Error(), "too short")
	})
	t.Run("exactly at minimum", func(t *testing.T) {
		content := strings.Repeat("あ", MinContentRunes)
		assert.NoError(t, validatePageContent(&scraper.PageContent{Content: content}))
	})
	t.Run("long enough", func(t *testing.T) {
		content := strings.Repeat("あ", MinContentRunes+1)
		assert.NoError(t, validatePageContent(&scraper.PageContent{Content: content}))
	})
}

func TestSummarizeContentWithRetry(t *testing.T) {
	page := &scraper.PageContent{URL: "https://example.com", Title: "t", Content: strings.Repeat("あ", MinContentRunes)}

	t.Run("retries content-aware failures", func(t *testing.T) {
		s := &flakyContentAware{calls: 0}
		got, err := SummarizeContentWithRetry(context.Background(), s, page, noWait)
		require.NoError(t, err)
		assert.Equal(t, 2, s.calls)
		assert.Equal(t, page.URL, got.URL)
	})
	t.Run("non-aware delegates to SummarizeWithRetry", func(t *testing.T) {
		want := &PageSummary{Title: "t", Summary: "s", URL: page.URL}
		s := &retryTestSummarizer{results: []*PageSummary{want}}
		got, err := SummarizeContentWithRetry(context.Background(), s, page, noWait)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})
	t.Run("permanent failure aborts", func(t *testing.T) {
		s := &flakyContentAware{alwaysPermanent: true}
		_, err := SummarizeContentWithRetry(context.Background(), s, page, noWait)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "non-retryable")
	})
}

// flakyContentAware fails once then succeeds, or always fails permanently.
type flakyContentAware struct {
	calls           int
	alwaysPermanent bool
}

func (f *flakyContentAware) SummarizeContent(_ context.Context, page *scraper.PageContent) (*PageSummary, error) {
	f.calls++
	if f.alwaysPermanent {
		return nil, &permanentError{err: errors.New("permanent")}
	}
	if f.calls < 2 {
		return nil, errors.New("transient")
	}
	return &PageSummary{URL: page.URL}, nil
}
func (f *flakyContentAware) Summarize(context.Context, string) (*PageSummary, error) {
	return nil, errors.New("not implemented")
}
func (f *flakyContentAware) Close() error { return nil }
