package cmd

import (
	"context"
	"log/slog"

	pond "github.com/alitto/pond/v2"
	"github.com/mopemope/quicknews/ent"
	"github.com/mopemope/quicknews/models/article"
	"github.com/mopemope/quicknews/models/feed"
	"github.com/mopemope/quicknews/models/summary"
)

func fetchArticles(client *ent.Client) {
	fetchCmd := FetchCmd{
		feedRepos:    feed.NewFeedRepository(client),
		articleRepos: article.NewArticleRepository(client),
		summaryRepos: summary.NewRepository(client),
	}
	ctx := context.Background()
	items, err := fetchCmd.getItems(ctx)
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
