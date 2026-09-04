package gemini

import (
	"context"
	"log/slog"
	"time"

	"github.com/cockroachdb/errors"
)

const (
	// DefaultSummarizeTimeout bounds a single Summarize API call.
	DefaultSummarizeTimeout = 120 * time.Second
	// DefaultMaxAttempts is the number of attempts made by SummarizeWithRetry.
	DefaultMaxAttempts = 3
)

// RetryWaiter waits for the given duration or until the context is cancelled.
type RetryWaiter func(ctx context.Context, duration time.Duration) error

// DefaultRetryWait sleeps for the duration, aborting early if the context is done.
func DefaultRetryWait(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// SummarizeWithRetry summarizes the page at url, retrying transient failures
// with quadratic backoff (1s, 4s). wait overrides the sleep function (useful
// for tests). It returns the last error if all attempts fail.
func SummarizeWithRetry(ctx context.Context, s Summarizer, url string, wait RetryWaiter) (*PageSummary, error) {
	if wait == nil {
		wait = DefaultRetryWait
	}

	var lastErr error
	for attempt := range DefaultMaxAttempts {
		pageSummary, err := s.Summarize(ctx, url)
		if err == nil && pageSummary != nil {
			return pageSummary, nil
		}
		lastErr = err
		if lastErr == nil {
			lastErr = errors.New("summarizer returned nil summary")
		}

		slog.Warn("retrying to summarize page", "link", url, "attempt", attempt+1, "error", lastErr)
		if attempt == DefaultMaxAttempts-1 {
			break
		}

		backoff := time.Duration((attempt+1)*(attempt+1)) * time.Second
		if waitErr := wait(ctx, backoff); waitErr != nil {
			return nil, waitErr
		}
	}
	return nil, errors.Wrapf(lastErr, "failed to summarize page after %d attempts", DefaultMaxAttempts)
}
