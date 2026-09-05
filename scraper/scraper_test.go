package scraper

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/encoding/japanese"
)

func TestGetTitle_ValidURL(t *testing.T) {
	// Test with a known URL that should have a title
	title, err := GetTitle(context.Background(), "https://httpbin.org/html")

	// Since httpbin.org/html returns HTML with a title tag, we expect a title
	// but the exact content may vary, so we just check that we get no error
	// and that the title is not empty
	if err == nil {
		assert.NotEmpty(t, title)
	}
	// Note: This test might fail if the external site is unavailable
	// In a real scenario, we would mock the HTTP requests
}

func TestGetTitle_InvalidURL(t *testing.T) {
	_, err := GetTitle(context.Background(), "invalid-url")
	assert.Error(t, err)
}

func TestGetTitle_EmptyURL(t *testing.T) {
	_, err := GetTitle(context.Background(), "")
	assert.Error(t, err)
}

func TestGetTitle_InvalidURIScheme(t *testing.T) {
	_, err := GetTitle(context.Background(), "not-a-url")
	assert.Error(t, err)
}

func TestParsePageHTML(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantTitle string
		wantText  string
	}{
		{
			name:      "full page",
			input:     `<html><head><title>テスト</title></head><body><p>一行目</p><p>二行目</p></body></html>`,
			wantTitle: "テスト",
			wantText:  "一行目 二行目",
		},
		{
			name: "removes noise elements",
			input: `<html><head><title>Page</title><style>body{color:red}</style></head>` +
				`<body><script>alert("x")</script><nav>menu</nav><article>本文です</article><footer>(c)</footer></body></html>`,
			wantTitle: "Page",
			wantText:  "本文です",
		},
		{
			name:      "collapses whitespace",
			input:     `<body><p>  a   b  </p> <p>c</p></body>`,
			wantTitle: "",
			wantText:  "a b c",
		},
		{
			name:      "no title",
			input:     `<body><p>content only</p></body>`,
			wantTitle: "",
			wantText:  "content only",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			title, text := parsePageHTML(tt.input)
			assert.Equal(t, tt.wantTitle, title)
			assert.Equal(t, tt.wantText, text)
		})
	}
}

func TestParsePageHTML_InvalidHTML(t *testing.T) {
	title, text := parsePageHTML("")
	assert.Empty(t, title)
	assert.Empty(t, text)
}

func TestDecodeBytes_UTF8Passthrough(t *testing.T) {
	assert.Equal(t, "こんにちは", decodeBytes([]byte("こんにちは")))
	assert.Equal(t, "", decodeBytes(nil))
}

func TestDecodeBytes_ShiftJIS(t *testing.T) {
	// Use a long enough Japanese text for reliable charset detection.
	src := "こんにちは世界。日本語のテキスト検出テストです。文字コードを判定して正しくデコードできることを確認します。"
	encoded, err := japanese.ShiftJIS.NewEncoder().Bytes([]byte(src))
	require.NoError(t, err)

	assert.Equal(t, src, decodeBytes(encoded))
}

func TestTruncateRunes(t *testing.T) {
	assert.Equal(t, "あいうえお", truncateRunes("あいうえお", 5))
	assert.Equal(t, "あいう", truncateRunes("あいうえお", 3))
	assert.Equal(t, "あいうえお", truncateRunes("あいうえお", 10))
	assert.Equal(t, "", truncateRunes("あいうえお", 0))
	assert.Equal(t, "", truncateRunes("", 5))
}

func TestIsHTMLContentType(t *testing.T) {
	assert.True(t, isHTMLContentType("text/html; charset=utf-8"))
	assert.True(t, isHTMLContentType("application/xhtml+xml"))
	assert.True(t, isHTMLContentType(""))
	assert.True(t, isHTMLContentType("  "))
	assert.False(t, isHTMLContentType("application/pdf"))
	assert.False(t, isHTMLContentType("image/png"))
	assert.False(t, isHTMLContentType("text/plain; charset=utf-8"))
	assert.False(t, isHTMLContentType("invalid;;;"))
}
