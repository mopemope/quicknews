package gemini

import (
	"context"
	"os"
	"testing"

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

	client, err := NewClient(context.Background())

	require.NoError(t, err, "NewClient should not return an error with a valid API key")
	require.NotNil(t, client, "NewClient should return a non-nil client")
	err = client.Close()
	assert.NoError(t, err, "Close should not return an error")
}

// TestNewClient_NoApiKey tests client creation when the API key is missing.
func TestNewClient_NoApiKey(t *testing.T) {
	// Temporarily unset the API key for this test
	originalApiKey := os.Getenv("GEMINI_API_KEY")
	os.Unsetenv("GEMINI_API_KEY")
	defer os.Setenv("GEMINI_API_KEY", originalApiKey) // Restore original value

	client, err := NewClient(context.Background())

	assert.Error(t, err, "NewClient should return an error when API key is missing")
	assert.Nil(t, client, "NewClient should return a nil client when API key is missing")
	assert.Contains(t, err.Error(), "GEMINI_API_KEY environment variable not set", "Error message should indicate missing API key")
}

// TestSummarizeText tests the Summarize method of the Gemini client.
func TestSummarizeText(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping test: GEMINI_API_KEY environment variable not set")
	}

	client, err := NewClient(context.Background())
	require.NoError(t, err)
	defer client.Close()

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
