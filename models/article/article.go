package article

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/google/uuid"
	"github.com/mopemope/quicknews/ent"
	"github.com/mopemope/quicknews/ent/article"
	"github.com/mopemope/quicknews/ent/feed"
	"github.com/mopemope/quicknews/ent/summary"
	"github.com/mopemope/quicknews/pkg/clock"
	"github.com/mopemope/quicknews/pkg/database"
)

type ArticleRepository interface {
	GetById(ctx context.Context, id uuid.UUID) (*ent.Article, error)
	GetByFeed(ctx context.Context, feedID uuid.UUID) (ent.Articles, error)
	GetByUnreaded(ctx context.Context, feedID uuid.UUID) (ent.Articles, error)
	GetFromURL(ctx context.Context, url string) (*ent.Article, error)
	Save(ctx context.Context, article *ent.Article) (*ent.Article, error)
	SaveAll(ctx context.Context, articles ent.Articles) error
}

type ArticleRepositoryImpl struct {
	client *ent.Client
}

// NewArticleRepository creates a new instance of ArticleRepository.
func NewArticleRepository(client *ent.Client) ArticleRepository {
	return &ArticleRepositoryImpl{
		client: client,
	}
}

func (r *ArticleRepositoryImpl) GetById(ctx context.Context, id uuid.UUID) (*ent.Article, error) {
	article, err := r.client.Article.
		Query().
		Where(article.IDEQ(id)).
		WithFeed().
		WithSummary().
		Only(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get article by ID")
	}
	return article, nil
}

func (r *ArticleRepositoryImpl) GetByFeed(ctx context.Context, feedID uuid.UUID) (ent.Articles, error) {
	articles, err := r.client.Article.
		Query().
		Where(article.HasFeedWith(feed.ID(feedID))).
		WithSummary().
		Order(ent.Desc(article.FieldPublishedAt)).
		All(ctx)

	if err != nil {
		return nil, errors.Wrap(err, "failed to get articles by feed ID")
	}
	return articles, nil
}

func (r *ArticleRepositoryImpl) GetByUnreaded(ctx context.Context, feedID uuid.UUID) (ent.Articles, error) {
	articles, err := r.client.Article.
		Query().
		Where(article.HasFeedWith(feed.ID(feedID))).
		WithSummary(func(q *ent.SummaryQuery) {
			q.Where(summary.Readed(false))
		}).
		Order(ent.Desc(article.FieldPublishedAt)).
		All(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get articles by feed ID")
	}
	return articles, nil
}

// GetFromURL retrieves an article from the database by its URL.
func (r *ArticleRepositoryImpl) GetFromURL(ctx context.Context, url string) (*ent.Article, error) {
	article, err := r.client.Article.
		Query().
		Where(article.URL(url)).
		WithSummary().
		Only(ctx)

	if err != nil && !ent.IsNotFound(err) {
		return nil, errors.Wrap(err, "failed to get article")
	}

	return article, nil
}

func (r *ArticleRepositoryImpl) Save(ctx context.Context, article *ent.Article) (*ent.Article, error) {

	now := clock.Now()
	var newArticle *ent.Article
	err := database.WithTx(ctx, r.client, func(tx *ent.Tx) error {
		var err error
		newArticle, err = tx.Article.
			Create().
			SetTitle(article.Title).
			SetURL(article.URL).
			SetDescription(article.Description).
			SetContent(article.Content).
			SetCreatedAt(now).
			SetPublishedAt(article.PublishedAt).
			SetFeed(article.Edges.Feed).
			Save(ctx)

		if err != nil {
			return errors.Wrap(err, "failed to save feed")
		}
		return nil
	})
	return newArticle, err
}

// SaveAll saves all articles to the database.
func (r *ArticleRepositoryImpl) SaveAll(ctx context.Context, articles ent.Articles) error {
	now := clock.Now()

	return database.WithTx(ctx, r.client, func(tx *ent.Tx) error {
		for _, article := range articles {
			_, err := tx.Article.
				Create().
				SetTitle(article.Title).
				SetURL(article.URL).
				SetDescription(article.Description).
				SetContent(article.Content).
				SetCreatedAt(now).
				SetPublishedAt(article.PublishedAt).
				SetFeed(article.Edges.Feed).
				Save(ctx)

			if err != nil {
				return errors.Wrap(err, "failed to save feed")
			}
		}
		return nil
	})

}
