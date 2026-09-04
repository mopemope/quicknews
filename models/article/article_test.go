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

func TestArticleRepository_GetByDate_UsesSelectedDay(t *testing.T) {
	client := enttest.Open(t, dialect.SQLite, "file:ent?mode=memory&cache=shared&_fk=1")
	defer func() { _ = client.Close() }()

	repo := NewRepository(client)
	ctx := context.Background()

	feed, err := client.Feed.Create().
		SetURL("https://example.com/date-feed").
		SetTitle("Date Feed").
		SetDescription("Technology").
		SetLink("https://example.com").
		SetUpdatedAt(time.Now()).
		SetIsBookmark(false).
		Save(ctx)
	require.NoError(t, err)

	_, err = client.Article.Create().
		SetTitle("Previous day").
		SetURL("https://example.com/previous-day").
		SetDescription("Previous").
		SetContent("Previous").
		SetPublishedAt(time.Date(2026, 4, 21, 23, 59, 59, 0, time.UTC)).
		SetFeed(feed).
		Save(ctx)
	require.NoError(t, err)

	selected, err := client.Article.Create().
		SetTitle("Selected day").
		SetURL("https://example.com/selected-day").
		SetDescription("Selected").
		SetContent("Selected").
		SetPublishedAt(time.Date(2026, 4, 22, 10, 0, 0, 0, time.UTC)).
		SetFeed(feed).
		Save(ctx)
	require.NoError(t, err)

	_, err = client.Article.Create().
		SetTitle("Next day boundary").
		SetURL("https://example.com/next-day").
		SetDescription("Next").
		SetContent("Next").
		SetPublishedAt(time.Date(2026, 4, 23, 0, 0, 0, 0, time.UTC)).
		SetFeed(feed).
		Save(ctx)
	require.NoError(t, err)

	results, err := repo.GetByDate(ctx, feed.ID, "2026-04-22")
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, selected.ID, results[0].ID)
}

func TestArticleRepository_Delete_InvalidID(t *testing.T) {
	client := enttest.Open(t, dialect.SQLite, "file:ent?mode=memory&cache=shared&_fk=1")
	defer func() { _ = client.Close() }()

	repo := NewRepository(client)

	err := repo.Delete(context.Background(), "not-a-uuid")
	require.Error(t, err)
	require.ErrorContains(t, err, "failed to parse article ID")
}

func TestArticleRepository_Search(t *testing.T) {
	client := enttest.Open(t, dialect.SQLite, "file:ent?mode=memory&cache=shared&_fk=1")
	defer func() { _ = client.Close() }()

	repo := NewRepository(client)
	ctx := context.Background()

	feed, err := client.Feed.Create().
		SetURL("https://example.com/feed").
		SetTitle("Tech Feed").
		SetDescription("Technology").
		SetLink("https://example.com").
		SetUpdatedAt(time.Now()).
		SetIsBookmark(false).
		Save(ctx)
	require.NoError(t, err)

	first, err := client.Article.Create().
		SetTitle("Go MCP Integration").
		SetURL("https://example.com/go-mcp").
		SetDescription("Local AI tools").
		SetContent("A practical guide for model context protocol servers.").
		SetPublishedAt(time.Date(2026, 4, 22, 10, 0, 0, 0, time.UTC)).
		SetFeed(feed).
		Save(ctx)
	require.NoError(t, err)

	second, err := client.Article.Create().
		SetTitle("SQLite Notes").
		SetURL("https://example.com/sqlite").
		SetDescription("Database article").
		SetContent("Indexing and query planning.").
		SetPublishedAt(time.Date(2026, 4, 21, 10, 0, 0, 0, time.UTC)).
		SetFeed(feed).
		Save(ctx)
	require.NoError(t, err)

	_, err = client.Summary.Create().
		SetURL(first.URL).
		SetTitle("MCP server summary").
		SetSummary("Connect quicknews articles to other AI tools.").
		SetArticle(first).
		SetFeed(feed).
		Save(ctx)
	require.NoError(t, err)

	_, err = client.Summary.Create().
		SetURL(second.URL).
		SetTitle("Storage summary").
		SetSummary("SQLite is used for local persistence.").
		SetArticle(second).
		SetFeed(feed).
		Save(ctx)
	require.NoError(t, err)

	results, err := repo.Search(ctx, SearchOptions{Query: "mcp ai", Limit: 10})
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, "Go MCP Integration", results[0].Title)
	require.NotNil(t, results[0].Edges.Feed)
	require.NotNil(t, results[0].Edges.Summary)

	results, err = repo.Search(ctx, SearchOptions{Query: "LOCAL persistence", Limit: 10})
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, "SQLite Notes", results[0].Title)

	results, err = repo.Search(ctx, SearchOptions{Query: "article", Limit: 1, Offset: 1})
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, "SQLite Notes", results[0].Title)

	results, err = repo.Search(ctx, SearchOptions{Query: "   ", Limit: 10})
	require.NoError(t, err)
	require.Empty(t, results)
}

func TestArticleRepository_ReadStatusFiltering(t *testing.T) {
	client := enttest.Open(t, dialect.SQLite, "file:ent?mode=memory&cache=shared&_fk=1")
	defer func() { _ = client.Close() }()

	repo := NewRepository(client)
	ctx := context.Background()

	feed, err := client.Feed.Create().
		SetURL("https://example.com/feed").
		SetTitle("Test Feed").
		SetDescription("Test Description").
		SetLink("https://example.com").
		SetUpdatedAt(time.Now()).
		SetIsBookmark(false).
		Save(ctx)
	require.NoError(t, err)

	createArticleWithSummary := func(url, title string, readed bool) *ent.Article {
		a, err := client.Article.Create().
			SetTitle(title).
			SetURL(url).
			SetDescription("Test Description").
			SetContent("Test Content").
			SetCreatedAt(time.Now()).
			SetPublishedAt(time.Now()).
			SetFeed(feed).
			Save(ctx)
		require.NoError(t, err)

		_, err = client.Summary.Create().
			SetURL(url).
			SetTitle(title + " Summary").
			SetSummary("This is a test summary").
			SetReaded(readed).
			SetArticle(a).
			SetFeed(feed).
			Save(ctx)
		require.NoError(t, err)
		return a
	}

	readArticle := createArticleWithSummary("https://example.com/read", "Read Article", true)
	unreadArticle := createArticleWithSummary("https://example.com/unread", "Unread Article", false)

	// GetByFeed returns both read and unread articles
	all, err := repo.GetByFeed(ctx, feed.ID)
	require.NoError(t, err)
	assert.Len(t, all, 2)

	// GetByUnreaded returns only unread articles
	unread, err := repo.GetByUnreaded(ctx, feed.ID)
	require.NoError(t, err)
	assert.Len(t, unread, 1)
	assert.Equal(t, unreadArticle.ID, unread[0].ID)

	// Summary edges are loaded with Readed status
	for _, a := range all {
		require.NotNil(t, a.Edges.Summary)
	}

	fresh, err := repo.GetById(ctx, readArticle.ID)
	require.NoError(t, err)
	require.NotNil(t, fresh.Edges.Summary)
	assert.True(t, fresh.Edges.Summary.Readed)
}
