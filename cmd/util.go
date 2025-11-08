package cmd

import (
	"context"
	"log/slog"

	pond "github.com/alitto/pond/v2"
	"github.com/mopemope/quicknews/cmd/fetch"
	"github.com/mopemope/quicknews/config"
	"github.com/mopemope/quicknews/ent"
	"github.com/mopemope/quicknews/models/article"
	"github.com/mopemope/quicknews/models/feed"
	"github.com/mopemope/quicknews/models/summary"
)

func fetchArticles(client *ent.Client, config *config.Config) {
	feedRepos := feed.NewRepository(client)
	articleRepos := article.NewRepository(client)
	summaryRepos := summary.NewRepository(client)

	feedProcessor := fetch.NewFeedProcessor(feedRepos, articleRepos, summaryRepos, config)
	ctx := context.Background()
	items, err := feedProcessor.GetItems(ctx)
	if err != nil {
		slog.Error("Error fetching items", "error", err)
		return
	}
	pool := pond.NewPool(3)
	for _, item := range items {
		pool.Submit(func() {
			item.Process()
		})
	}
	pool.StopAndWait()
}
