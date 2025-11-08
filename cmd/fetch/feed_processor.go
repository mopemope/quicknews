package fetch

import (
	"context"
	"sync"

	pond "github.com/alitto/pond/v2"
	"github.com/cockroachdb/errors"
	"github.com/mmcdole/gofeed"
	"github.com/mopemope/quicknews/config"
	"github.com/mopemope/quicknews/ent"
	"github.com/mopemope/quicknews/models/article"
	"github.com/mopemope/quicknews/models/feed"
	"github.com/mopemope/quicknews/models/summary"
	"github.com/mopemope/quicknews/tui/progress"
)

// FeedProcessor handles the processing of feeds
type FeedProcessor struct {
	feedRepos    feed.FeedRepository
	articleRepos article.ArticleRepository
	summaryRepos summary.SummaryRepository
	config       *config.Config
}

// NewFeedProcessor creates a new FeedProcessor
func NewFeedProcessor(feedRepos feed.FeedRepository, articleRepos article.ArticleRepository, summaryRepos summary.SummaryRepository, config *config.Config) *FeedProcessor {
	return &FeedProcessor{
		feedRepos:    feedRepos,
		articleRepos: articleRepos,
		summaryRepos: summaryRepos,
		config:       config,
	}
}

// GetItems retrieves all items that need to be processed from all feeds
func (fp *FeedProcessor) GetItems(ctx context.Context) ([]progress.QueueItem, error) {
	items := make([]progress.QueueItem, 0)

	feeds, err := fp.feedRepos.All(ctx)
	if err != nil {
		return nil, err
	}

	if len(feeds) == 0 {
		return nil, nil
	}

	var itemsMutex sync.Mutex

	// Use a worker pool to limit concurrency
	pool := pond.NewPool(5)

	for _, feed := range feeds {
		if feed.IsBookmark {
			// skip bookmark feeds
			continue
		}
		feedData := feed // capture the current feed
		pool.Submit(func() {
			res, err := fp.processFeed(ctx, feedData)
			if err != nil {
				return
			}
			itemsMutex.Lock()
			items = append(items, res...)
			itemsMutex.Unlock()
		})
	}

	pool.StopAndWait()
	return items, nil
}

// processFeed handles fetching and processing a single feed
func (fp *FeedProcessor) processFeed(ctx context.Context, feed *ent.Feed) ([]progress.QueueItem, error) {
	items := make([]progress.QueueItem, 0)

	parser := gofeed.NewParser()

	parsedFeed, err := parser.ParseURLWithContext(feed.URL, ctx)
	if err != nil {
		return nil, errors.Wrap(err, "fetch error")
	}

	updatedFeed, err := fp.feedRepos.UpdateFeed(ctx, feed, parsedFeed)
	if err != nil {
		return nil, errors.Wrap(err, "error updating feed")
	}
	feed = updatedFeed

	for _, item := range parsedFeed.Items {
		articleProcessor := NewArticleProcessor(feed, item, fp.articleRepos, fp.summaryRepos, fp.config)
		items = append(items, &QueueItemWrapper{processor: articleProcessor, name: item.Title})
	}

	return items, nil
}

// QueueItemWrapper wraps the ArticleProcessor to implement the progress.QueueItem interface
type QueueItemWrapper struct {
	processor *ArticleProcessor
	name      string
}

func (q *QueueItemWrapper) DisplayName() string {
	return q.name
}

func (q *QueueItemWrapper) URL() string {
	return q.processor.feedItem.Link
}

func (q *QueueItemWrapper) Process() {
	ctx := context.Background()
	// Process might need to return errors differently, but for now let's just log them
	if err := q.processor.Process(ctx); err != nil {
		// Log the error, as the UI layer might not handle errors from this method directly
	}
}
