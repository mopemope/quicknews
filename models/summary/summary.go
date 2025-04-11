package summary

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/cockroachdb/errors"
	"github.com/google/uuid"
	"github.com/mopemope/quicknews/clock"
	"github.com/mopemope/quicknews/config"
	"github.com/mopemope/quicknews/database"
	"github.com/mopemope/quicknews/ent"
	"github.com/mopemope/quicknews/ent/summary"
	"github.com/mopemope/quicknews/tts"
)

type SummaryRepository interface {
	GetFromURL(ctx context.Context, url string) (*ent.Summary, error)
	Save(ctx context.Context, sum *ent.Summary) (*ent.Summary, error)
	GetUnlistened(ctx context.Context) ([]*ent.Summary, error)
	UpdateListened(ctx context.Context, sum *ent.Summary) error
	UpdateReaded(ctx context.Context, sum *ent.Summary) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type SummaryRepositoryImpl struct {
	client *ent.Client
	mutex  *sync.Mutex
}

func NewSummaryRepository(client *ent.Client) SummaryRepository {
	return &SummaryRepositoryImpl{
		client: client,
		mutex:  &sync.Mutex{},
	}
}

func (r *SummaryRepositoryImpl) GetFromURL(ctx context.Context, url string) (*ent.Summary, error) {
	sum, err := r.client.Summary.
		Query().
		Where(summary.URL(url)).
		Only(ctx)

	if err != nil {
		return nil, errors.Wrap(err, "failed to get summary from URL")
	}

	return sum, nil

}

func (r *SummaryRepositoryImpl) Save(ctx context.Context, sum *ent.Summary) (*ent.Summary, error) {

	now := clock.Now()
	var created *ent.Summary

	err := database.WithTx(ctx, r.client, func(tx *ent.Tx) error {
		slog.Debug("Saving summary",
			slog.String("articleTitle", sum.Edges.Article.Title),
			slog.String("articleUrl", sum.Edges.Article.URL),
			slog.String("summaryTitle", sum.Title),
			slog.String("summaryUrl", sum.URL),
		)

		saved, err := tx.Summary.
			Create().
			SetTitle(sum.Title).
			SetSummary(sum.Summary).
			SetURL(sum.URL).
			SetCreatedAt(now).
			SetArticle(sum.Edges.Article).
			SetFeed(sum.Edges.Feed).
			Save(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to save summary")
		}
		saved.Edges.Article = sum.Edges.Article
		saved.Edges.Feed = sum.Edges.Feed
		created = saved
		return nil
	})

	return created, err
}

func (r *SummaryRepositoryImpl) GetUnlistened(ctx context.Context) ([]*ent.Summary, error) {
	sums, err := r.client.Summary.
		Query().
		Where(summary.Listend(false)).
		WithFeed().
		WithArticle().
		Order(ent.Asc(summary.FieldCreatedAt)).
		All(ctx)

	if err != nil {
		return nil, errors.Wrap(err, "failed to get unlistened summaries")
	}
	return sums, nil
}

func (r *SummaryRepositoryImpl) UpdateListened(ctx context.Context, sum *ent.Summary) error {
	return database.WithTx(ctx, r.client, func(tx *ent.Tx) error {
		_, err := tx.Summary.
			UpdateOneID(sum.ID).
			SetListend(true).
			Save(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to update summary as listened")
		}
		return nil
	})
}

func (r *SummaryRepositoryImpl) UpdateReaded(ctx context.Context, sum *ent.Summary) error {
	return database.WithTx(ctx, r.client, func(tx *ent.Tx) error {
		_, err := tx.Summary.
			UpdateOneID(sum.ID).
			SetReaded(true).
			Save(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to update summary as listened")
		}
		return nil
	})
}

func (r *SummaryRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return database.WithTx(ctx, r.client, func(tx *ent.Tx) error {
		if err := tx.Summary.DeleteOneID(id).Exec(ctx); err != nil {
			return errors.Wrap(err, "failed to delete summary")
		}
		return nil
	})
}

// GetAudioData generates audio data for the given summary using the configured TTS engine.
func GetAudioData(ctx context.Context, sum *ent.Summary, cfg *config.Config) ([]byte, error) { // Accept config
	if sum.Edges.Feed == nil {
		return nil, errors.New("summary feed edge is not loaded")
	}
	feed := sum.Edges.Feed
	text := fmt.Sprintf(`
これはフィード %s の記事です。
タイトル
%s
解説
%s
`, feed.Title, sum.Title, sum.Summary)

	ttsEngine := tts.NewTTSEngine(cfg) // Pass config to TTSEngine factory
	audioData, err := ttsEngine.SynthesizeText(ctx, text)
	// Check for specific credentials error if applicable, otherwise wrap generally
	if err != nil {
		if errors.Is(err, tts.ErrNoCredentials) {
			// Return the specific error if it's about credentials
			return nil, err
		}
		return nil, errors.Wrap(err, "failed to synthesize text")
	}
	return audioData, nil
}
