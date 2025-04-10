package cmd

import (
	"context"
	"log/slog"

	"github.com/mopemope/quicknews/config"
	"github.com/mopemope/quicknews/ent"
	"github.com/mopemope/quicknews/models/bookmark"
)

type BookmarkCmd struct {
	URLs []string `arg:"" name:"url" help:"URLs of the bookmark to add." required:""`
}

func (a *BookmarkCmd) Run(client *ent.Client, config *config.Config) error {
	ctx := context.Background()
	bookmarkRepos, err := bookmark.NewRepository(ctx, client, config)
	if err != nil {
		return err
	}
	for _, url := range a.URLs {
		if err := bookmarkRepos.AddBookmark(ctx, url); err != nil {
			slog.Error("failed to add bookmark", slog.Any("url", url), slog.Any("error", err))
		}
	}

	return nil
}
