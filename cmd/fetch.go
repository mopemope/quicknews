package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/cockroachdb/errors"
	"github.com/mopemope/quicknews/cmd/fetch"
	"github.com/mopemope/quicknews/config"
	"github.com/mopemope/quicknews/ent"
	"github.com/mopemope/quicknews/models/article"
	"github.com/mopemope/quicknews/models/feed"
	"github.com/mopemope/quicknews/models/summary"
	"github.com/mopemope/quicknews/tui/progress"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// FetchCmd represents the fetch command.
type FetchCmd struct {
	Interval time.Duration `short:"i" help:"Fetch feeds updated within the specified interval (e.g., 24h). Default is 0 (fetch all)."`
}

func (cmd *FetchCmd) Run(client *ent.Client, config *config.Config) error {
	ctx := context.Background()

	feedRepos := feed.NewRepository(client)
	articleRepos := article.NewRepository(client)
	summaryRepos := summary.NewRepository(client)

	feedProcessor := fetch.NewFeedProcessor(feedRepos, articleRepos, summaryRepos, config)

	for {
		items, err := feedProcessor.GetItems(ctx)
		if err != nil {
			return err
		}

		itemCount := len(items)
		if itemCount > 0 {
			if IsTTY() {
				if itemCount > 50 {
					if _, err := tea.NewProgram(progress.NewParallelProgressModel(items, "Fetching", 5)).Run(); err != nil {
						return errors.Wrap(err, "error running progress")
					}
				} else {
					if _, err := tea.NewProgram(progress.NewSingleProgressModel(ctx,
						&progress.Config{
							Client:        client,
							Config:        config,
							Items:         items,
							ProgressLabel: "Fetching",
						})).Run(); err != nil {
						return errors.Wrap(err, "error running progress")
					}
				}
			} else {
				// Non-TTY mode: Process items sequentially without UI
				slog.Info("Processing items in non-TTY mode", "count", itemCount)
				for i, item := range items {
					slog.Info("Processing item", "progress", fmt.Sprintf("%d/%d", i+1, itemCount), "title", item.DisplayName())
					item.Process()
				}
				slog.Info("Finished processing items", "count", itemCount)
			}
		} else {
			fmt.Println("No new items to process.")
		}

		if cmd.Interval > 0 {
			time.Sleep(cmd.Interval)
		} else {
			break
		}
	}
	return nil
}
