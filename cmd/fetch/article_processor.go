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
	feed          *ent.Feed
	feedItem      *gofeed.Item
	articleRepos  article.ArticleRepository
	summaryRepos  summary.SummaryRepository
	config        *config.Config
	newSummarizer func(context.Context, *config.Config) (gemini.Summarizer, error)
	retryWait     func(context.Context, time.Duration) error
}

// NewArticleProcessor creates a new ArticleProcessor
func NewArticleProcessor(feed *ent.Feed, item *gofeed.Item, articleRepos article.ArticleRepository, summaryRepos summary.SummaryRepository, cfg *config.Config) *ArticleProcessor {
	return &ArticleProcessor{
		feed:         feed,
		feedItem:     item,
		articleRepos: articleRepos,
		summaryRepos: summaryRepos,
		config:       cfg,
		newSummarizer: func(ctx context.Context, cfg *config.Config) (gemini.Summarizer, error) {
			return gemini.NewClient(ctx, cfg)
		},
		retryWait: waitWithContext,
	}
}

func NewArticleProcessorWithSummarizer(feed *ent.Feed, item *gofeed.Item, articleRepos article.ArticleRepository, summaryRepos summary.SummaryRepository, cfg *config.Config, newSummarizer func(context.Context, *config.Config) (gemini.Summarizer, error)) *ArticleProcessor {
	processor := NewArticleProcessor(feed, item, articleRepos, summaryRepos, cfg)
	if newSummarizer != nil {
		processor.newSummarizer = newSummarizer
	}
	return processor
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
	geminiClient, err := ap.newSummarizer(ctx, ap.config)
	if err != nil {
		return errors.Wrap(err, "error creating gemini client")
	}
	defer func() {
		if err := geminiClient.Close(); err != nil {
			slog.Warn("failed to close summarizer", "error", err)
		}
	}()

	url := article.URL
	var pageSummary *gemini.PageSummary
	for i := 0; i < 3; i++ {
		pageSummary, err = geminiClient.Summarize(ctx, url)
		if err != nil || pageSummary == nil {
			// retry if error
			slog.Info("retrying to summarize page", "link", url, "error", err)
			if i == 2 {
				break
			}
			wait := (i + 1) * (i + 1)
			if waitErr := ap.retryWait(ctx, time.Duration(wait)*time.Second); waitErr != nil {
				return waitErr
			}
		} else {
			break
		}
	}
	if err != nil {
		return errors.Wrap(err, "error summarizing page")
	}
	if pageSummary == nil {
		return errors.New("summarizer returned nil summary")
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

func waitWithContext(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
