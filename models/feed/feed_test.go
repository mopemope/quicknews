package feed

import (
	"context"
	"testing"

	"entgo.io/ent/dialect"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mopemope/quicknews/ent"
	"github.com/mopemope/quicknews/ent/enttest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFeedRepository(t *testing.T) {
	client := enttest.Open(t, dialect.SQLite, "file:ent?mode=memory&cache=shared&_fk=1")
	defer func() { _ = client.Close() }()

	repo := NewRepository(client)

	ctx := context.Background()

	// Test Save
	input := &FeedInput{
		URL:         "https://example.com/feed",
		Title:       "Test Feed",
		Description: "Test Description",
		Link:        "https://example.com",
	}

	err := repo.Save(ctx, input, false)
	require.NoError(t, err)

	// Test Exist
	exists, err := repo.Exist(ctx, "https://example.com/feed")
	require.NoError(t, err)
	assert.True(t, exists)

	// Test All (should return the feed we just created)
	feeds, err := repo.All(ctx)
	require.NoError(t, err)
	assert.Len(t, feeds, 1)
	assert.Equal(t, "Test Feed", feeds[0].Title)

	// Test GetByID
	feed, err := repo.GetByID(ctx, feeds[0].ID)
	require.NoError(t, err)
	assert.Equal(t, "Test Feed", feed.Title)

	// Test ExistBookmarkFeed (should be false since we didn't create a bookmark feed)
	exists, err = repo.ExistBookmarkFeed(ctx)
	require.NoError(t, err)
	assert.False(t, exists)

	// Test GetBookmarkFeed (should return an error since no bookmark feed exists)
	bookmarkFeed, err := repo.GetBookmarkFeed(ctx)
	assert.Error(t, err)
	assert.Nil(t, bookmarkFeed)

	// Test Save bookmark feed
	err = repo.Save(ctx, &FeedInput{
		URL:         "https://quicknews.org/bookmark/rss",
		Title:       "Bookmark",
		Description: "Bookmark",
		Link:        "https://quicknews.org/bookmark/rss",
	}, true)
	require.NoError(t, err)

	// Test ExistBookmarkFeed (should now be true)
	exists, err = repo.ExistBookmarkFeed(ctx)
	require.NoError(t, err)
	assert.True(t, exists)

	// Test GetBookmarkFeed (should now return the bookmark feed)
	bookmarkFeed, err = repo.GetBookmarkFeed(ctx)
	require.NoError(t, err)
	assert.Equal(t, "Bookmark", bookmarkFeed.Title)

	// Test SaveFeeds (bulk save)
	inputs := []*FeedInput{
		{
			URL:         "https://example.com/feed2",
			Title:       "Test Feed 2",
			Description: "Test Description 2",
			Link:        "https://example.com/2",
		},
		{
			URL:         "https://example.com/feed3",
			Title:       "Test Feed 3",
			Description: "Test Description 3",
			Link:        "https://example.com/3",
		},
	}

	err = repo.SaveFeeds(ctx, inputs)
	require.NoError(t, err)

	// Verify bulk save
	allFeeds, err := repo.All(ctx)
	require.NoError(t, err)
	assert.Len(t, allFeeds, 4) // 2 from bulk save + 1 existing + 1 bookmark
}

func TestFeedRepository_GetByID_NotFound(t *testing.T) {
	client := enttest.Open(t, dialect.SQLite, "file:ent?mode=memory&cache=shared&_fk=1")
	defer func() { _ = client.Close() }()

	repo := NewRepository(client)
	ctx := context.Background()

	nonExistentID := uuid.New()
	feed, err := repo.GetByID(ctx, nonExistentID)
	assert.Error(t, err)
	assert.Nil(t, feed)
}

func TestFeedRepository_DeleteWithArticle(t *testing.T) {
	client := enttest.Open(t, dialect.SQLite, "file:ent?mode=memory&cache=shared&_fk=1")
	defer func() { _ = client.Close() }()

	repo := NewRepository(client)
	ctx := context.Background()

	// Create a feed
	input := &FeedInput{
		URL:         "https://example.com/delete-test",
		Title:       "Delete Test Feed",
		Description: "Test Description",
		Link:        "https://example.com",
	}

	err := repo.Save(ctx, input, false)
	require.NoError(t, err)

	// Get the feed to get its ID
	feeds, err := repo.All(ctx)
	require.NoError(t, err)
	var feedToDelete *ent.Feed
	for _, f := range feeds {
		if f.Title == "Delete Test Feed" {
			feedToDelete = f
			break
		}
	}
	require.NotNil(t, feedToDelete)

	// Delete the feed
	err = repo.DeleteWithArticle(ctx, feedToDelete.ID)
	require.NoError(t, err)

	// Verify the feed was deleted
	remainingFeeds, err := repo.All(ctx)
	require.NoError(t, err)
	assert.Len(t, remainingFeeds, len(feeds)-1)

	// Try to get the deleted feed
	deletedFeed, err := repo.GetByID(ctx, feedToDelete.ID)
	assert.Error(t, err)
	assert.Nil(t, deletedFeed)
}

func TestFeedRepository_DeleteWithArticle_Bookmark(t *testing.T) {
	client := enttest.Open(t, dialect.SQLite, "file:ent?mode=memory&cache=shared&_fk=1")
	defer func() { _ = client.Close() }()

	repo := NewRepository(client)
	ctx := context.Background()

	// Create a bookmark feed
	input := &FeedInput{
		URL:         "https://example.com/bookmark-delete",
		Title:       "Bookmark Delete Test",
		Description: "Test Description",
		Link:        "https://example.com",
	}

	err := repo.Save(ctx, input, true) // bookmark = true
	require.NoError(t, err)

	// Get the bookmark feed to get its ID
	feeds, err := repo.All(ctx)
	require.NoError(t, err)
	var bookmarkFeed *ent.Feed
	for _, f := range feeds {
		if f.Title == "Bookmark Delete Test" {
			bookmarkFeed = f
			break
		}
	}
	require.NotNil(t, bookmarkFeed)

	// Try to delete the bookmark feed - should not delete anything
	err = repo.DeleteWithArticle(ctx, bookmarkFeed.ID)
	require.NoError(t, err)

	// Bookmark feed should still exist (delete is skipped for bookmark feeds)
	remainingFeeds, err := repo.All(ctx)
	require.NoError(t, err)
	assert.Len(t, remainingFeeds, len(feeds)) // same number as before
}
