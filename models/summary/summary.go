package summary

import (
	"context"
	"fmt"
	"sync"

	"github.com/cockroachdb/errors"
	"github.com/mopemope/quicknews/ent"
	"github.com/mopemope/quicknews/ent/summary"
	"github.com/mopemope/quicknews/pkg/clock"
	"github.com/mopemope/quicknews/pkg/database"
	"github.com/mopemope/quicknews/pkg/tts"
)

type SummaryRepository interface {
	GetFromURL(ctx context.Context, url string) (*ent.Summary, error)
	Save(ctx context.Context, sum *ent.Summary) error
	UpdateAudioData(ctx context.Context, url string, audioData []byte) error
	GetUnlistened(ctx context.Context) ([]*ent.Summary, error)
	UpdateListened(ctx context.Context, sum *ent.Summary) error
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

func (r *SummaryRepositoryImpl) Save(ctx context.Context, sum *ent.Summary) error {

	now := clock.Now()

	text := fmt.Sprintf("この記事のタイトルは %s です。 \n要約\n\n%s", sum.Title, sum.Summary)

	audioData, err := tts.SynthesizeText(ctx, text)
	if err != nil && err != tts.ErrNoCredentials {
		return errors.Wrap(err, "failed to synthesize text")
	}

	return database.WithTx(ctx, r.client, func(tx *ent.Tx) error {
		_, err := tx.Summary.
			Create().
			SetTitle(sum.Title).
			SetSummary(sum.Summary).
			SetURL(sum.URL).
			SetCreatedAt(now).
			SetAudioData(audioData).
			SetArticle(sum.Edges.Article).
			SetFeed(sum.Edges.Feed).
			Save(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to save summary")
		}
		return nil
	})
}

func (r *SummaryRepositoryImpl) UpdateAudioData(ctx context.Context, url string, audioData []byte) error {
	return database.WithTx(ctx, r.client, func(tx *ent.Tx) error {
		sum, err := tx.Summary.
			Query().
			Where(summary.URL(url)).
			Only(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to get summary for update")
		}

		_, err = tx.Summary.
			UpdateOne(sum).
			SetAudioData(audioData).
			Save(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to update summary with audio data")
		}
		return nil
	})
}

func (r *SummaryRepositoryImpl) GetUnlistened(ctx context.Context) ([]*ent.Summary, error) {
	sums, err := r.client.Summary.
		Query().
		Where(summary.AudioDataNotNil()).
		Where(summary.Listend(false)).
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
