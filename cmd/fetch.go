package cmd

import (
	"context"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/alitto/pond/v2"
	"github.com/cockroachdb/errors"
	"github.com/mmcdole/gofeed"
	"github.com/mopemope/quicknews/ent"
	"github.com/mopemope/quicknews/models/article"
	"github.com/mopemope/quicknews/models/feed"
	"github.com/mopemope/quicknews/models/summary"
	"github.com/mopemope/quicknews/pkg/gemini"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// FetchCmd represents the fetch command.
type FetchCmd struct {
	Interval time.Duration `short:"i" help:"Fetch feeds updated within the specified interval (e.g., 24h). Default is 0 (fetch all)."`

	pool         pond.Pool
	client       *ent.Client
	feedRepos    feed.FeedRepository
	articleRepos article.ArticleRepository
	summaryRepos summary.SummaryRepository
}

// Run executes the fetch command.
func (cmd *FetchCmd) Run(client *ent.Client) error {
	ctx := context.Background()
	var newArticlesCount atomic.Int32

	cmd.pool = pond.NewPool(3)
	cmd.client = client
	cmd.feedRepos = feed.NewFeedRepository(client)
	cmd.articleRepos = article.NewArticleRepository(client)
	cmd.summaryRepos = summary.NewSummaryRepository(client)

	for {
		// Check if there are any bookmark feeds
		count, err := cmd.checkFeeds(ctx)
		if err != nil {
			return err
		}
		// Log the number of new articles
		newArticlesCount.Add(count)
		if cmd.Interval > 0 {
			time.Sleep(cmd.Interval)
		} else {
			break
		}
	}

	cmd.pool.StopAndWait()
	return nil
}

func (cmd *FetchCmd) checkFeeds(ctx context.Context) (int32, error) {
	var newArticlesCount atomic.Int32
	feeds, err := cmd.feedRepos.All(ctx)
	if err != nil {
		return 0, err
	}

	if len(feeds) == 0 {
		slog.Info("No feeds registered. Use 'add' command to add feeds.")
		return 0, nil
	}

	slog.Info("Fetching articles", "count", len(feeds))
	feedPool := pond.NewPool(3)

	for _, f := range feeds {
		if f.IsBookmark {
			// Skip bookmark feeds
			continue
		}
		feedPool.Submit(func() {
			count, err := cmd.processFeed(ctx, f)
			if err != nil {
				slog.Error("Error processing feed", "feed", f.URL, "error", err)
				return
			}
			newArticlesCount.Add(int32(count))
		})
	}

	feedPool.StopAndWait()
	return newArticlesCount.Load(), nil
}

// processFeed handles fetching and processing a single feed
func (cmd *FetchCmd) processFeed(
	ctx context.Context,
	feed *ent.Feed,
) (int, error) {

	fp := gofeed.NewParser()
	slog.Info("Fetching feed", "title", feed.Title, "url", feed.URL)

	parsedFeed, err := fp.ParseURLWithContext(feed.URL, ctx)
	if err != nil {
		return 0, errors.Wrap(err, "fetch error")
	}

	updatedFeed, err := cmd.feedRepos.UpdateFeed(ctx, feed, parsedFeed)
	if err != nil {
		return 0, errors.Wrap(err, "error updating feed")
	}
	feed = updatedFeed
	var newArticlesCount atomic.Int32

	for _, item := range parsedFeed.Items {

		cmd.pool.Submit(func() {
			processed, err := cmd.processItem(ctx, feed, item)
			if err != nil {
				slog.Error("Error processing item", "title", item.Title, "link", item.Link, "error", err)
				return
			}
			if processed {
				newArticlesCount.Add(1)
			}
		})

	}

	return int(newArticlesCount.Load()), nil
}

// processItem handles processing a single feed item
func (cmd *FetchCmd) processItem(
	ctx context.Context,
	feed *ent.Feed,
	item *gofeed.Item,
) (bool, error) {

	article, err := cmd.articleRepos.GetFromURL(ctx, item.Link)
	if err != nil {
		slog.Error("Error checking if article exists", "link", item.Link, "error", err)
		return false, err
	}

	if article == nil {
		slog.Info("Processing item", "title", item.Title, "link", item.Link)
		newArticle := &ent.Article{
			Title:       item.Title,
			URL:         item.Link,
			Description: item.Description,
			Content:     item.Content,
		}
		newArticle.Edges.Feed = feed

		// PublishedParsed があれば設定
		if item.PublishedParsed != nil {
			newArticle.PublishedAt = *item.PublishedParsed
		} else if item.UpdatedParsed != nil {
			newArticle.PublishedAt = *item.UpdatedParsed
		}

		article, err = cmd.articleRepos.Save(ctx, newArticle)
		if err != nil {
			slog.Error("Error saving article", "link", item.Link, "error", err)
			return false, err
		}
		article.Edges.Feed = feed
		slog.Debug("Saved article", "link", item.Link, "id", newArticle.ID)
	}

	if article.Edges.Summary == nil {
		if err := cmd.processSummary(ctx, article); err != nil {
			slog.Error("Error processing summary", "link", article.URL, "error", err)
			return false, err
		}
	}

	return true, nil
}

// processSummary generates and saves a summary for the given article
func (cmd *FetchCmd) processSummary(
	ctx context.Context,
	article *ent.Article,
) error {

	geminiClient, err := gemini.NewClient(ctx)
	if err != nil {
		return errors.Wrap(err, "error creating gemini client")
	}

	pageSummary, err := geminiClient.Summarize(ctx, article.URL)
	if err != nil {
		return errors.Wrap(err, "error summarizing article")
	}

	sum := &ent.Summary{
		URL:     article.URL,
		Title:   pageSummary.Title,
		Summary: pageSummary.Summary,
		Readed:  false,
		Listend: false,
	}
	sum.Edges.Article = article
	sum.Edges.Feed = article.Edges.Feed

	slog.Debug("Saving summary", "title", sum.Title, "summary", sum.Summary)
	return cmd.summaryRepos.Save(ctx, sum)
}
