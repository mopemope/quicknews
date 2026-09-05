package summarizer

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	openai "github.com/openai/openai-go/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mopemope/quicknews/config"
	"github.com/mopemope/quicknews/scraper"
)

const testResponseJSON = `{
  "id": "resp_test",
  "object": "response",
  "created_at": 1700000000,
  "status": "completed",
  "model": "gpt-5-mini",
  "output": [
    {
      "id": "msg_1",
      "type": "message",
      "role": "assistant",
      "status": "completed",
      "content": [
        {"type": "output_text", "text": "{\"title\": \"テスト記事\", \"summary\": \"これは要約です\"}", "annotations": []}
      ]
    }
  ],
  "usage": {"input_tokens": 10, "output_tokens": 20, "total_tokens": 30}
}`

// newTestOpenAIClient builds an OpenAIClient backed by a fake page fetcher
// and an httptest server acting as an OpenAI-compatible endpoint.
func newTestOpenAIClient(t *testing.T, handler http.HandlerFunc, page *scraper.PageContent) *OpenAIClient {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	client, err := NewOpenAIClient(context.Background(), &config.Config{
		OpenAIApiKey:  "test-key",
		OpenAIBaseURL: server.URL,
		OpenAIModel:   "gpt-5-mini",
	})
	require.NoError(t, err)

	client.fetchPage = func(context.Context, string) (*scraper.PageContent, error) {
		return page, nil
	}
	return client
}

func TestOpenAIClient_Summarize(t *testing.T) {
	var gotPrompt string
	client := newTestOpenAIClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var payload map[string]any
		_ = json.Unmarshal(body, &payload)
		if items, ok := payload["input"].([]any); ok && len(items) > 0 {
			if msg, ok := items[0].(map[string]any); ok {
				gotPrompt, _ = msg["content"].(string)
			}
		}
		assert.Equal(t, "gpt-5-mini", payload["model"])
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(testResponseJSON))
	}, pageContentForTest())

	got, err := client.Summarize(context.Background(), "https://example.com/article")
	require.NoError(t, err)
	assert.Equal(t, "テスト記事", got.Title)
	assert.Equal(t, "これは要約です", got.Summary)
	assert.Equal(t, "https://example.com/article", got.URL)
	assert.Contains(t, gotPrompt, "https://example.com/article")
}

func TestOpenAIClient_Summarize_PageFetchError(t *testing.T) {
	client := newTestOpenAIClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("API should not be called when page fetch fails")
	}, nil)
	client.fetchPage = func(context.Context, string) (*scraper.PageContent, error) {
		return nil, assert.AnError
	}

	_, err := client.Summarize(context.Background(), "https://example.com/article")
	require.Error(t, err)
	assert.ErrorIs(t, err, assert.AnError)
}

func TestOpenAIClient_Summarize_EmptyOutput(t *testing.T) {
	client := newTestOpenAIClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id": "resp", "object": "response", "status": "completed", "output": []}`))
	}, pageContentForTest())

	_, err := client.Summarize(context.Background(), "https://example.com/article")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no content")
}

func TestOpenAIClient_Summarize_APIErrorNotWrapped(t *testing.T) {
	client := newTestOpenAIClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": {"message": "invalid api key", "type": "invalid_request_error"}}`))
	}, pageContentForTest())

	_, err := client.Summarize(context.Background(), "https://example.com/article")
	require.Error(t, err)

	var oaErr *openai.Error
	require.ErrorAs(t, err, &oaErr)
	assert.Equal(t, http.StatusUnauthorized, oaErr.StatusCode)
}

func TestNewOpenAIClient_RequiresApiKey(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")
	_, err := NewOpenAIClient(context.Background(), &config.Config{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "OPENAI_API_KEY")
}

func TestOpenAIClient_ModelName(t *testing.T) {
	client, err := NewOpenAIClient(context.Background(), &config.Config{OpenAIApiKey: "k"})
	require.NoError(t, err)
	assert.Equal(t, config.DefaultOpenAIModel, client.modelName())

	client2, err := NewOpenAIClient(context.Background(), &config.Config{OpenAIApiKey: "k", OpenAIModel: "gpt-4o-mini"})
	require.NoError(t, err)
	assert.Equal(t, "gpt-4o-mini", client2.modelName())
}

func TestNew_SelectsProvider(t *testing.T) {
	t.Run("openai without key fails", func(t *testing.T) {
		t.Setenv("OPENAI_API_KEY", "")
		_, err := New(context.Background(), &config.Config{SummarizeProvider: config.SummarizeProviderOpenAI})
		require.Error(t, err)
	})
	t.Run("unknown provider fails", func(t *testing.T) {
		_, err := New(context.Background(), &config.Config{SummarizeProvider: "claude"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown summarize provider")
	})
	t.Run("gemini without key fails", func(t *testing.T) {
		t.Setenv("GEMINI_API_KEY", "")
		_, err := New(context.Background(), &config.Config{})
		require.Error(t, err)
	})
	t.Run("openai with key succeeds", func(t *testing.T) {
		client, err := New(context.Background(), &config.Config{SummarizeProvider: config.SummarizeProviderOpenAI, OpenAIApiKey: "k"})
		require.NoError(t, err)
		_, ok := client.(*OpenAIClient)
		assert.True(t, ok)
		require.NoError(t, client.Close())
	})
}
