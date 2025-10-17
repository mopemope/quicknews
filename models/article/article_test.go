package article

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

func TestArticleRepository(t *testing.T) {
	client := enttest.Open(t, dialect.SQLite, "file:ent?mode=memory&cache=shared&_fk=1")
	defer func() { _ = client.Close() }()

	repo := NewRepository(client)

	ctx := context.Background()

	// First create a feed to associate with the article
	feed, err := client.Feed.Create().
		SetURL("https://example.com/feed").
		SetTitle("Test Feed").
		SetDescription("Test Description").
		SetLink("https://example.com").
		SetUpdatedAt(time.Now()).
		SetIsBookmark(false).
		Save(ctx)
	require.NoError(t, err)

	// Test Save
	article := &ent.Article{
		ID:          uuid.New(),
		Title:       "Test Article",
		URL:         "https://example.com/article",
		Description: "Test Description",
		Content:     "Test Content",
		PublishedAt: time.Now(),
	}
	article.Edges.Feed = feed // Set the feed edge

	savedArticle, err := repo.Save(ctx, article)
	require.NoError(t, err)
	assert.Equal(t, "Test Article", savedArticle.Title)

	// Test GetById
	retrievedArticle, err := repo.GetById(ctx, savedArticle.ID)
	require.NoError(t, err)
	assert.Equal(t, "Test Article", retrievedArticle.Title)

	// Test GetFromURL
	retrievedByUrl, err := repo.GetFromURL(ctx, "https://example.com/article")
	require.NoError(t, err)
	assert.Equal(t, "Test Article", retrievedByUrl.Title)

	// Test GetByFeed (empty since no feed is associated yet)
	feedArticles, err := repo.GetByFeed(ctx, uuid.New())
	require.NoError(t, err)
	assert.Empty(t, feedArticles)

	// Test GetByUnreaded (empty since no summary is associated yet)
	unreadArticles, err := repo.GetByUnreaded(ctx, uuid.New())
	require.NoError(t, err)
	assert.Empty(t, unreadArticles)

	// Test GetByDate (empty since no articles match the date yet)
	dateArticles, err := repo.GetByDate(ctx, uuid.New(), time.Now().Format("2006-01-02"))
	require.NoError(t, err)
	assert.Empty(t, dateArticles)

	// Test Delete
	err = repo.Delete(ctx, savedArticle.ID.String())
	require.NoError(t, err)

	// Verify deletion
	deletedArticle, err := repo.GetById(ctx, savedArticle.ID)
	assert.Error(t, err)
	assert.Nil(t, deletedArticle)
}

func TestArticleRepository_GetById_NotFound(t *testing.T) {
	client := enttest.Open(t, dialect.SQLite, "file:ent?mode=memory&cache=shared&_fk=1")
	defer func() { _ = client.Close() }()

	repo := NewRepository(client)
	ctx := context.Background()

	nonExistentID := uuid.New()
	article, err := repo.GetById(ctx, nonExistentID)
	assert.Error(t, err)
	assert.Nil(t, article)
}

func TestArticleRepository_GetFromURL_NotFound(t *testing.T) {
	client := enttest.Open(t, dialect.SQLite, "file:ent?mode=memory&cache=shared&_fk=1")
	defer func() { _ = client.Close() }()

	repo := NewRepository(client)
	ctx := context.Background()

	article, err := repo.GetFromURL(ctx, "https://nonexistent.com")
	assert.NoError(t, err) // Should return nil without error
	assert.Nil(t, article)
}

func TestArticleRepository_SaveAll(t *testing.T) {
	client := enttest.Open(t, dialect.SQLite, "file:ent?mode=memory&cache=shared&_fk=1")
	defer func() { _ = client.Close() }()

	repo := NewRepository(client)
	ctx := context.Background()

	// First create a feed to associate with the articles
	feed, err := client.Feed.Create().
		SetURL("https://example.com/feed").
		SetTitle("Test Feed").
		SetDescription("Test Description").
		SetLink("https://example.com").
		SetUpdatedAt(time.Now()).
		SetIsBookmark(false).
		Save(ctx)
	require.NoError(t, err)

	// Create multiple articles
	articles := ent.Articles{
		{
			ID:          uuid.New(),
			Title:       "Article 1",
			URL:         "https://example.com/article1",
			Description: "Description 1",
			Content:     "Content 1",
			PublishedAt: time.Now(),
		},
		{
			ID:          uuid.New(),
			Title:       "Article 2",
			URL:         "https://example.com/article2",
			Description: "Description 2",
			Content:     "Content 2",
			PublishedAt: time.Now(),
		},
	}

	// Set the feed for each article
	for i := range articles {
		articles[i].Edges.Feed = feed
	}

	err = repo.SaveAll(ctx, articles)
	require.NoError(t, err)

	// Verify articles were saved
	article1, err := repo.GetFromURL(ctx, "https://example.com/article1")
	require.NoError(t, err)
	assert.Equal(t, "Article 1", article1.Title)

	article2, err := repo.GetFromURL(ctx, "https://example.com/article2")
	require.NoError(t, err)
	assert.Equal(t, "Article 2", article2.Title)
}
