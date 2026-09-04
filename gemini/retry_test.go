package gemini

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type retryTestSummarizer struct {
	results []*PageSummary
	errs    []error
	calls   int
}

func (s *retryTestSummarizer) Summarize(context.Context, string) (*PageSummary, error) {
	idx := s.calls
	s.calls++
	// Beyond the configured slices, keep returning the last configured values
	// so repeated attempts behave like the first ones.
	if len(s.errs) > 0 && idx >= len(s.errs) {
		idx = len(s.errs) - 1
	}
	if len(s.errs) > 0 && s.errs[idx] != nil {
		return nil, s.errs[idx]
	}
	if len(s.results) == 0 {
		return nil, nil
	}
	if idx >= len(s.results) {
		idx = len(s.results) - 1
	}
	return s.results[idx], nil
}

func (s *retryTestSummarizer) Close() error { return nil }

func noWait(context.Context, time.Duration) error { return nil }

func TestSummarizeWithRetry_SucceedsOnFirstAttempt(t *testing.T) {
	want := &PageSummary{Title: "title", Summary: "summary"}
	s := &retryTestSummarizer{results: []*PageSummary{want}}

	got, err := SummarizeWithRetry(context.Background(), s, "https://example.com", noWait)
	require.NoError(t, err)
	assert.Equal(t, want, got)
	assert.Equal(t, 1, s.calls)
}

func TestSummarizeWithRetry_SucceedsAfterTransientFailure(t *testing.T) {
	want := &PageSummary{Title: "title", Summary: "summary"}
	s := &retryTestSummarizer{
		results: []*PageSummary{nil, want},
		errs:    []error{errors.New("transient"), nil},
	}

	got, err := SummarizeWithRetry(context.Background(), s, "https://example.com", noWait)
	require.NoError(t, err)
	assert.Equal(t, want, got)
	assert.Equal(t, 2, s.calls)
}

func TestSummarizeWithRetry_ExhaustsAttempts(t *testing.T) {
	s := &retryTestSummarizer{errs: []error{errors.New("boom")}}

	_, err := SummarizeWithRetry(context.Background(), s, "https://example.com", noWait)
	require.Error(t, err)
	assert.ErrorContains(t, err, "boom")
	assert.ErrorContains(t, err, "after 3 attempts")
	assert.Equal(t, DefaultMaxAttempts, s.calls)
}

func TestSummarizeWithRetry_NilResultCountsAsFailure(t *testing.T) {
	// All attempts return (nil, nil): must be retried and eventually fail.
	s := &retryTestSummarizer{}

	_, err := SummarizeWithRetry(context.Background(), s, "https://example.com", noWait)
	require.Error(t, err)
	assert.ErrorContains(t, err, "nil summary")
	assert.Equal(t, DefaultMaxAttempts, s.calls)
}

func TestSummarizeWithRetry_ContextCancelStopsRetry(t *testing.T) {
	s := &retryTestSummarizer{errs: []error{errors.New("transient")}}
	ctx, cancel := context.WithCancel(context.Background())

	_, err := SummarizeWithRetry(ctx, s, "https://example.com", func(ctx context.Context, _ time.Duration) error {
		cancel()
		return ctx.Err()
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, 1, s.calls)
}
