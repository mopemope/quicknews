package rss

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mopemope/quicknews/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRSS(t *testing.T) {
	podcastConfig := &config.Podcast{
		ChannelTitle: "Test Podcast",
		ChannelLink:  "https://example.com",
		ChannelDesc:  "A test podcast",
		Author:       "Test Author",
	}

	rss := NewRSS(podcastConfig)

	assert.Equal(t, "2.0", rss.Version)
	assert.Equal(t, "http://www.itunes.com/dtds/podcast-1.0.dtd", rss.XMLNamespaceItunes)
	assert.Equal(t, "Test Podcast", rss.Channel.Title)
	assert.Equal(t, "https://example.com", rss.Channel.Link)
	assert.Equal(t, "A test podcast", rss.Channel.Description)
	assert.Equal(t, "ja", rss.Channel.Language)
	assert.Equal(t, "no", rss.Channel.ItunesExplicit)
	assert.Empty(t, rss.Channel.Items)

	// Check that pubDate is recent (within last minute)
	parsedTime, err := time.Parse(time.RFC1123Z, rss.Channel.PubDate)
	require.NoError(t, err)

	now := time.Now()
	diff := now.Sub(parsedTime)
	assert.Less(t, diff, time.Minute)
}

func TestRSS_AddItem(t *testing.T) {
	podcastConfig := &config.Podcast{}
	rss := NewRSS(podcastConfig)

	// Add an item
	rssItem := RSSItem{
		Title:       "Test Item",
		Link:        "https://example.com/item",
		Guid:        "test-guid",
		PubDate:     time.Now().Format(time.RFC1123Z),
		Description: "Test description",
		AudioURL:    "https://example.com/audio.mp3",
		Length:      "12345",
		MimeType:    "audio/mpeg",
	}

	rss.AddItem(rssItem)

	assert.Len(t, rss.Channel.Items, 1)
	item := rss.Channel.Items[0]

	assert.Equal(t, "Test Item", item.Title)
	assert.Equal(t, "https://example.com/item", item.Link)
	assert.Equal(t, "test-guid", item.Guid)
	assert.Equal(t, "Test description", item.Description)
	assert.Equal(t, "https://example.com/audio.mp3", item.Enclosure.URL)
	assert.Equal(t, "12345", item.Enclosure.Length)
	assert.Equal(t, "audio/mpeg", item.Enclosure.Type)
}

func TestRSS_AddItem_InheritsChannelProperties(t *testing.T) {
	podcastConfig := &config.Podcast{}
	rss := NewRSS(podcastConfig)

	// Set some channel properties
	rss.Channel.ItunesAuthor = "Channel Author"
	rss.Channel.ItunesSubtitle = "Channel Subtitle"
	rss.Channel.ItunesSummary = "Channel Summary"
	rss.Channel.ItunesImage = ItunesImage{Href: "https://example.com/image.jpg"}
	rss.Channel.ItunesExplicit = "yes"

	rssItem := RSSItem{
		Title:       "Test Item",
		Link:        "https://example.com/item",
		Guid:        "test-guid",
		PubDate:     time.Now().Format(time.RFC1123Z),
		Description: "Test description",
		AudioURL:    "https://example.com/audio.mp3",
		Length:      "12345",
		MimeType:    "audio/mpeg",
	}

	rss.AddItem(rssItem)

	assert.Len(t, rss.Channel.Items, 1)
	item := rss.Channel.Items[0]

	// Check that the item inherits channel properties
	assert.Equal(t, "Channel Author", item.ItunesAuthor)
	assert.Equal(t, "Channel Subtitle", item.ItunesSubtitle)
	assert.Equal(t, "Channel Summary", item.ItunesSummary)
	assert.Equal(t, "https://example.com/image.jpg", item.ItunesImage.Href)
	assert.Equal(t, "yes", item.ItunesExplicit)
}

func TestRSS_WriteToFile(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test_rss.xml")

	podcastConfig := &config.Podcast{
		ChannelTitle: "Test Podcast",
		ChannelLink:  "https://example.com",
		ChannelDesc:  "A test podcast",
	}

	rss := NewRSS(podcastConfig)

	// Add an item
	rssItem := RSSItem{
		Title:       "Test Item",
		Link:        "https://example.com/item",
		Guid:        "test-guid",
		PubDate:     time.Now().Format(time.RFC1123Z),
		Description: "Test description",
		AudioURL:    "https://example.com/audio.mp3",
		Length:      "12345",
		MimeType:    "audio/mpeg",
	}

	rss.AddItem(rssItem)

	err := rss.WriteToFile(filePath)
	require.NoError(t, err)

	// Check if file was created
	_, err = os.Stat(filePath)
	assert.NoError(t, err)

	// Read the file and check content
	content, err := os.ReadFile(filePath)
	assert.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, "<rss version=\"2.0\" xmlns:itunes=\"http://www.itunes.com/dtds/podcast-1.0.dtd\">")
	assert.Contains(t, contentStr, "<title>Test Podcast</title>")
	assert.Contains(t, contentStr, "<title>Test Item</title>")
	assert.Contains(t, contentStr, "Test description")
}

func TestRSS_WriteToFile_InvalidPath(t *testing.T) {
	podcastConfig := &config.Podcast{}
	rss := NewRSS(podcastConfig)

	err := rss.WriteToFile("/invalid/path/file.xml")
	assert.Error(t, err)
}
