package fetch

import (
	"context"
	"log/slog"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/mmcdole/gofeed"
	"github.com/mopemope/quicknews/config"
	"github.com/mopemope/quicknews/ent"
	"github.com/mopemope/quicknews/gemini"
	"github.com/mopemope/quicknews/models/article"
	"github.com/mopemope/quicknews/models/summary"
	"github.com/mopemope/quicknews/org"
)

// ArticleProcessor handles the processing of individual articles
type ArticleProcessor struct {
	feed         *ent.Feed
	feedItem     *gofeed.Item
	articleRepos article.ArticleRepository
	summaryRepos summary.SummaryRepository
	config       *config.Config
}

// NewArticleProcessor creates a new ArticleProcessor
func NewArticleProcessor(feed *ent.Feed, item *gofeed.Item, articleRepos article.ArticleRepository, summaryRepos summary.SummaryRepository, config *config.Config) *ArticleProcessor {
	return &ArticleProcessor{
		feed:         feed,
		feedItem:     item,
		articleRepos: articleRepos,
		summaryRepos: summaryRepos,
		config:       config,
	}
}

// Process handles the processing of an article
func (ap *ArticleProcessor) Process(ctx context.Context) error {
	article, err := ap.articleRepos.GetFromURL(ctx, ap.feedItem.Link)
	if err != nil {
		return errors.Wrap(err, "error checking if article exists")
	}

	if article == nil {
		slog.Info("Processing item", "title", ap.feedItem.Title, "link", ap.feedItem.Link)
		newArticle := &ent.Article{
			Title:       ap.feedItem.Title,
			URL:         ap.feedItem.Link,
			Description: ap.feedItem.Description,
			Content:     ap.feedItem.Content,
		}
		newArticle.Edges.Feed = ap.feed

		// Set PublishedAt if available
		if ap.feedItem.PublishedParsed != nil {
			newArticle.PublishedAt = *ap.feedItem.PublishedParsed
		} else if ap.feedItem.UpdatedParsed != nil {
			newArticle.PublishedAt = *ap.feedItem.UpdatedParsed
		}

		article, err = ap.articleRepos.Save(ctx, newArticle)
		if err != nil {
			return errors.Wrap(err, "error saving article")
		}
		article.Edges.Feed = ap.feed
		slog.Debug("Saved article", "link", ap.feedItem.Link, "id", newArticle.ID)
	}

	if article.Edges.Summary == nil {
		if err := ap.processSummary(ctx, article); err != nil {
			return errors.Wrap(err, "error processing summary")
		}
	}

	return nil
}

// processSummary handles the summarization of an article
func (ap *ArticleProcessor) processSummary(ctx context.Context, article *ent.Article) error {
	geminiClient, err := gemini.NewClient(ctx, ap.config)
	if err != nil {
		return errors.Wrap(err, "error creating gemini client")
	}

	url := article.URL
	var pageSummary *gemini.PageSummary
	for i := 0; i < 3; i++ {
		pageSummary, err = geminiClient.Summarize(ctx, url)
		if err != nil || pageSummary == nil {
			// retry if error
			slog.Info("retrying to summarize page", "link", url, "error", err)
			wait := (i + 1) * (i + 1)
			time.Sleep(time.Duration(wait) * time.Second)
		} else {
			break
		}
	}
	if err != nil {
		return errors.Wrap(err, "error summarizing page")
	}

	sum := &ent.Summary{
		URL:      url,
		Title:    pageSummary.Title,
		Summary:  pageSummary.Summary,
		Readed:   false,
		Listened: false,
	}
	sum.Edges.Article = article
	sum.Edges.Feed = article.Edges.Feed

	slog.Debug("Saving summary", "title", sum.Title, "summary", sum.Summary)
	created, err := ap.summaryRepos.Save(ctx, sum)
	if err != nil {
		slog.Error("Error saving summary", "link", article.URL, "error", err)
		return err
	}

	// Save audio data if configured
	if ap.config.SaveAudioData {
		totalLen := len(created.Summary) + len(created.Title)
		if totalLen > 5000 {
			// skip
			slog.Warn("Skip summary because it is too long",
				slog.Any("total length", totalLen),
				slog.Any("title", pageSummary.Title),
			)
		} else {
			filename, err := summary.SaveAudioData(ctx, created, ap.config)
			if err != nil {
				return err
			}
			if filename != nil {
				if err := ap.summaryRepos.UpdateAudioFile(ctx, created.ID, *filename); err != nil {
					return err
				}
			}
		}
	}

	if err := org.ExportOrg(ap.config, created); err != nil {
		return err
	}

	return nil
}
