package feed

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/google/uuid"
	"github.com/mmcdole/gofeed"
	"github.com/mopemope/quicknews/ent"
	"github.com/mopemope/quicknews/ent/feed"
	"github.com/mopemope/quicknews/pkg/clock"
	"github.com/mopemope/quicknews/pkg/database"
)

type FeedInput struct {
	URL         string
	Title       string
	Description string
	Link        string
}

type FeedRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*ent.Feed, error)
	All(ctx context.Context) ([]*ent.Feed, error)
	UpdateFeed(ctx context.Context, feed *ent.Feed, parsedFeed *gofeed.Feed) (*ent.Feed, error)
	// Exist checks if a feed with the given URL already exists.
	Exist(ctx context.Context, url string) (bool, error)
	Save(ctx context.Context, input *FeedInput) error
}

type FeedRepositoryImpl struct {
	client *ent.Client
}

func NewFeedRepository(client *ent.Client) FeedRepository {
	return &FeedRepositoryImpl{
		client: client,
	}
}

func (r *FeedRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*ent.Feed, error) {
	feed, err := r.client.Feed.
		Query().
		Where(feed.IDEQ(id)).
		Only(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get feed by ID")
	}
	return feed, nil
}

// UpdateFeed updates the feed with the given ID using the parsed feed data.
func (r *FeedRepositoryImpl) UpdateFeed(ctx context.Context, f *ent.Feed, parsedFeed *gofeed.Feed) (*ent.Feed, error) {

	var updatedFeed *ent.Feed
	err := database.WithTx(ctx, r.client, func(tx *ent.Tx) error {

		updateQuery := tx.Feed.UpdateOne(f).
			SetTitle(parsedFeed.Title).
			SetDescription(parsedFeed.Description).
			SetLink(parsedFeed.Link)

		if parsedFeed.UpdatedParsed != nil {
			updateQuery.SetUpdatedAt(*parsedFeed.UpdatedParsed)
		} else if len(parsedFeed.Items) > 0 && parsedFeed.Items[0].PublishedParsed != nil {
			updateQuery.SetUpdatedAt(*parsedFeed.Items[0].PublishedParsed)
		}

		var err error
		updatedFeed, err = updateQuery.Save(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to update feed")
		}
		return nil
	})
	return updatedFeed, err
}

func (r *FeedRepositoryImpl) All(ctx context.Context) ([]*ent.Feed, error) {
	feeds, err := r.client.Feed.
		Query().
		Order(feed.ByOrder()).
		All(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get all feeds")
	}
	return feeds, nil
}

func (r *FeedRepositoryImpl) Exist(ctx context.Context, url string) (bool, error) {
	exists, err := r.client.Feed.
		Query().
		Where(feed.URLEQ(url)).
		Exist(ctx)
	if err != nil {
		return false, errors.Wrap(err, "failed to check if feed exists")
	}
	return exists, nil
}

func (r *FeedRepositoryImpl) Save(ctx context.Context, input *FeedInput) error {
	now := clock.Now()

	return database.WithTx(ctx, r.client, func(tx *ent.Tx) error {
		_, err := tx.Feed.
			Create().
			SetURL(input.URL).
			SetTitle(input.Title).
			SetDescription(input.Description).
			SetLink(input.Link).
			SetUpdatedAt(now). // Set initial updated_at
			Save(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to save feed")
		}
		return nil
	})
}
