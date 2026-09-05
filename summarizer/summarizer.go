package summarizer

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/mopemope/quicknews/config"
)

// New creates a Summarizer for the provider configured in cfg.
// Supported providers are config.SummarizeProviderGemini (default) and
// config.SummarizeProviderOpenAI.
func New(ctx context.Context, cfg *config.Config) (Summarizer, error) {
	provider := config.SummarizeProviderGemini
	if cfg != nil && cfg.SummarizeProvider != "" {
		provider = cfg.SummarizeProvider
	}
	switch provider {
	case config.SummarizeProviderGemini:
		return NewGeminiClient(ctx, cfg)
	case config.SummarizeProviderOpenAI:
		return NewOpenAIClient(ctx, cfg)
	default:
		return nil, errors.Newf("unknown summarize provider: %s (supported: %s, %s)", provider, config.SummarizeProviderGemini, config.SummarizeProviderOpenAI)
	}
}
