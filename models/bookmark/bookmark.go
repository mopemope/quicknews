package bookmark

import (
	"context"
	"log/slog"
	"time"

	"github.com/mopemope/quicknews/config"
	"github.com/mopemope/quicknews/ent"
	"github.com/mopemope/quicknews/ent/article"
	"github.com/mopemope/quicknews/ent/feed"
	"github.com/mopemope/quicknews/org"
	"github.com/mopemope/quicknews/pkg/clock"
	"github.com/mopemope/quicknews/pkg/database"
	"github.com/mopemope/quicknews/pkg/gemini"
	"github.com/mopemope/quicknews/pkg/scraper"
	"github.com/pkg/errors"
)

type Repository interface {
	AddBookmark(ctx context.Context, url string) error
}

type RepositoryImpl struct {
	client       *ent.Client
	config       *config.Config
	geminiClient *gemini.Client
}

func NewRepository(ctx context.Context, client *ent.Client, config *config.Config) (Repository, error) {
	geminiClient, err := gemini.NewClient(ctx, config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create gemini client")
	}
	return &RepositoryImpl{
		client:       client,
		config:       config,
		geminiClient: geminiClient,
	}, nil
}

func (r *RepositoryImpl) GetBookmarkFeed(ctx context.Context) (*ent.Feed, error) {
	feed, err := r.client.Feed.
		Query().
		Where(feed.IsBookmarkEQ(true)).
		Only(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get bookmark feeds")
	}
	return feed, nil
}

func (r *RepositoryImpl) AddBookmark(ctx context.Context, url string) error {
	bookmarkFeed, err := r.GetBookmarkFeed(ctx)
	if err != nil {
		return err
	}

	return database.WithTx(ctx, r.client, func(tx *ent.Tx) error {

		existArticle, err := tx.Article.Query().
			Where(article.URL(url)).
			WithFeed().
			WithSummary().
			Only(ctx)
		if err != nil && !ent.IsNotFound(err) {
			return errors.Wrap(err, "failed to query existing article")
		}

		if existArticle != nil {
			// Handle existing article: update its feed association
			return r.handleExistingArticle(ctx, tx, existArticle, bookmarkFeed)
		}

		// Create new article and summary
		return r.createNewBookmarkArticle(ctx, tx, url, bookmarkFeed)
	})
}

// handleExistingArticle handles the case where the article already exists.
func (r *RepositoryImpl) handleExistingArticle(ctx context.Context, tx *ent.Tx, existArticle *ent.Article, bookmarkFeed *ent.Feed) error {
	if existArticle.Edges.Feed.IsBookmark {
		// already bookmarked
		slog.Warn("already bookmarked", slog.Any("url", existArticle.URL))
		return nil
	}

	// Update the article and summary to point to the bookmark feed
	if err := tx.Article.
		UpdateOneID(existArticle.ID).
		SetFeedID(bookmarkFeed.ID).
		Exec(ctx); err != nil {
		return errors.Wrap(err, "failed to update article")
	}

	if existArticle.Edges.Summary == nil {
		// Summary might not exist if it failed previously or was deleted
		slog.Warn("summary not found for existing article, skipping summary update", slog.Any("article_id", existArticle.ID))
		return nil
	}

	if err := tx.Summary.
		UpdateOneID(existArticle.Edges.Summary.ID).
		SetFeedID(bookmarkFeed.ID).
		Exec(ctx); err != nil {
		// Log the error but don't fail the whole transaction,
		// as the article itself was successfully moved.
		slog.Error("failed to update summary feed", slog.Any("summary_id", existArticle.Edges.Summary.ID), slog.Any("error", err))
	}
	return nil
}

// createNewBookmarkArticle creates a new article, summary, and exports it.
func (r *RepositoryImpl) createNewBookmarkArticle(ctx context.Context, tx *ent.Tx, url string, bookmarkFeed *ent.Feed) error {
	// get title from url
	title, err := scraper.GetTitle(url)
	if err != nil {
		return errors.Wrap(err, "failed to get title")
	}

	now := clock.Now()
	article, err := tx.Article.Create().
		SetTitle(title).
		SetURL(url).
		SetDescription("").
		SetContent("").
		SetCreatedAt(now).
		SetPublishedAt(now).
		SetFeed(bookmarkFeed).
		Save(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to create article")
	}
	article.Edges.Feed = bookmarkFeed // Set edge for immediate use

	pageSummary, err := r.summarizePage(ctx, url)
	if err != nil {
		// Log the error but proceed to create the summary entry without the AI summary
		slog.Error("failed to summarize page, creating summary entry without AI summary", slog.Any("url", url), slog.Any("error", err))
		pageSummary = &gemini.PageSummary{
			URL:     url,
			Title:   title, // Use scraped title as fallback
			Summary: "",    // Empty summary
		}
	}

	sum := &ent.Summary{
		URL:     article.URL,
		Title:   pageSummary.Title,
		Summary: pageSummary.Summary,
		Readed:  false,
		Listend: false,
		Edges: ent.SummaryEdges{ // Set edges directly
			Article: article,
			Feed:    bookmarkFeed,
		},
	}

	slog.Debug("Saving summary", "title", sum.Title)

	createdSummary, err := tx.Summary.
		Create().
		SetTitle(sum.Title).
		SetSummary(sum.Summary).
		SetURL(sum.URL).
		SetCreatedAt(now).
		SetArticle(sum.Edges.Article). // Use edge directly
		SetFeed(sum.Edges.Feed).       // Use edge directly
		Save(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to save summary")
	}
	sum = createdSummary // Update sum with the created entity including ID

	if err := org.ExportOrg(r.config, sum); err != nil {
		// Log the error but don't fail the transaction
		slog.Error("failed to export org", slog.Any("summary_id", sum.ID), slog.Any("error", err))
	}
	return nil
}

func (r *RepositoryImpl) summarizePage(ctx context.Context, url string) (*gemini.PageSummary, error) {
	var pageSummary *gemini.PageSummary
	var err error
	const maxRetries = 3
	const baseWaitSeconds = 1

	for i := range maxRetries {
		pageSummary, err = r.geminiClient.Summarize(ctx, url)
		if err == nil && pageSummary != nil {
			return pageSummary, nil // Success
		}

		slog.Warn("retrying to summarize page", "link", url, "attempt", i+1, "error", err)
		waitDuration := time.Duration(baseWaitSeconds*(i+1)*(i+1)) * time.Second // Exponential backoff (1, 4, 9 seconds)
		time.Sleep(waitDuration)
	}

	return nil, errors.Wrapf(err, "failed to summarize page after %d attempts", maxRetries)
}
