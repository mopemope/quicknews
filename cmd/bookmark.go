package cmd

import (
	"context"
	"log/slog"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/mopemope/quicknews/ent"
	"github.com/mopemope/quicknews/models/article"
	"github.com/mopemope/quicknews/models/feed"
	"github.com/mopemope/quicknews/models/summary"
	"github.com/mopemope/quicknews/pkg/clock"
	"github.com/mopemope/quicknews/pkg/gemini"
	"github.com/mopemope/quicknews/pkg/scraper"
)

type BookmarkCmd struct {
	URLs         []string `arg:"" name:"url" help:"URLs of the bookmark to add." required:""`
	feedRepos    feed.FeedRepository
	articleRepos article.ArticleRepository
	summaryRepos summary.SummaryRepository
	geminiClient *gemini.Client
}

func (a *BookmarkCmd) Run(client *ent.Client) error {
	ctx := context.Background()

	if err := a.initializeRepositories(client); err != nil {
		return err
	}

	if err := a.initializeGeminiClient(ctx); err != nil {
		return err
	}

	bookmarkFeed, err := a.feedRepos.GetBookmarkFeed(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get bookmark feed")
	}

	for _, url := range a.URLs {
		if err := a.processURL(ctx, url, bookmarkFeed); err != nil {
			slog.Error("failed to process URL", "url", url, "error", err)
			// Continue processing other URLs even if one fails
		}
	}

	return nil
}

func (a *BookmarkCmd) initializeRepositories(client *ent.Client) error {
	a.feedRepos = feed.NewFeedRepository(client)
	a.articleRepos = article.NewArticleRepository(client)
	a.summaryRepos = summary.NewSummaryRepository(client)
	return nil // Currently no error conditions, but added for future flexibility
}

func (a *BookmarkCmd) initializeGeminiClient(ctx context.Context) error {
	var err error
	a.geminiClient, err = gemini.NewClient(ctx)
	if err != nil {
		return errors.Wrap(err, "error creating gemini client")
	}
	return nil
}

func (a *BookmarkCmd) processURL(ctx context.Context, url string, bookmarkFeed *ent.Feed) error {
	article, needsSummary, err := a.getOrCreateArticle(ctx, url, bookmarkFeed)
	if err != nil {
		return errors.Wrap(err, "failed to get or create article")
	}
	if article == nil { // Article already exists and is a bookmark, or other handled case
		return nil
	}
	if !needsSummary {
		slog.Info("article already exists and has summary", "url", url)
		return nil
	}

	pageSummary, err := a.summarizePage(ctx, url)
	if err != nil {
		return errors.Wrap(err, "failed to summarize page")
	}

	if err := a.saveSummary(ctx, article, bookmarkFeed, pageSummary); err != nil {
		return errors.Wrap(err, "failed to save summary")
	}

	return nil
}

// getOrCreateArticle checks if an article exists. If it exists and is a bookmark, it returns nil.
// If it exists but is not a bookmark, it updates it (TODO).
// If it doesn't exist, it creates a new bookmark article and returns it.
// It also returns a boolean indicating if a summary needs to be created.
func (a *BookmarkCmd) getOrCreateArticle(ctx context.Context, url string, bookmarkFeed *ent.Feed) (*ent.Article, bool, error) {
	existingArticle, err := a.articleRepos.GetFromURL(ctx, url)
	if err != nil && !ent.IsNotFound(err) {
		return nil, false, errors.Wrap(err, "failed to check if article exists")
	}

	if existingArticle != nil {
		if existingArticle.Edges.Feed != nil && existingArticle.Edges.Feed.IsBookmark {
			slog.Info("article already exists as a bookmark", "url", existingArticle.URL)
			// Check if summary exists
			_, summaryErr := a.summaryRepos.GetFromURL(ctx, url)
			if summaryErr == nil {
				return existingArticle, false, nil // Article and summary exist
			}
			if !ent.IsNotFound(summaryErr) {
				return nil, false, errors.Wrap(summaryErr, "failed to check for existing summary")
			}
			// Article exists, but summary doesn't, need to create summary
			return existingArticle, true, nil
		}

		// Article exists but is not a bookmark (e.g., from a regular feed)
		// TODO: Implement logic to update the article to also belong to the bookmark feed if desired.
		// For now, we just log and potentially skip creating a duplicate summary if one exists.
		slog.Info("article exists but is not a bookmark, potentially updating", "url", existingArticle.URL)
		_, summaryErr := a.summaryRepos.GetFromURL(ctx, url)
		if summaryErr == nil {
			return existingArticle, false, nil // Summary exists, no need to recreate
		}
		if !ent.IsNotFound(summaryErr) {
			return nil, false, errors.Wrap(summaryErr, "failed to check for existing summary")
		}
		// Need to create summary for existing non-bookmark article
		return existingArticle, true, nil
	}

	// Article does not exist, create a new one
	title, err := scraper.GetTitle(url)
	if err != nil {
		return nil, false, errors.Wrap(err, "failed to get title")
	}

	newArticle := &ent.Article{
		Title:       title,
		URL:         url,
		Description: "", // Description can be added later if needed
		PublishedAt: clock.Now(),
	}
	newArticle.Edges.Feed = bookmarkFeed
	savedArticle, err := a.articleRepos.Save(ctx, newArticle)
	if err != nil {
		return nil, false, errors.Wrap(err, "failed to save new bookmark article")
	}
	savedArticle.Edges.Feed = bookmarkFeed // Set edge for immediate use

	return savedArticle, true, nil // New article created, needs summary
}

func (a *BookmarkCmd) summarizePage(ctx context.Context, url string) (*gemini.PageSummary, error) {
	var pageSummary *gemini.PageSummary
	var err error
	const maxRetries = 3
	const baseWaitSeconds = 1

	for i := range maxRetries {
		pageSummary, err = a.geminiClient.Summarize(ctx, url)
		if err == nil && pageSummary != nil {
			return pageSummary, nil // Success
		}

		slog.Warn("retrying to summarize page", "link", url, "attempt", i+1, "error", err)
		waitDuration := time.Duration(baseWaitSeconds*(i+1)*(i+1)) * time.Second // Exponential backoff (1, 4, 9 seconds)
		time.Sleep(waitDuration)
	}

	return nil, errors.Wrapf(err, "failed to summarize page after %d attempts", maxRetries)
}

func (a *BookmarkCmd) saveSummary(ctx context.Context, article *ent.Article, feed *ent.Feed, pageSummary *gemini.PageSummary) error {
	sum := &ent.Summary{
		URL:     article.URL, // Use article URL for consistency
		Title:   pageSummary.Title,
		Summary: pageSummary.Summary,
		Readed:  false,
		Listend: false,
	}
	// Edges can be set for immediate use if needed, but IDs are primary
	sum.Edges.Article = article
	sum.Edges.Feed = feed

	slog.Debug("Saving summary", "title", sum.Title)
	if err := a.summaryRepos.Save(ctx, sum); err != nil {
		return errors.Wrap(err, "error saving summary")
	}
	slog.Info("Summary saved successfully", "url", article.URL)
	return nil
}
