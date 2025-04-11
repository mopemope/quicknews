package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/cockroachdb/errors"
	"github.com/mmcdole/gofeed"
	"github.com/mopemope/quicknews/config"
	"github.com/mopemope/quicknews/ent"
	"github.com/mopemope/quicknews/models/article"
	"github.com/mopemope/quicknews/models/feed"
	"github.com/mopemope/quicknews/models/summary"
	"github.com/mopemope/quicknews/org"
	"github.com/mopemope/quicknews/pkg/gemini"
	"github.com/mopemope/quicknews/tui/progress"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// FetchCmd represents the fetch command.
type FetchCmd struct {
	Interval time.Duration `short:"i" help:"Fetch feeds updated within the specified interval (e.g., 24h). Default is 0 (fetch all)."`

	feedRepos    feed.FeedRepository
	articleRepos article.ArticleRepository
	summaryRepos summary.SummaryRepository
	config       *config.Config
}

type Article struct {
	name         string
	feed         *ent.Feed
	feedItem     *gofeed.Item
	articleRepos article.ArticleRepository
	summaryRepos summary.SummaryRepository
	config       *config.Config
}

func (cmd *FetchCmd) NewArticle(feed *ent.Feed, item *gofeed.Item) *Article {
	return &Article{
		name:         item.Title,
		feed:         feed,
		feedItem:     item,
		articleRepos: cmd.articleRepos,
		summaryRepos: cmd.summaryRepos,
		config:       cmd.config,
	}
}

func (a *Article) DisplayName() string {
	return a.name
}

func (a *Article) URL() string {
	return a.feedItem.Link
}

func (a *Article) Process() {
	ctx := context.Background()
	article, err := a.articleRepos.GetFromURL(ctx, a.feedItem.Link)
	if err != nil {
		slog.Error("Error checking if article exists", "link", a.feedItem.Link, "error", err)
		return
	}

	if article == nil {
		slog.Info("Processing item", "title", a.feedItem.Title, "link", a.feedItem.Link)
		newArticle := &ent.Article{
			Title:       a.feedItem.Title,
			URL:         a.feedItem.Link,
			Description: a.feedItem.Description,
			Content:     a.feedItem.Content,
		}
		newArticle.Edges.Feed = a.feed

		// PublishedParsed があれば設定
		if a.feedItem.PublishedParsed != nil {
			newArticle.PublishedAt = *a.feedItem.PublishedParsed
		} else if a.feedItem.UpdatedParsed != nil {
			newArticle.PublishedAt = *a.feedItem.UpdatedParsed
		}

		article, err = a.articleRepos.Save(ctx, newArticle)
		if err != nil {
			slog.Error("Error saving article", "link", a.feedItem.Link, "error", err)
			return
		}
		article.Edges.Feed = a.feed
		slog.Debug("Saved article", "link", a.feedItem.Link, "id", newArticle.ID)
	}

	if article.Edges.Summary == nil {
		if err := a.processSummary(ctx, article); err != nil {
			slog.Error("Error processing summary", "link", article.URL, "error", err)
			return
		}

	}
}

func (a *Article) processSummary(ctx context.Context, article *ent.Article) error {
	geminiClient, err := gemini.NewClient(ctx, a.config)
	if err != nil {
		return errors.Wrap(err, "error creating gemini client")
	}

	url := article.URL
	var pageSummary *gemini.PageSummary
	for i := range 3 {
		pageSummary, err = geminiClient.Summarize(ctx, url)
		if err != nil || pageSummary == nil {
			// retry if error
			slog.Info("retrying to summarize page", "link", url, "error", err)
			i += 1
			wait := i * i
			time.Sleep(time.Duration(wait) * time.Second)
		} else {
			break
		}
	}
	if err != nil {
		return errors.Wrap(err, "error summarizing page")
	}

	sum := &ent.Summary{
		URL:     url,
		Title:   pageSummary.Title,
		Summary: pageSummary.Summary,
		Readed:  false,
		Listend: false,
	}
	sum.Edges.Article = article
	sum.Edges.Feed = article.Edges.Feed

	slog.Debug("Saving summary", "title", sum.Title, "summary", sum.Summary)
	created, err := a.summaryRepos.Save(ctx, sum)
	if err != nil {
		slog.Error("Error saving summary", "link", article.URL, "error", err)
		return err
	}
	if err := org.ExportOrg(a.config, created); err != nil {
		return err
	}
	return nil
}

func (cmd *FetchCmd) getItems(ctx context.Context) ([]progress.QueueItem, error) {

	items := make([]progress.QueueItem, 0)

	feeds, err := cmd.feedRepos.All(ctx)
	if err != nil {
		return nil, err
	}

	if len(feeds) == 0 {
		slog.Info("No feeds registered. Use 'add' command to add feeds.")
		return nil, nil
	}

	for _, feed := range feeds {
		if feed.IsBookmark {
			// skip bookmark feeds
			continue
		}
		res, err := cmd.processFeed(ctx, feed)
		if err != nil {
			return nil, err
		}
		items = append(items, res...)
	}

	return items, nil
}

func (cmd *FetchCmd) Run(client *ent.Client, config *config.Config) error {
	ctx := context.Background()
	cmd.feedRepos = feed.NewFeedRepository(client)
	cmd.articleRepos = article.NewArticleRepository(client)
	cmd.summaryRepos = summary.NewSummaryRepository(client)
	cmd.config = config

	for {
		items, err := cmd.getItems(ctx)
		if err != nil {
			return err
		}

		itemCount := len(items)
		if itemCount > 0 {
			if itemCount > 50 {
				if _, err := tea.NewProgram(progress.NewParallelProgressModel(items, "Fetching", 5)).Run(); err != nil {
					return errors.Wrap(err, "error running progress")
				}
			} else {
				if _, err := tea.NewProgram(progress.NewSingleProgressModel(ctx,
					&progress.Config{
						Client:        client,
						Config:        config,
						Items:         items,
						ProgressLabel: "Fetching",
					})).Run(); err != nil {
					return errors.Wrap(err, "error running progress")
				}
			}

		} else {
			fmt.Println("No new items to process.")
		}

		if cmd.Interval > 0 {
			time.Sleep(cmd.Interval)
		} else {
			break
		}
	}
	return nil
}

// processFeed handles fetching and processing a single feed
func (cmd *FetchCmd) processFeed(
	ctx context.Context,
	feed *ent.Feed,
) ([]progress.QueueItem, error) {

	items := make([]progress.QueueItem, 0)

	fp := gofeed.NewParser()
	slog.Info("Fetching feed", "title", feed.Title, "url", feed.URL)

	parsedFeed, err := fp.ParseURLWithContext(feed.URL, ctx)
	if err != nil {
		return nil, errors.Wrap(err, "fetch error")
	}

	updatedFeed, err := cmd.feedRepos.UpdateFeed(ctx, feed, parsedFeed)
	if err != nil {
		return nil, errors.Wrap(err, "error updating feed")
	}
	feed = updatedFeed

	for _, item := range parsedFeed.Items {
		items = append(items, cmd.NewArticle(feed, item))
	}

	return items, nil
}
