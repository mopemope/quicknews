package cmd

import (
	stderrors "errors"
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
	ctx := RunContext()

	feedRepos := feed.NewRepository(client)
	articleRepos := article.NewRepository(client)
	summaryRepos := summary.NewRepository(client)

	feedProcessor := fetch.NewFeedProcessor(feedRepos, articleRepos, summaryRepos, config)

	for {
		items, err := feedProcessor.GetItems(ctx)
		if err != nil && len(items) == 0 {
			return err
		}
		pendingErr := err
		if pendingErr != nil {
			slog.Error("failed to fetch one or more feeds", "error", pendingErr)
		}

		itemCount := len(items)
		if itemCount > 0 {
			if IsTTY() {
				if itemCount > 50 {
					model := progress.NewParallelProgressModel(items, "Fetching", 5)
					finalModel, err := tea.NewProgram(model).Run()
					if err != nil {
						return errors.Wrap(err, "error running progress")
					}
					pendingErr = stderrors.Join(pendingErr, progressModelErr(finalModel))
				} else {
					model := progress.NewSingleProgressModel(ctx,
						&progress.Config{
							Client:        client,
							Config:        config,
							Items:         items,
							ProgressLabel: "Fetching",
						})
					finalModel, err := tea.NewProgram(model).Run()
					if err != nil {
						return errors.Wrap(err, "error running progress")
					}
					pendingErr = stderrors.Join(pendingErr, progressModelErr(finalModel))
				}
			} else {
				// Non-TTY mode: Process items sequentially without UI
				slog.Info("Processing items in non-TTY mode", "count", itemCount)
				itemErrs := make([]error, 0)
				for i, item := range items {
					slog.Info("Processing item", "progress", fmt.Sprintf("%d/%d", i+1, itemCount), "title", item.DisplayName())
					if err := item.Process(); err != nil {
						slog.Error("failed to process item", "title", item.DisplayName(), "url", item.URL(), "error", err)
						itemErrs = append(itemErrs, err)
					}
				}
				slog.Info("Finished processing items", "count", itemCount)
				pendingErr = stderrors.Join(pendingErr, stderrors.Join(itemErrs...))
			}
		} else {
			fmt.Println("No new items to process.")
		}

		if pendingErr != nil {
			return pendingErr
		}

		if cmd.Interval > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(cmd.Interval):
			}
		} else {
			break
		}
	}
	return nil
}

func progressModelErr(model tea.Model) error {
	errModel, ok := model.(interface{ Err() error })
	if !ok {
		return nil
	}
	return errModel.Err()
}
