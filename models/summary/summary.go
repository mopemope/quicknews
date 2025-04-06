package summary

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path"
	"strings"
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
	GetUnlistened(ctx context.Context) ([]*ent.Summary, error)
	UpdateListened(ctx context.Context, sum *ent.Summary) error
	UpdateReaded(ctx context.Context, sum *ent.Summary) error
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

	return database.WithTx(ctx, r.client, func(tx *ent.Tx) error {
		sum, err := tx.Summary.
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
		if err := ExportOrg(sum); err != nil {
			slog.Error("failed to export org", "error", err)
		}
		return nil
	})
}

func (r *SummaryRepositoryImpl) GetUnlistened(ctx context.Context) ([]*ent.Summary, error) {
	sums, err := r.client.Summary.
		Query().
		Where(summary.Listend(false)).
		WithFeed().
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

// GetAudioData generates audio data for the given summary using Google TTS.
func GetAudioData(ctx context.Context, sum *ent.Summary) ([]byte, error) {
	feed := sum.Edges.Feed
	text := fmt.Sprintf(`
これはフィード %s の記事です。
タイトル
%s
要約
%s
`, feed.Title, sum.Title, sum.Summary)

	audioData, err := tts.SynthesizeText(ctx, text)
	if err != nil && err != tts.ErrNoCredentials {
		return nil, errors.Wrap(err, "failed to synthesize text")
	}
	return audioData, nil
}

// ExportOrg exports the summary to an Org file.
func ExportOrg(sum *ent.Summary) error {
	dst := os.Getenv("EXPORT_ORG")
	if dst == "" {
		return nil
	}
	if sum.Edges.Feed == nil || sum.Edges.Article == nil {
		return nil
	}
	feed := sum.Edges.Feed
	article := sum.Edges.Article

	dst = path.Join(dst, convertPathName(feed.Title))
	if err := os.MkdirAll(dst, os.ModePerm); err != nil {
		return errors.Wrap(err, "failed to create directory")
	}

	timestamp := sum.CreatedAt.Format("20060102150405")
	orgFile := timestamp + "-" + convertPathName(sum.Title)
	dst = path.Join(dst, orgFile)

	contentTemplate := `:PROPERTIES:
:ID:       %s
:FEEDURL:  %s
:FEED:     %s
:LINK:     %s
:TITLE:    %s
:END:
#+TITLE:   %s
#+TAGS: feed
#+STARTUP: overview
#+STARTUP: inlineimages
#+OPTIONS: ^:nil

# [[%s][%s]] :feed:

%s
`
	content := fmt.Sprintf(contentTemplate,
		sum.ID,
		feed.URL,
		feed.Title,
		article.URL,
		article.Title,
		sum.Title,
		sum.URL,
		sum.Title,
		sum.Summary)
	return os.WriteFile(dst, []byte(content), os.ModePerm)
}

func convertPathName(name string) string {
	return strings.ReplaceAll(name, " ", "_")
}
