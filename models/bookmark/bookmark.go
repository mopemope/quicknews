package bookmark

import (
	"context"
	"log/slog"
	"time"

	"github.com/mopemope/quicknews/config"
	"github.com/mopemope/quicknews/ent"
	"github.com/mopemope/quicknews/ent/article"
	"github.com/mopemope/quicknews/ent/feed"
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
			return errors.Wrap(err, "failed to check if article exists")
		}

		if existArticle != nil {
			if existArticle.Edges.Feed.IsBookmark {
				// already bookmarked
				slog.Warn("already bookmarked", slog.Any("url", url))
				return nil
			}

			// Update the article and summary to point to the bookmark feed
			if err := tx.Article.
				UpdateOneID(existArticle.ID).
				SetFeedID(bookmarkFeed.ID).
				Exec(ctx); err != nil {
				return errors.Wrap(err, "failed to update article")
			}

			if err := tx.Summary.
				UpdateOneID(existArticle.Edges.Summary.ID).
				SetFeedID(bookmarkFeed.ID).
				Exec(ctx); err != nil {
				return errors.Wrap(err, "failed to update article")
			}
			return nil
		}

		// create new article
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
		article.Edges.Feed = bookmarkFeed

		pageSummary, err := r.summarizePage(ctx, url)
		if err != nil {
			return err
		}
		sum := &ent.Summary{
			URL:     article.URL,
			Title:   pageSummary.Title,
			Summary: pageSummary.Summary,
			Readed:  false,
			Listend: false,
		}
		// Edges can be set for immediate use if needed, but IDs are primary
		sum.Edges.Article = article
		sum.Edges.Feed = bookmarkFeed

		slog.Debug("Saving summary", "title", sum.Title)

		_, err = tx.Summary.
			Create().
			SetTitle(sum.Title).
			SetSummary(sum.Summary).
			SetURL(sum.URL).
			SetCreatedAt(now).
			SetArticle(sum.Edges.Article).
			SetFeed(sum.Edges.Feed).
			Save(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to save summary")
		}

		return nil
	})
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
