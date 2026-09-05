package summarizer

import (
	"context"
	"log/slog"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/mopemope/quicknews/scraper"
)

// MinContentRunes is the minimum page text length required before a summary
// is attempted. Shorter text usually means a blocked or JS-only page whose
// summary would be useless.
const MinContentRunes = 200

// ContentAwareSummarizer is implemented by summarizers that can accept
// already-fetched page content instead of re-fetching the URL themselves.
type ContentAwareSummarizer interface {
	SummarizeContent(ctx context.Context, page *scraper.PageContent) (*PageSummary, error)
}

// SummarizeContentWithRetry summarizes pre-fetched page content, applying the
// same retry policy as SummarizeWithRetry. Summarizers without content
// support fall back to SummarizeWithRetry (they fetch the URL themselves).
func SummarizeContentWithRetry(ctx context.Context, s Summarizer, page *scraper.PageContent, wait RetryWaiter) (*PageSummary, error) {
	if aware, ok := s.(ContentAwareSummarizer); ok {
		return summarizeWithRetry(ctx, func() (*PageSummary, error) {
			return aware.SummarizeContent(ctx, page)
		}, page.URL, wait)
	}
	return SummarizeWithRetry(ctx, s, page.URL, wait)
}

// summarizeWithRetry drives fn with the shared retry policy.
func summarizeWithRetry(ctx context.Context, fn func() (*PageSummary, error), url string, wait RetryWaiter) (*PageSummary, error) {
	if wait == nil {
		wait = DefaultRetryWait
	}

	var lastErr error
	for attempt := range DefaultMaxAttempts {
		pageSummary, err := fn()
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

// validatePageContent rejects pages whose extracted text is too short to
// produce a meaningful summary.
func validatePageContent(page *scraper.PageContent) error {
	if page == nil {
		return &permanentError{err: errors.New("page content is nil")}
	}
	runes := len([]rune(page.Content))
	if runes < MinContentRunes {
		return &permanentError{err: errors.Newf("page content too short: %d runes (minimum %d)", runes, MinContentRunes)}
	}
	return nil
}
