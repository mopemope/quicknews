package cmd

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
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

func validateReadMode(nonInteractive bool, isTTY bool) error {
	if nonInteractive {
		return nil
	}
	if !isTTY {
		return errors.New("read command requires TTY for TUI mode. Use --non-interactive flag for non-TTY environments")
	}
	return nil
}

// Run executes the TUI command.
func (t *ReadCmd) Run(client *ent.Client, cfg *config.Config) error {
	if err := validateReadMode(t.NonInteractive, IsTTY()); err != nil {
		return err
	}

	if t.NonInteractive {
		slog.Info("Running in non-interactive mode - only background fetching")
	} else {
		slog.Debug("Starting TUI mode")
	}

	t.applyTTSConfig(cfg)
	ctx, stop := signal.NotifyContext(RunContext(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if !t.NoFetch {
		go func() {
			fetchTicker := time.NewTicker(time.Hour)
			defer fetchTicker.Stop()

			for {
				fetchArticles(ctx, client, cfg)
				select {
				case <-fetchTicker.C:
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	if !t.NonInteractive {
		model := tui.InitialModelWithContext(ctx, client, cfg)
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
		<-ctx.Done()
	}

	return nil
}

func (t *ReadCmd) applyTTSConfig(cfg *config.Config) {
	if t.SpeakingRate == nil {
		t.SpeakingRate = &cfg.SpeakingRate
	}
	tts.SpeachOpt.SpeakingRate = *t.SpeakingRate

	if t.Voicevox {
		tts.SpeachOpt.Engine = "voicevox"
		tts.SpeachOpt.Speaker = t.Speaker
		if cfg.VoiceVox == nil {
			cfg.VoiceVox = &config.VoiceVox{Speaker: t.Speaker}
		} else {
			cfg.VoiceVox.Speaker = t.Speaker
		}
	} else if cfg.VoiceVox != nil {
		tts.SpeachOpt.Engine = "voicevox"
		tts.SpeachOpt.Speaker = cfg.VoiceVox.Speaker
	}
}
