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
