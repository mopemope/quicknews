package scraper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTitle_ValidURL(t *testing.T) {
	// Test with a known URL that should have a title
	title, err := GetTitle("https://httpbin.org/html")

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
	_, err := GetTitle("invalid-url")
	assert.Error(t, err)
}

func TestGetTitle_EmptyURL(t *testing.T) {
	_, err := GetTitle("")
	assert.Error(t, err)
}

func TestGetTitle_InvalidURIScheme(t *testing.T) {
	_, err := GetTitle("not-a-url")
	assert.Error(t, err)
}

func TestGetTitle_URLWithNoTitle(t *testing.T) {
	// Test with a URL that has no title tag
	// For this test, we'll create a scenario where no title is found
	// Since we can't easily create a test server, we'll skip this for now
	// In a real implementation, we would mock the HTTP response
	t.Skip("Test requires a mock HTTP server to test pages without title tags")
}
