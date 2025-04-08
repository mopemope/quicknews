package cmd

import (
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/cockroachdb/errors"
	"github.com/mopemope/quicknews/ent"
	"github.com/mopemope/quicknews/tui"
)

// ReadCmd represents the TUI command.
type ReadCmd struct {
	Confirm bool `short:"c" help:"Ask for confirmation before action."`
	// TUI command specific flags can be added here.
}

// Run executes the TUI command.
func (t *ReadCmd) Run(client *ent.Client) error {
	slog.Debug("Starting TUI mode")

	go func() {
		fetchArticles(client)
	}()

	model := tui.InitialModel(client, t.Confirm)
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
