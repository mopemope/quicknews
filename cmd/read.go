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
	NoFetch        bool     `help:"Do not fetch articles background."`
	SpeakingRate   *float64 `short:"s" help:"Set the speaking rate."`
	Voicevox       bool     `help:"Use the voicevox engine." `
	Speaker        int      `help:"Set the voicevox speaker." default:"10"`
	NonInteractive bool     `help:"Run in non-interactive mode without TUI (useful for systemd services)."`
}

// Run executes the TUI command.
func (t *ReadCmd) Run(client *ent.Client, config *config.Config) error {
	if t.NonInteractive {
		slog.Info("Running in non-interactive mode - only background fetching")
	} else {
		if !IsTTY() {
			return errors.New("read command requires TTY for TUI mode. Use --non-interactive flag for non-TTY environments")
		}
		slog.Debug("Starting TUI mode")
	}

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
				fetchArticles(client, config)
				time.Sleep(time.Hour)
			}
		}()
	}

	if !t.NonInteractive {
		model := tui.InitialModel(client, config)
		p := tea.NewProgram(model,
			tea.WithAltScreen(),
			tea.WithMouseCellMotion(),
		)
		if _, err := p.Run(); err != nil {
			return errors.Wrap(err, "error running program")
		}
		slog.Debug("Exiting TUI mode")
	} else {
		slog.Info("Non-interactive mode: Running background fetch only, press Ctrl+C to stop")
		// Keep the process alive for systemd
		select {} // Block forever
	}

	return nil
}
