package bookmark

import (
	"context"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mopemope/quicknews/config"
	"github.com/mopemope/quicknews/ent/article"
	"github.com/mopemope/quicknews/ent/enttest"
	"github.com/mopemope/quicknews/ent/summary"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBookmarkRepository(t *testing.T) {
	client := enttest.Open(t, dialect.SQLite, "file:ent?mode=memory&cache=shared&_fk=1")
	defer func() { _ = client.Close() }()

	// Create a minimal config for testing
	config := &config.Config{
		SaveAudioData: false,
	}

	// First, create a bookmark feed (this is needed for bookmark functionality)
	bookmarkFeed, err := client.Feed.Create().
		SetURL("https://quicknews.org/bookmark/rss").
		SetTitle("Bookmark").
		SetDescription("Bookmark").
		SetLink("https://quicknews.org/bookmark/rss").
		SetUpdatedAt(time.Now()).
		SetIsBookmark(true).
		Save(context.Background())
	require.NoError(t, err)

	// Create the bookmark repository
	repo, err := NewRepository(context.Background(), client, config)
	require.NoError(t, err)

	ctx := context.Background()

	// Get the bookmark feed
	fetchedBookmarkFeed, err := repo.GetBookmarkFeed(ctx)
	require.NoError(t, err)
	assert.Equal(t, bookmarkFeed.ID, fetchedBookmarkFeed.ID)
}

func TestBookmarkRepository_AddBookmark(t *testing.T) {
	client := enttest.Open(t, dialect.SQLite, "file:ent?mode=memory&cache=shared&_fk=1")
	defer func() { _ = client.Close() }()

	// Create a minimal config for testing
	config := &config.Config{
		SaveAudioData: false,
	}

	// First, create a bookmark feed (this is needed for bookmark functionality)
	bookmarkFeed, err := client.Feed.Create().
		SetURL("https://quicknews.org/bookmark/rss").
		SetTitle("Bookmark").
		SetDescription("Bookmark").
		SetLink("https://quicknews.org/bookmark/rss").
		SetUpdatedAt(time.Now()).
		SetIsBookmark(true).
		Save(context.Background())
	require.NoError(t, err)

	feed, err := client.Feed.Create().
		SetURL("https://quicknews.org/feed").
		SetTitle("Source Feed").
		SetDescription("Source Feed").
		SetLink("https://quicknews.org/feed").
		SetUpdatedAt(time.Now()).
		Save(context.Background())
	require.NoError(t, err)

	url := "https://example.com/article"
	articleEntity, err := client.Article.Create().
		SetTitle("Example").
		SetURL(url).
		SetDescription("desc").
		SetContent("content").
		SetCreatedAt(time.Now()).
		SetPublishedAt(time.Now()).
		SetFeed(feed).
		Save(context.Background())
	require.NoError(t, err)

	_, err = client.Summary.Create().
		SetTitle("Example Summary").
		SetSummary("summary").
		SetURL(url).
		SetCreatedAt(time.Now()).
		SetArticle(articleEntity).
		SetFeed(feed).
		Save(context.Background())
	require.NoError(t, err)

	// Create the bookmark repository
	repo, err := NewRepository(context.Background(), client, config)
	require.NoError(t, err)

	ctx := context.Background()

	require.NoError(t, repo.AddBookmark(ctx, url))

	updatedArticle, err := client.Article.Query().
		Where(article.URL(url)).
		WithFeed().
		Only(ctx)
	require.NoError(t, err)
	assert.Equal(t, bookmarkFeed.ID, updatedArticle.Edges.Feed.ID)
	updatedSummary, err := client.Summary.Query().
		Where(summary.URL(url)).
		WithFeed().
		Only(ctx)
	require.NoError(t, err)
	require.NotNil(t, updatedSummary.Edges.Feed)
	assert.True(t, updatedSummary.Edges.Feed.IsBookmark)
}

func TestNewRepository_SucceedsWithoutGeminiKey(t *testing.T) {
	client := enttest.Open(t, dialect.SQLite, "file:ent?mode=memory&cache=shared&_fk=1")
	defer func() { _ = client.Close() }()

	_, err := NewRepository(context.Background(), client, &config.Config{})
	require.NoError(t, err)
}
