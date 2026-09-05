package summarizer

import (
	"strings"
	"testing"

	"github.com/mopemope/quicknews/config"
	"github.com/mopemope/quicknews/scraper"
	"github.com/stretchr/testify/assert"
)

func pageContentForTest() *scraper.PageContent {
	return &scraper.PageContent{
		URL:     "https://example.com/article",
		Title:   "テスト記事",
		Content: strings.Repeat("本文のテキストです。", 40),
	}
}

func TestPromptFor_Default(t *testing.T) {
	cfg := &config.Config{}
	got := promptFor(cfg, "https://example.com/a")

	assert.Contains(t, got, "URL: https://example.com/a")
	assert.Contains(t, got, "JSON オブジェクトのみ")
	assert.NotContains(t, got, "%s")
}

func TestPromptFor_CustomPromptWithPlaceholder(t *testing.T) {
	custom := "次の記事を要約して: %s"
	cfg := &config.Config{Prompt: &config.Prompt{Summary: &custom}}

	got := promptFor(cfg, "https://example.com/a")
	assert.Equal(t, "次の記事を要約して: https://example.com/a", got)
}

func TestPromptFor_CustomPromptWithLiteralPercent(t *testing.T) {
	custom := "記事 %s を要約。成果率は 100% 以上であること。"
	cfg := &config.Config{Prompt: &config.Prompt{Summary: &custom}}

	got := promptFor(cfg, "https://example.com/a")
	assert.Equal(t, "記事 https://example.com/a を要約。成果率は 100% 以上であること。", got)
}

func TestPromptFor_CustomPromptWithoutPlaceholder(t *testing.T) {
	custom := "記事を要約してください。"
	cfg := &config.Config{Prompt: &config.Prompt{Summary: &custom}}

	got := promptFor(cfg, "https://example.com/a")
	assert.Contains(t, got, "記事を要約してください。")
	assert.Contains(t, got, "https://example.com/a")
}

func TestPromptFor_BlankCustomPromptFallsBack(t *testing.T) {
	blank := "   "
	cfg := &config.Config{Prompt: &config.Prompt{Summary: &blank}}

	got := promptFor(cfg, "https://example.com/a")
	assert.Contains(t, got, "URL: https://example.com/a")
	assert.Contains(t, got, "JSON オブジェクトのみ")
}

func TestApplyPromptTemplate(t *testing.T) {
	assert.Equal(t, "a b", applyPromptTemplate("a %s", "b"))
	assert.Equal(t, "a 100% b", applyPromptTemplate("a 100% %s", "b"))
	assert.Equal(t, "x\n\nvalue", applyPromptTemplate("x", "value"))
}

func TestOpenAIPromptFor(t *testing.T) {
	cfg := &config.Config{}
	page := pageContentForTest()

	got := openaiPromptFor(cfg, page)
	assert.Contains(t, got, "URL: https://example.com/article")
	assert.Contains(t, got, "タイトル: テスト記事")
	assert.Contains(t, got, "本文のテキストです")
	assert.Contains(t, got, "JSON オブジェクトのみ")
}

func TestOpenAIPromptFor_CustomPrompt(t *testing.T) {
	custom := "ページ内容 %s を要約してください。"
	cfg := &config.Config{Prompt: &config.Prompt{Summary: &custom}}

	got := openaiPromptFor(cfg, pageContentForTest())
	assert.Contains(t, got, "ページ内容 URL: https://example.com/article")
}

func TestFormatPageContent(t *testing.T) {
	assert.Equal(t, "(ページの取得に失敗しました)", formatPageContent(nil))

	got := formatPageContent(pageContentForTest())
	assert.Contains(t, got, "URL: https://example.com/article")
	assert.Contains(t, got, "タイトル: テスト記事")
	assert.Contains(t, got, "本文:\n本文のテキストです")
}
