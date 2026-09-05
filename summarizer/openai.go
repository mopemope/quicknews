package summarizer

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
	"github.com/openai/openai-go/v3/shared"

	"github.com/cockroachdb/errors"
	"github.com/mopemope/quicknews/config"
	"github.com/mopemope/quicknews/scraper"
)

// openaiSummaryPrompt is the default instruction for OpenAI based summarizers.
// The page text is appended after the instruction block.
const openaiSummaryPrompt = `
あなたはWebサイトのコンテンツを詳しく解説するアシスタントです。
以下に与えられたWebページの本文を読み、そのページのタイトルと主要な内容を正確に把握し、テキスト形式で出力してください。

ページ本文:
%s

%s
`

// OpenAIClient summarizes pages using the OpenAI Responses API.
// It is compatible with any OpenAI-compatible endpoint via config.OpenAIBaseURL.
type OpenAIClient struct {
	client    openai.Client
	config    *config.Config
	fetchPage func(ctx context.Context, url string) (*scraper.PageContent, error)
}

// NewOpenAIClient creates a new OpenAI summarizer client.
// It expects the OpenAI API key in the OPENAI_API_KEY environment variable if not provided via config.
func NewOpenAIClient(_ context.Context, cfg *config.Config) (*OpenAIClient, error) {
	var apiKey string
	if cfg != nil {
		apiKey = cfg.OpenAIApiKey
	}
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	if apiKey == "" {
		return nil, errors.New("OPENAI_API_KEY environment variable not set")
	}

	opts := []option.RequestOption{option.WithAPIKey(apiKey)}
	if cfg != nil && cfg.OpenAIBaseURL != "" {
		opts = append(opts, option.WithBaseURL(cfg.OpenAIBaseURL))
	}

	if cfg == nil {
		cfg = &config.Config{}
	}

	return &OpenAIClient{
		client:    openai.NewClient(opts...),
		config:    cfg,
		fetchPage: fetchPageContent,
	}, nil
}

// Close implements Summarizer.
func (c *OpenAIClient) Close() error {
	return nil
}

// Summarize fetches the page content and asks the model for a JSON summary.
func (c *OpenAIClient) Summarize(ctx context.Context, url string) (*PageSummary, error) {
	page, err := c.fetchPage(ctx, url)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch page content")
	}

	modelName := c.modelName()
	prompt := openaiPromptFor(c.config, page)
	slog.Debug("Sending request to OpenAI API", slog.String("model", modelName), slog.String("url", url))
	callCtx, cancel := context.WithTimeout(ctx, DefaultSummarizeTimeout)
	defer cancel()

	res, err := c.client.Responses.New(callCtx, responses.ResponseNewParams{
		Model: shared.ResponsesModel(modelName),
		Input: responses.ResponseNewParamsInputUnion{
			OfInputItemList: responses.ResponseInputParam{
				responses.ResponseInputItemParamOfMessage(prompt, responses.EasyInputMessageRoleUser),
			},
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate content")
	}

	summary := strings.TrimSpace(res.OutputText())
	if summary == "" {
		slog.Warn("OpenAI API returned no content")
		return nil, errors.New("openai API returned no content")
	}

	result, err := parseResponse(summary)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse llm response")
	}
	if result == nil {
		return nil, errors.New("parsed result is nil")
	}

	result.URL = url
	slog.Debug("Successfully received summary from OpenAI API")
	return result, nil
}

func (c *OpenAIClient) modelName() string {
	if c.config != nil && c.config.OpenAIModel != "" {
		return c.config.OpenAIModel
	}
	return config.DefaultOpenAIModel
}

// openaiPromptFor builds the prompt containing the fetched page text.
// A custom prompt from config takes precedence; its %s placeholder receives
// the page text.
func openaiPromptFor(cfg *config.Config, page *scraper.PageContent) string {
	pageText := formatPageContent(page)
	if cfg != nil && cfg.Prompt != nil && cfg.Prompt.Summary != nil && strings.TrimSpace(*cfg.Prompt.Summary) != "" {
		return applyPromptTemplate(*cfg.Prompt.Summary, pageText)
	}
	return fmt.Sprintf(openaiSummaryPrompt, pageText, defaultInstructionBlock)
}

// formatPageContent renders the fetched page for the prompt.
func formatPageContent(page *scraper.PageContent) string {
	if page == nil {
		return "(ページの取得に失敗しました)"
	}
	var sb strings.Builder
	sb.WriteString("URL: ")
	sb.WriteString(page.URL)
	if page.Title != "" {
		sb.WriteString("\nタイトル: ")
		sb.WriteString(page.Title)
	}
	if page.Content != "" {
		sb.WriteString("\n\n本文:\n")
		sb.WriteString(page.Content)
	}
	return sb.String()
}

// fetchPageContent retrieves the page body used as model input.
func fetchPageContent(ctx context.Context, url string) (*scraper.PageContent, error) {
	page, err := scraper.GetPageContent(ctx, url)
	if err != nil {
		return nil, err
	}
	if page.Content == "" && page.Title == "" {
		return nil, errors.New("page has no readable content")
	}
	return page, nil
}
