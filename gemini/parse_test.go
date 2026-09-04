package gemini

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseResponse_JSON(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want *PageSummary
	}{
		{
			name: "plain json",
			in:   `{"title": "記事のタイトル", "summary": "記事の解説"}`,
			want: &PageSummary{Title: "記事のタイトル", Summary: "記事の解説"},
		},
		{
			name: "json with code fence",
			in:   "```json\n{\"title\": \"タイトル\", \"summary\": \"解説\"}\n```",
			want: &PageSummary{Title: "タイトル", Summary: "解説"},
		},
		{
			name: "json with leading chatter",
			in:   "了解しました。\n{\"title\": \"タイトル\", \"summary\": \"解説\"}\n以上です。",
			want: &PageSummary{Title: "タイトル", Summary: "解説"},
		},
		{
			name: "json with braces inside strings",
			in:   `{"title": "Go {言語} の話", "summary": "コードは { } で囲む"}`,
			want: &PageSummary{Title: "Go {言語} の話", Summary: "コードは { } で囲む"},
		},
		{
			name: "json with whitespace",
			in:   "  {\"title\": \"  タイトル  \", \"summary\": \"\\n解説\\n\"}  ",
			want: &PageSummary{Title: "タイトル", Summary: "解説"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseResponse(tt.in)
			require.NoError(t, err)
			assert.Equal(t, tt.want.Title, got.Title)
			assert.Equal(t, tt.want.Summary, got.Summary)
		})
	}
}

func TestParseResponse_LegacyFormat(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want *PageSummary
	}{
		{
			name: "plain legacy",
			in:   "記事のタイトル\n-----\n記事の解説",
			want: &PageSummary{Title: "記事のタイトル", Summary: "記事の解説"},
		},
		{
			name: "legacy with markdown decorations",
			in:   "# **実際のタイトル**\n了解しました。\n-----\n解説1行目\n\n解説2行目",
			want: &PageSummary{Title: "実際のタイトル", Summary: "解説1行目\n解説2行目"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseResponse(tt.in)
			require.NoError(t, err)
			assert.Equal(t, tt.want.Title, got.Title)
			assert.Equal(t, tt.want.Summary, got.Summary)
		})
	}
}

func TestParseResponse_Errors(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{name: "empty", in: ""},
		{name: "no delimiter", in: "タイトルと解説が混在しているテキスト"},
		{name: "multiple delimiters", in: "a\n-----\nb\n-----\nc"},
		{name: "json missing summary", in: `{"title": "タイトルのみ"}`},
		{name: "json missing title", in: `{"summary": "解説のみ"}`},
		{name: "json missing both", in: `{"url": "https://example.com"}`},
		{name: "invalid json object only", in: `{"title": `},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseResponse(tt.in)
			assert.Error(t, err, "input %q should fail to parse", tt.in)
		})
	}
}

func TestExtractJSONObject(t *testing.T) {
	got, ok := extractJSONObject(`前置き {"a": {"b": "}"}, "c": "}"} 後続`)
	require.True(t, ok)
	assert.Equal(t, `{"a": {"b": "}"}, "c": "}"}`, got)

	_, ok = extractJSONObject("JSON object なし")
	assert.False(t, ok)

	_, ok = extractJSONObject(`{"閉じられていない: "value"`)
	assert.False(t, ok)
}

func TestStripCodeFence(t *testing.T) {
	assert.Equal(t, `{"a": 1}`, stripCodeFence("```json\n{\"a\": 1}\n```"))
	assert.Equal(t, `{"a": 1}`, stripCodeFence("```\n{\"a\": 1}\n```"))
	assert.Equal(t, "plain", stripCodeFence("plain"))
}
