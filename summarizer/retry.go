package summarizer

import (
	"context"
	"net/http"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/mopemope/quicknews/scraper"
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
	return summarizeWithRetry(ctx, func() (*PageSummary, error) {
		return s.Summarize(ctx, url)
	}, url, wait)
}

// permanentError wraps failures that must not be retried (e.g. pages with
// no readable content).
type permanentError struct {
	err error
}

func (e *permanentError) Error() string { return e.err.Error() }

func (e *permanentError) Unwrap() error { return e.err }

// NonRetryable marks the error as non-retryable for SummarizeWithRetry.
func (e *permanentError) NonRetryable() bool { return true }

// nonRetryable is implemented by errors that will not succeed on retry.
type nonRetryable interface {
	NonRetryable() bool
}

// retryableError reports whether err is worth another attempt.
// Known client errors (4xx other than 408/429), context cancellation and
// errors marked NonRetryable are not retried; unknown failures are treated
// as transient.
func retryableError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	var nr nonRetryable
	if errors.As(err, &nr) {
		return false
	}

	if status, ok := httpStatusFromError(err); ok {
		return status == http.StatusTooManyRequests ||
			status == http.StatusRequestTimeout ||
			status >= http.StatusInternalServerError
	}
	return true
}

// httpStatusFromError extracts an HTTP status code from provider API errors
// and scraped-page fetch errors.
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
	var scrapeErr *scraper.HTTPStatusError
	if errors.As(err, &scrapeErr) && scrapeErr.StatusCode > 0 {
		return scrapeErr.StatusCode, true
	}
	return 0, false
}
