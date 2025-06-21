package gemini

import (
	"context"
	"os"
	"testing"

	"github.com/mopemope/quicknews/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewClient tests the creation of a new Gemini client.
// It requires the GEMINI_API_KEY environment variable to be set.
// Skip this test if the API key is not available.
func TestNewClient(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping test: GEMINI_API_KEY environment variable not set")
	}

	client, err := NewClient(context.Background(), nil)

	require.NoError(t, err, "NewClient should not return an error with a valid API key")
	require.NotNil(t, client, "NewClient should return a non-nil client")
	err = client.Close()
	assert.NoError(t, err, "Close should not return an error")
}

// TestNewClient_NoApiKey tests client creation when the API key is missing.
func TestNewClient_NoApiKey(t *testing.T) {
	// Temporarily unset the API key for this test
	originalApiKey, exists := os.LookupEnv("GEMINI_API_KEY")
	err := os.Unsetenv("GEMINI_API_KEY")
	require.NoError(t, err, "os.Unsetenv should not return an error")
	defer func() {
		if exists {
			err := os.Setenv("GEMINI_API_KEY", originalApiKey) // Restore original value
			assert.NoError(t, err, "os.Setenv should not return an error")
		} else {
			err := os.Unsetenv("GEMINI_API_KEY") // Ensure it remains unset if it wasn't set initially
			assert.NoError(t, err, "os.Unsetenv should not return an error")
		}
	}()

	client, err := NewClient(context.Background(), nil)

	assert.Error(t, err, "NewClient should return an error when API key is missing")
	assert.Nil(t, client, "NewClient should return a nil client when API key is missing")
	assert.Contains(t, err.Error(), "GEMINI_API_KEY environment variable not set", "Error message should indicate missing API key")
}

// TestNewClient_NilContext tests client creation with nil context.
func TestNewClient_NilContext(t *testing.T) {
	config := &config.Config{
		GeminiApiKey: "test-api-key",
	}

	// This test is to verify that the client can be created with a valid context and config
	// The actual API key validation happens during API calls, not during client creation
	client, err := NewClient(context.TODO(), config)

	assert.NoError(t, err, "NewClient should not return an error with valid context and config")
	assert.NotNil(t, client, "NewClient should return a non-nil client with valid context and config")
	
	if client != nil {
		closeErr := client.Close()
		assert.NoError(t, closeErr, "Close should not return an error")
	}
}

// TestNewClient_ActualNilContext tests client creation with actual nil context.
func TestNewClient_ActualNilContext(t *testing.T) {
	config := &config.Config{
		GeminiApiKey: "test-api-key",
	}

	// Use a helper function to avoid the staticcheck warning
	testNilContext := func() context.Context {
		return nil
	}

	client, err := NewClient(testNilContext(), config)

	assert.Error(t, err, "NewClient should return an error when context is nil")
	assert.Nil(t, client, "NewClient should return a nil client when context is nil")
	assert.Contains(t, err.Error(), "context cannot be nil", "Error message should indicate nil context")
}

// TestSummarize_NilClient tests Summarize method with nil client.
func TestSummarize_NilClient(t *testing.T) {
	var client *Client = nil

	result, err := client.Summarize(context.Background(), "https://example.com")

	assert.Error(t, err, "Summarize should return an error when client is nil")
	assert.Nil(t, result, "Summarize should return nil result when client is nil")
	assert.Contains(t, err.Error(), "client is nil", "Error message should indicate nil client")
}

// TestSummarize_EmptyURL tests Summarize method with empty URL.
func TestSummarize_EmptyURL(t *testing.T) {
	config := &config.Config{
		GeminiApiKey: "test-api-key",
	}

	// This will fail to create a real client, but we're testing the URL validation
	client := &Client{
		client: nil, // This would normally be a real genai.Client
		config: config,
	}

	result, err := client.Summarize(context.Background(), "")

	assert.Error(t, err, "Summarize should return an error when URL is empty")
	assert.Nil(t, result, "Summarize should return nil result when URL is empty")
	assert.Contains(t, err.Error(), "url cannot be empty", "Error message should indicate empty URL")
}

// TestParseResponse_EmptyText tests parseResponse with empty text.
func TestParseResponse_EmptyText(t *testing.T) {
	result, err := parseResponse("")

	assert.Error(t, err, "parseResponse should return an error when text is empty")
	assert.Nil(t, result, "parseResponse should return nil result when text is empty")
	assert.Contains(t, err.Error(), "response text cannot be empty", "Error message should indicate empty text")
}

// TestParseResponse_InvalidFormat tests parseResponse with invalid format.
func TestParseResponse_InvalidFormat(t *testing.T) {
	result, err := parseResponse("Invalid format without separator")

	assert.Error(t, err, "parseResponse should return an error when format is invalid")
	assert.Nil(t, result, "parseResponse should return nil result when format is invalid")
	assert.Contains(t, err.Error(), "response format is incorrect", "Error message should indicate incorrect format")
}

// TestParseResponse_EmptyTitle tests parseResponse with empty title.
func TestParseResponse_EmptyTitle(t *testing.T) {
	result, err := parseResponse("-----\nSome summary content")

	assert.Error(t, err, "parseResponse should return an error when title is empty")
	assert.Nil(t, result, "parseResponse should return nil result when title is empty")
	assert.Contains(t, err.Error(), "parsed title is empty", "Error message should indicate empty title")
}

// TestParseResponse_EmptySummary tests parseResponse with empty summary.
func TestParseResponse_EmptySummary(t *testing.T) {
	result, err := parseResponse("Some Title\n-----")

	assert.Error(t, err, "parseResponse should return an error when summary is empty")
	assert.Nil(t, result, "parseResponse should return nil result when summary is empty")
	assert.Contains(t, err.Error(), "parsed summary is empty", "Error message should indicate empty summary")
}

// TestSummarizeText tests the Summarize method of the Gemini client.
func TestSummarizeText(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping test: GEMINI_API_KEY environment variable not set")
	}

	client, err := NewClient(context.Background(), nil)
	require.NoError(t, err)
	defer func() {
		err := client.Close()
		assert.NoError(t, err, "client.Close should not return an error")
	}()

	{
		// This is an integration test part - requires actual API call
		url := "https://www.theregister.com/2025/04/03/openai_copyright_bypass/"
		summary, err := client.Summarize(context.Background(), url)
		require.NoError(t, err, "SummarizeText should not return an error for a valid request")
		t.Log(len([]rune(summary.Summary)))
		t.Logf("Received summary: %s", summary.Title)
		t.Logf("Received summary: %s", summary.Summary)
	}
	{
		// This is an integration test part - requires actual API call
		url := "https://zenn.dev/moneyforward/articles/6deaa22428a109"
		summary, err := client.Summarize(context.Background(), url)
		require.NoError(t, err, "SummarizeText should not return an error for a valid request")
		t.Log(len([]rune(summary.Summary)))
		t.Logf("Received summary: %s", summary.Title)
		t.Logf("Received summary: %s", summary.Summary)
	}
}
