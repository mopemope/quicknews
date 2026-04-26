package fetch

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/mmcdole/gofeed"
	"github.com/mopemope/quicknews/config"
	"github.com/mopemope/quicknews/ent"
	"github.com/mopemope/quicknews/models/article"
	"github.com/mopemope/quicknews/models/feed"
	"github.com/mopemope/quicknews/models/summary"
	"github.com/stretchr/testify/require"
)

type fakeFeedRepository struct {
	feeds []*ent.Feed
}

func (r *fakeFeedRepository) GetByID(context.Context, uuid.UUID) (*ent.Feed, error) {
	return nil, nil
}

func (r *fakeFeedRepository) GetBookmarkFeed(context.Context) (*ent.Feed, error) {
	return nil, nil
}

func (r *fakeFeedRepository) ExistBookmarkFeed(context.Context) (bool, error) {
	return false, nil
}

func (r *fakeFeedRepository) All(context.Context) ([]*ent.Feed, error) {
	return r.feeds, nil
}

func (r *fakeFeedRepository) UpdateFeed(_ context.Context, feed *ent.Feed, parsedFeed *gofeed.Feed) (*ent.Feed, error) {
	feed.Title = parsedFeed.Title
	return feed, nil
}

func (r *fakeFeedRepository) Exist(context.Context, string) (bool, error) {
	return false, nil
}

func (r *fakeFeedRepository) Save(context.Context, *feed.FeedInput, bool) error {
	return nil
}

func (r *fakeFeedRepository) SaveFeeds(context.Context, []*feed.FeedInput) error {
	return nil
}

func (r *fakeFeedRepository) DeleteWithArticle(context.Context, uuid.UUID) error {
	return nil
}

var _ feed.FeedRepository = (*fakeFeedRepository)(nil)

func TestFeedProcessor_GetItems_ReturnsFeedErrors(t *testing.T) {
	processor := NewFeedProcessor(
		&fakeFeedRepository{feeds: []*ent.Feed{
			{ID: uuid.New(), Title: "Broken Feed", URL: "://invalid-feed-url"},
		}},
		&fakeArticleRepository{},
		&fakeSummaryRepository{},
		&config.Config{},
	)

	items, err := processor.GetItems(context.Background())
	require.Error(t, err)
	require.ErrorContains(t, err, "Broken Feed")
	require.Empty(t, items)
}

var _ article.ArticleRepository = (*fakeArticleRepository)(nil)
var _ summary.SummaryRepository = (*fakeSummaryRepository)(nil)
