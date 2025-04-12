package cmd

import (
	"context"

	"github.com/mopemope/quicknews/config"
	"github.com/mopemope/quicknews/ent"
	"github.com/mopemope/quicknews/models/summary"
)

type ExportAudioCmd struct {
}

func (e *ExportAudioCmd) Run(client *ent.Client, config *config.Config) error {
	summaryRepos := summary.NewRepository(client)
	ctx := context.Background()
	sums, err := summaryRepos.GetAll(ctx)
	if err != nil {
		return err
	}
	for _, sum := range sums {
		// reset audio file name
		f, err := summary.SaveAudioData(ctx, sum, config)
		if err != nil {
			return err
		}
		if f != nil {
			if err := summaryRepos.UpdateAudioFile(ctx, sum.ID, *f); err != nil {
				return err
			}
		}
	}
	return nil
}
