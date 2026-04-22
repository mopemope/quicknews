package gemini

import "context"

// Summarizer is the minimal interface needed by fetch processing.
type Summarizer interface {
	Summarize(ctx context.Context, url string) (*PageSummary, error)
	Close() error
}
