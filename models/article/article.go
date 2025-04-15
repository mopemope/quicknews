package article

import (
	"context"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/google/uuid"
	"github.com/mopemope/quicknews/clock"
	"github.com/mopemope/quicknews/database"
	"github.com/mopemope/quicknews/ent"
	"github.com/mopemope/quicknews/ent/article"
	"github.com/mopemope/quicknews/ent/feed"
	"github.com/mopemope/quicknews/ent/summary"
)

type ArticleRepository interface {
	GetById(ctx context.Context, id uuid.UUID) (*ent.Article, error)
	GetByFeed(ctx context.Context, feedID uuid.UUID) (ent.Articles, error)
	GetByUnreaded(ctx context.Context, feedID uuid.UUID) (ent.Articles, error)
	GetFromURL(ctx context.Context, url string) (*ent.Article, error)
	GetByDate(ctx context.Context, feedId uuid.UUID, date string) (ent.Articles, error)
	Save(ctx context.Context, article *ent.Article) (*ent.Article, error)
	SaveAll(ctx context.Context, articles ent.Articles) error
	Delete(ctx context.Context, id string) error
}

type ArticleRepositoryImpl struct {
	client *ent.Client
}

// NewRepository creates a new instance of ArticleRepository.
func NewRepository(client *ent.Client) ArticleRepository {
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
		Order(ent.Desc(article.FieldPublishedAt), ent.Desc(article.FieldCreatedAt)).
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
		Where(article.HasSummaryWith(summary.Readed(false))).
		WithSummary().
		WithFeed().
		Order(ent.Desc(article.FieldPublishedAt), ent.Desc(article.FieldCreatedAt)).
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
		WithFeed().
		WithSummary().
		Only(ctx)

	if err != nil && !ent.IsNotFound(err) {
		return nil, errors.Wrap(err, "failed to get article")
	}

	return article, nil
}

func (r *ArticleRepositoryImpl) GetByDate(ctx context.Context, feedId uuid.UUID, date string) (ent.Articles, error) {
	baseDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse date")
	}

	end := baseDate.UTC()
	start := end.AddDate(0, 0, -1)

	articles, err := r.client.Article.
		Query().
		Where(article.PublishedAtGT(start)).
		Where(article.PublishedAtLTE(end)).
		Where(article.HasFeedWith(feed.ID(feedId))).
		WithFeed().
		WithSummary().
		Order(ent.Asc(article.FieldPublishedAt)).
		All(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get articles by date")
	}
	return articles, nil
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

func (r *ArticleRepositoryImpl) Delete(ctx context.Context, id string) error {
	return database.WithTx(ctx, r.client, func(tx *ent.Tx) error {
		delArticle, err := tx.Article.
			Query().
			Where(article.IDEQ(uuid.MustParse(id))).
			Only(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to get article by ID")
		}

		if _, err := tx.Summary.
			Delete().
			Where(summary.HasArticleWith(article.IDEQ(delArticle.ID))).
			Exec(ctx); err != nil {
			return errors.Wrap(err, "failed to delete summary")
		}

		if _, err := tx.Article.
			Delete().
			Where(article.IDEQ(delArticle.ID)).
			Exec(ctx); err != nil {
			return errors.Wrap(err, "failed to delete article")
		}

		return nil
	})
}
