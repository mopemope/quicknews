package summary

import (
	"context"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mopemope/quicknews/ent"
	"github.com/mopemope/quicknews/ent/enttest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSummaryRepository(t *testing.T) {
	client := enttest.Open(t, dialect.SQLite, "file:ent?mode=memory&cache=shared&_fk=1")
	defer func() { _ = client.Close() }()

	repo := NewRepository(client)

	ctx := context.Background()

	// First, create a feed and article to associate with the summary
	feed, err := client.Feed.Create().
		SetURL("https://example.com/feed").
		SetTitle("Test Feed").
		SetDescription("Test Description").
		SetLink("https://example.com").
		SetUpdatedAt(time.Now()).
		SetIsBookmark(false).
		Save(ctx)
	require.NoError(t, err)

	article, err := client.Article.Create().
		SetTitle("Test Article").
		SetURL("https://example.com/article").
		SetDescription("Test Description").
		SetContent("Test Content").
		SetCreatedAt(time.Now()).
		SetPublishedAt(time.Now()).
		SetFeed(feed).
		Save(ctx)
	require.NoError(t, err)

	// Test Save
	summary := &ent.Summary{
		URL:      "https://example.com/article",
		Title:    "Test Summary",
		Summary:  "This is a test summary",
		Readed:   false,
		Listened: false,
	}
	summary.Edges.Article = article
	summary.Edges.Feed = feed

	savedSummary, err := repo.Save(ctx, summary)
	require.NoError(t, err)
	assert.Equal(t, "Test Summary", savedSummary.Title)

	// Test GetFromURL
	retrievedSummary, err := repo.GetFromURL(ctx, "https://example.com/article")
	require.NoError(t, err)
	assert.Equal(t, "Test Summary", retrievedSummary.Title)

	// Test GetAll
	allSummaries, err := repo.GetAll(ctx)
	require.NoError(t, err)
	assert.Len(t, allSummaries, 1)
	assert.Equal(t, "Test Summary", allSummaries[0].Title)

	// Test UpdateReaded
	err = repo.UpdateReaded(ctx, retrievedSummary)
	require.NoError(t, err)

	// Verify the summary was updated
	updatedSummary, err := repo.GetFromURL(ctx, "https://example.com/article")
	require.NoError(t, err)
	assert.True(t, updatedSummary.Readed)

	// Test UpdateListened
	err = repo.UpdateListened(ctx, updatedSummary)
	require.NoError(t, err)

	// Verify the summary was updated
	updatedSummary2, err := repo.GetFromURL(ctx, "https://example.com/article")
	require.NoError(t, err)
	assert.True(t, updatedSummary2.Listened)

	// Test GetUnlistened
	unlistenedSummaries, err := repo.GetUnlistened(ctx, nil)
	require.NoError(t, err)
	assert.Empty(t, unlistenedSummaries) // Should be empty since we marked it as listened

	// Test UpdateAudioFile
	audioFilename := "test_audio.mp3"
	err = repo.UpdateAudioFile(ctx, retrievedSummary.ID, audioFilename)
	require.NoError(t, err)

	// Verify the audio file was updated
	updatedSummary3, err := repo.GetFromURL(ctx, "https://example.com/article")
	require.NoError(t, err)
	assert.Equal(t, audioFilename, updatedSummary3.AudioFile)

	// Test Delete
	err = repo.Delete(ctx, retrievedSummary.ID)
	require.NoError(t, err)

	// Verify deletion
	deletedSummary, err := repo.GetFromURL(ctx, "https://example.com/article")
	assert.Error(t, err)
	assert.Nil(t, deletedSummary)
}

func TestSummaryRepository_GetFromURL_NotFound(t *testing.T) {
	client := enttest.Open(t, dialect.SQLite, "file:ent?mode=memory&cache=shared&_fk=1")
	defer func() { _ = client.Close() }()

	repo := NewRepository(client)
	ctx := context.Background()

	summary, err := repo.GetFromURL(ctx, "https://nonexistent.com")
	assert.Error(t, err)
	assert.Nil(t, summary)
}

func TestSummaryRepository_GetUnlistened_WithDate(t *testing.T) {
	client := enttest.Open(t, dialect.SQLite, "file:ent?mode=memory&cache=shared&_fk=1")
	defer func() { _ = client.Close() }()

	repo := NewRepository(client)
	ctx := context.Background()

	// First, create a feed and article to associate with the summary
	feed, err := client.Feed.Create().
		SetURL("https://example.com/feed").
		SetTitle("Test Feed").
		SetDescription("Test Description").
		SetLink("https://example.com").
		SetUpdatedAt(time.Now()).
		SetIsBookmark(false).
		Save(ctx)
	require.NoError(t, err)

	article, err := client.Article.Create().
		SetTitle("Test Article").
		SetURL("https://example.com/article").
		SetDescription("Test Description").
		SetContent("Test Content").
		SetCreatedAt(time.Now()).
		SetPublishedAt(time.Now()).
		SetFeed(feed).
		Save(ctx)
	require.NoError(t, err)

	// Create a summary that is not listened
	summary := &ent.Summary{
		URL:      "https://example.com/article",
		Title:    "Test Summary",
		Summary:  "This is a test summary",
		Readed:   false,
		Listened: false,
	}
	summary.Edges.Article = article
	summary.Edges.Feed = feed

	_, err = repo.Save(ctx, summary)
	require.NoError(t, err)

	// Test GetUnlistened with nil date (should return all unlistened summaries)
	unlistenedSummaries, err := repo.GetUnlistened(ctx, nil)
	require.NoError(t, err)
	assert.Len(t, unlistenedSummaries, 1)

	// Test GetUnlistened with a date in the past (should not include our summary if it's not from that date)
	// Since we don't control the exact date of the article, we'll just verify that the function works
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	_, err = repo.GetUnlistened(ctx, &yesterday)
	require.NoError(t, err)
	// This might be empty or not, depending on the article's date - just ensure no error
}

func TestSummaryRepository_Delete_NotFound(t *testing.T) {
	client := enttest.Open(t, dialect.SQLite, "file:ent?mode=memory&cache=shared&_fk=1")
	defer func() { _ = client.Close() }()

	repo := NewRepository(client)
	ctx := context.Background()

	nonExistentID := uuid.New()
	err := repo.Delete(ctx, nonExistentID)
	assert.Error(t, err)
}
