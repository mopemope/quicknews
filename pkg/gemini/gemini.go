package gemini

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/google/generative-ai-go/genai"
	_ "github.com/mopemope/quicknews/pkg/log"
	"google.golang.org/api/option"
)

var ModelName = "gemini-2.0-flash"

const summaryPrompt = `
あなたはWebサイトのコンテンツを要約するAIです。
以下のURLのWebサイトにアクセスし、そのページのタイトルと主要な内容を正確に把握し、テキスト形式で出力してください。

URL: %s

出力する際は、以下のルールを厳守してください。

1.  **タイトル:** Webサイトのタイトルを正確に日本語に翻訳し、余計な修飾は加えないでください。
2.  **要約:** 記事の主要な内容を、客観的で分かりやすいニュース記事のようなスタイルで要約してください。
3.  **文字数:** 要約の文字数は800文字以上を目安とし、内容を十分に伝えられるように記述してください。ただし、情報量が少ない場合は、可能な範囲で詳細に記述してください。
4.  **区切り文字:** タイトルと要約の間には、必ず「-----」という区切り文字を入れてください。
5.  **エラー処理:**
    * 指定されたURLが存在しない場合や、アクセスできない場合は、「指定されたURLにアクセスできませんでした。」と出力してください。
    * Webサイトの内容が要約に適さない場合（例：画像や動画が主体である、内容が極めて短いなど）は、「このWebサイトは要約に適していません。」と出力してください。
6.  **出力形式は以下です**

<記事のタイトル>
-----
<記事の要約>

`

type PageSummary struct {
	Title   string `json:"title"`
	Summary string `json:"summary"`
}

// Client wraps the genai.Client.
type Client struct {
	genaiClient *genai.Client
}

// NewClient creates a new Gemini client.
// It expects the Google API Key to be set in the GEMINI_API_KEY environment variable.
func NewClient(ctx context.Context) (*Client, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, errors.New("GEMINI_API_KEY environment variable not set")
	}

	genaiClient, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}

	return &Client{
		genaiClient: genaiClient,
	}, nil
}

// Close closes the underlying genai.Client.
func (c *Client) Close() error {
	return c.genaiClient.Close()
}

// Summarize sends a request to the Gemini API to summarize the given text.
func (c *Client) Summarize(ctx context.Context, url string) (*PageSummary, error) {
	model := c.genaiClient.GenerativeModel(ModelName)

	prompt := fmt.Sprintf(summaryPrompt, url)

	slog.Debug("Sending request to Gemini API", slog.String("model", ModelName)) // Log model name using the variable

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate content")
	}

	// Aggregate text parts from the response
	var summary string
	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		for _, part := range resp.Candidates[0].Content.Parts {
			if textPart, ok := part.(genai.Text); ok {
				summary += string(textPart)
			}
		}
	} else {
		slog.Warn("Gemini API returned no content or candidates")
		return nil, errors.New("gemini API returned no content")
	}

	summary = strings.TrimSpace(summary)

	// Parse JSON if the response is wrapped in code blocks
	result, err := parseResponse(summary)
	if err != nil {
		slog.Error("Failed to parse JSON from response", "error", err, "summary", summary)
		return nil, errors.Wrap(err, "failed to parse JSON from response")
	}
	if result == nil {
		return nil, errors.New("parsed result is nil")
	}

	slog.Debug("Successfully received summary from Gemini API")
	return result, nil
}

// parseResponse parses JSON from text that may be wrapped in code blocks
func parseResponse(text string) (*PageSummary, error) {
	text = strings.TrimSpace(text)
	result := strings.Split(text, "-----")
	if len(result) != 2 {
		return nil, errors.New("response format is incorrect")
	}

	// clean up the title
	title := strings.ReplaceAll(result[0], "#", "")
	title = strings.ReplaceAll(title, "**記事のタイトル**", "")
	title = strings.ReplaceAll(title, "*", "")

	summaryResponse := PageSummary{
		Title:   strings.TrimSpace(title),
		Summary: strings.TrimSpace(result[1]),
	}
	return &summaryResponse, nil
}
