package cmd

import (
	"context"
	"log/slog"

	"github.com/cockroachdb/errors"
	"github.com/mopemope/quicknews/ent"
	"github.com/mopemope/quicknews/models/summary"
	"github.com/mopemope/quicknews/pkg/tts"
)

type PlayCmd struct {
}

func (a *PlayCmd) Run(client *ent.Client) error {

	ctx := context.Background()
	repo := summary.NewSummaryRepository(client)

	res, err := repo.GetUnlistened(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get unlistened summaries")
	}

	slog.Info("unlistened summaries", "count", len(res))
	for _, sum := range res {
		if err := tts.PlayAudioData(sum.AudioData); err != nil {
			slog.Error("failed to play audio data", "error", err)
			continue
		}
		if err := repo.UpdateListened(ctx, sum); err != nil {
			return err
		}
	}
	return nil
}
