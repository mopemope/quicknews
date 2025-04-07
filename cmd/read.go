package cmd

import (
	"context"
	"log/slog"

	pond "github.com/alitto/pond/v2"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/cockroachdb/errors"
	"github.com/mopemope/quicknews/ent"
	"github.com/mopemope/quicknews/models/article"
	"github.com/mopemope/quicknews/models/feed"
	"github.com/mopemope/quicknews/models/summary"
	"github.com/mopemope/quicknews/tui"
)

// ReadCmd represents the TUI command.
type ReadCmd struct {
	// TUI command specific flags can be added here.
}

// Run executes the TUI command.
func (t *ReadCmd) Run(client *ent.Client) error {
	slog.Debug("Starting TUI mode")

	go func() {
		fetchCmd := FetchCmd{
			feedRepos:    feed.NewFeedRepository(client),
			articleRepos: article.NewArticleRepository(client),
			summaryRepos: summary.NewSummaryRepository(client),
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
	}()

	model := tui.InitialModel(client)
	p := tea.NewProgram(model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	if _, err := p.Run(); err != nil {
		return errors.Wrap(err, "error running program")
	}
	slog.Debug("Exiting TUI mode")
	return nil
}
