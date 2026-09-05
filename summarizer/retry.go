package summarizer

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/cockroachdb/errors"
	openai "github.com/openai/openai-go/v3"
	"google.golang.org/genai"
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
// with quadratic backoff (1s, 4s). Non-retryable failures (e.g. authentication
// errors) abort immediately. wait overrides the sleep function (useful for
// tests). It returns the last error if all attempts fail.
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

		if attempt == DefaultMaxAttempts-1 {
			break
		}
		if !retryableError(lastErr) {
			slog.Warn("non-retryable summarize failure", "link", url, "error", lastErr)
			return nil, errors.Wrapf(lastErr, "non-retryable summarize failure for %s", url)
		}

		slog.Warn("retrying to summarize page", "link", url, "attempt", attempt+1, "error", lastErr)
		backoff := time.Duration((attempt+1)*(attempt+1)) * time.Second
		if waitErr := wait(ctx, backoff); waitErr != nil {
			return nil, waitErr
		}
	}
	return nil, errors.Wrapf(lastErr, "failed to summarize page after %d attempts", DefaultMaxAttempts)
}

// retryableError reports whether err is worth another attempt.
// Known client errors (4xx other than 408/429) and context cancellation are
// not retryable; unknown failures are treated as transient.
func retryableError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	if status, ok := httpStatusFromError(err); ok {
		return status == http.StatusTooManyRequests ||
			status == http.StatusRequestTimeout ||
			status >= http.StatusInternalServerError
	}
	return true
}

// httpStatusFromError extracts an HTTP status code from provider API errors.
func httpStatusFromError(err error) (int, bool) {
	var oaErr *openai.Error
	if errors.As(err, &oaErr) && oaErr.StatusCode > 0 {
		return oaErr.StatusCode, true
	}
	var gErr genai.APIError
	if errors.As(err, &gErr) && gErr.Code > 0 {
		return gErr.Code, true
	}
	var gPtr *genai.APIError
	if errors.As(err, &gPtr) && gPtr.Code > 0 {
		return gPtr.Code, true
	}
	return 0, false
}
