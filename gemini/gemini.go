package gemini

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/mopemope/quicknews/config"
	_ "github.com/mopemope/quicknews/log"
	"google.golang.org/genai"
)

var ModelName = "gemini-2.0-flash"

const defaultSummaryPrompt = `
あなたはWebサイトのコンテンツを要約するAIです。
以下のURLのWebサイトにアクセスし、そのページのタイトルと主要な内容を正確に把握し、テキスト形式で出力してください。

URL: %s

出力する際は、以下のルールを厳守してください。
1.  出力: 出力結果をプログラムで整形するのでタイトル、解説のみをシンプルなテキストで出力します。了解しました。などの返事は出力しません。
2.  タイトル: Webサイトのタイトルを正確に日本語に翻訳し、余計な修飾は加えないでください。
3.  解説: 記事の主要な内容を、客観的で分かりやすいニュース記事のようなスタイルで解説してください。**などの強調も不要です。
4.  1行の文字数: 解説の1行あたりの文字数は80文字程度にして下さい。長くなる場合は改行して下さい。1行あたりの文字が長くなりすぎないよう適度に句読点で改行を入れて下さい。
5.  文字数: 解説の文字数は800文字以上を目安とし、内容を十分に伝えられるように記述してください。ただし、情報量が少ない場合は、可能な範囲で詳細に記述してください。
6.  区切り文字: タイトルと解説の間には、必ず「-----」という区切り文字を入れてください。
7.  エラー処理:
    * 指定されたURLが存在しない場合や、アクセスできない場合は、「指定されたURLにアクセスできませんでした。」と出力してください。
    * Webサイトの内容が解説に適さない場合（例：画像や動画が主体である、内容が極めて短いなど）は、「このWebサイトは解説に適していません。」と出力してください。
8.  出力形式は以下です。

<記事のタイトル>
-----
<記事の解説>

`

type PageSummary struct {
	URL     string `json:"url"`
	Title   string `json:"title"`
	Summary string `json:"summary"`
}

// Client wraps the genai.Client.
type Client struct {
	client *genai.Client
	config *config.Config
}

// NewClient creates a new Gemini client.
// It expects the Google API Key to be set in the GEMINI_API_KEY environment variable.
func NewClient(ctx context.Context, config *config.Config) (*Client, error) {
	apiKey := config.GeminiApiKey
	if apiKey == "" {
		return nil, errors.New("GEMINI_API_KEY environment variable not set")
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create genai client")
	}

	return &Client{
		client: client,
		config: config,
	}, nil
}

// Close closes the underlying genai.Client.
func (c *Client) Close() error {
	return nil
}

// Summarize sends a request to the Gemini API to summarize the given text.
func (c *Client) Summarize(ctx context.Context, url string) (*PageSummary, error) {

	summaryPrompt := defaultSummaryPrompt
	if c.config.Prompt != nil && c.config.Prompt.Summary != nil {
		// custom prompt
		summaryPrompt = *c.config.Prompt.Summary
	}
	prompt := fmt.Sprintf(summaryPrompt, url)

	res, err := c.client.Models.GenerateContent(ctx,
		ModelName,
		genai.Text(prompt),
		&genai.GenerateContentConfig{
			Tools: []*genai.Tool{
				{
					GoogleSearch: &genai.GoogleSearch{},
				},
			},
		})
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate content")
	}
	slog.Debug("Sending request to Gemini API", slog.String("model", ModelName), slog.String("url", url))

	// Aggregate text parts from the response
	var summary string
	if len(res.Candidates) > 0 && len(res.Candidates[0].Content.Parts) > 0 {
		for _, part := range res.Candidates[0].Content.Parts {
			summary += part.Text
		}
	} else {
		slog.Warn("Gemini API returned no content or candidates")
		return nil, errors.New("gemini API returned no content")
	}

	summary = strings.TrimSpace(summary)

	// Parse JSON if the response is wrapped in code blocks
	result, err := parseResponse(summary)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse llm response")
	}
	if result == nil {
		return nil, errors.New("parsed result is nil")
	}

	result.URL = url
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
	title = strings.ReplaceAll(title, "了解しました。", "")
	title = strings.ReplaceAll(title, "了解いたしました。", "")
	title = strings.ReplaceAll(title, "*", "")
	title = strings.ReplaceAll(title, "\n", "")

	sum := strings.ReplaceAll(result[1], "\n\n", "\n")
	summaryResponse := PageSummary{
		URL:     "",
		Title:   strings.TrimSpace(title),
		Summary: strings.TrimSpace(sum),
	}
	return &summaryResponse, nil
}
