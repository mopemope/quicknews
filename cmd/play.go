package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/cockroachdb/errors"
	"github.com/mopemope/quicknews/config"
	"github.com/mopemope/quicknews/ent"
	"github.com/mopemope/quicknews/models/summary"
	"github.com/mopemope/quicknews/tts"

	"github.com/mopemope/quicknews/tui/progress"
)

type PlayCmd struct {
	NoFetch bool    `help:"Do not fetch articles background."`
	Date    *string "help:\"Set the date to play articles in YYYY-MM-DD format\""
}

type playArticle struct {
	summary *ent.Summary
	repo    summary.SummaryRepository
	config  *config.Config
}

func (a *playArticle) DisplayName() string {
	return a.summary.Edges.Article.Title
}

func (a *playArticle) URL() string {
	return a.summary.URL
}

func (a *playArticle) Process() {
	ctx := context.Background()
	// Pass config to GetAudioData
	audioData, err := summary.GetAudioData(ctx, a.summary, a.config)
	if err != nil {
		slog.Error("Failed to get audio data", "error", err)
		return
	}

	ttsEngine := tts.NewTTSEngine(a.config)
	if err := ttsEngine.PlayAudioData(audioData); err != nil {
		slog.Error("failed to play audio data", "error", err)
		return
	}
	// Update the summary as listened
	if err := a.repo.UpdateListened(ctx, a.summary); err != nil {
		slog.Error("failed to update listened status", "error", err)
	}
}

func newArticle(summary *ent.Summary, repo summary.SummaryRepository, config *config.Config) *playArticle {
	return &playArticle{
		summary: summary,
		repo:    repo,
		config:  config,
	}
}

func (a *PlayCmd) Run(client *ent.Client, config *config.Config) error {
	tts.SpeachOpt.SpeakingRate = config.SpeakingRate
	if config.VoiceVox != nil {
		tts.SpeachOpt.Engine = "voicevox"
		tts.SpeachOpt.Speaker = config.VoiceVox.Speaker
	}

	ctx := context.Background()
	if !a.NoFetch {
		go func() {
			for {
				fetchArticles(client, config)
				time.Sleep(time.Hour)
			}
		}()
	}

	repo := summary.NewRepository(client)

	res, err := repo.GetUnlistened(ctx, a.Date)
	if err != nil {
		return errors.Wrap(err, "failed to get unlistened summaries")
	}

	items := make([]progress.QueueItem, 0)
	for _, sum := range res {
		items = append(items, newArticle(sum, repo, config))
	}

	if len(items) > 0 {
		if IsTTY() {
			if _, err := tea.NewProgram(progress.NewSingleProgressModel(ctx,
				&progress.Config{
					Client:        client,
					Config:        config,
					Items:         items,
					ProgressLabel: "Playing",
				})).Run(); err != nil {
				return errors.Wrap(err, "error running progress")
			}
		} else {
			// Non-TTY mode: Process items sequentially without UI
			slog.Info("Playing items in non-TTY mode", "count", len(items))
			for i, item := range items {
				slog.Info("Playing item", "progress", fmt.Sprintf("%d/%d", i+1, len(items)), "title", item.DisplayName())
				item.Process()
			}
			slog.Info("Finished playing items", "count", len(items))
		}
	} else {
		fmt.Println("No new items to process.")
	}

	return nil
}
