package cmd

import (
	"log/slog"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/cockroachdb/errors"
	"github.com/mopemope/quicknews/config"
	"github.com/mopemope/quicknews/ent"
	"github.com/mopemope/quicknews/tts"
	"github.com/mopemope/quicknews/tui"
)

// ReadCmd represents the TUI command.
type ReadCmd struct {
	NoFetch      bool     `help:"Do not fetch articles background."`
	SpeakingRate *float64 `short:"s" help:"Set the speaking rate."`
	Voicevox     bool     `help:"Use the voicevox engine." `
	Speaker      int      `help:"Set the voicevox speaker." default:"10"`
}

// Run executes the TUI command.
func (t *ReadCmd) Run(client *ent.Client, config *config.Config) error {
	slog.Debug("Starting TUI mode")

	if t.SpeakingRate == nil {
		t.SpeakingRate = &config.SpeakingRate
	}
	tts.SpeachOpt.SpeakingRate = *t.SpeakingRate

	if config.VoiceVox != nil {
		tts.SpeachOpt.Engine = "voicevox"
		tts.SpeachOpt.Speaker = config.VoiceVox.Speaker
	}

	if !t.NoFetch {
		go func() {
			for {
				fetchArticles(client)
				time.Sleep(time.Hour)
			}
		}()
	}

	model := tui.InitialModel(client, config)
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
