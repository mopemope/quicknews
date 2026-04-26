package cmd

import (
	"context"
	stderrors "errors"
	"log/slog"
	"sync"

	pond "github.com/alitto/pond/v2"
	"github.com/mopemope/quicknews/cmd/fetch"
	"github.com/mopemope/quicknews/config"
	"github.com/mopemope/quicknews/ent"
	"github.com/mopemope/quicknews/models/article"
	"github.com/mopemope/quicknews/models/feed"
	"github.com/mopemope/quicknews/models/summary"
)

func fetchArticles(ctx context.Context, client *ent.Client, config *config.Config) {
	feedRepos := feed.NewRepository(client)
	articleRepos := article.NewRepository(client)
	summaryRepos := summary.NewRepository(client)

	feedProcessor := fetch.NewFeedProcessor(feedRepos, articleRepos, summaryRepos, config)
	items, err := feedProcessor.GetItems(ctx)
	if err != nil {
		slog.Error("Error fetching items", "error", err)
	}
	errs := make([]error, 0)
	if err != nil {
		errs = append(errs, err)
	}
	pool := pond.NewPool(3)
	var errsMu sync.Mutex
	for _, item := range items {
		pool.Submit(func() {
			if err := item.Process(); err != nil {
				slog.Error("failed to process item", "title", item.DisplayName(), "url", item.URL(), "error", err)
				errsMu.Lock()
				errs = append(errs, err)
				errsMu.Unlock()
			}
		})
	}
	pool.StopAndWait()
	if err := stderrors.Join(errs...); err != nil {
		slog.Error("background fetch completed with errors", "error", err)
	}
}
