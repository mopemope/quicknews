package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/cockroachdb/errors"
	"github.com/mopemope/quicknews/ent"
	"github.com/mopemope/quicknews/models/summary"
	"github.com/mopemope/quicknews/tts"

	"github.com/mopemope/quicknews/tui/progress"
)

type PlayCmd struct {
	SpeakingRate float64 `short:"s" help:"Set the speaking rate." default:"1.5"`
	NoFetch      bool    `help:"Do not fetch articles background."`
}

type playArticle struct {
	summary *ent.Summary
	repo    summary.SummaryRepository
}

func (a *playArticle) DisplayName() string {
	return a.summary.Title
}

func (a *playArticle) Process() {
	ctx := context.Background()
	audioData, err := summary.GetAudioData(ctx, a.summary)
	if err != nil {
		slog.Error("Failed to get audio data", "error", err)
		return
	}

	ttsEngine := tts.NewTTSEngine()
	if err := ttsEngine.PlayAudioData(audioData); err != nil {
		slog.Error("failed to play audio data", "error", err)
		return
	}
	// Update the summary as listened
	if err := a.repo.UpdateListened(ctx, a.summary); err != nil {
		slog.Error("failed to update listened status", "error", err)
	}
}

func newArticle(summary *ent.Summary, repo summary.SummaryRepository) *playArticle {
	return &playArticle{
		summary: summary,
		repo:    repo,
	}
}

func (a *PlayCmd) Run(client *ent.Client) error {

	tts.SpeachOpt.SpeakingRate = a.SpeakingRate
	ctx := context.Background()
	if !a.NoFetch {
		go func() {
			for {
				fetchArticles(client)
				time.Sleep(time.Hour)
			}
		}()
	}

	repo := summary.NewSummaryRepository(client)

	res, err := repo.GetUnlistened(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get unlistened summaries")
	}

	items := make([]progress.QueueItem, 0)
	for _, sum := range res {
		items = append(items, newArticle(sum, repo))
	}

	if len(items) > 0 {
		if _, err := tea.NewProgram(progress.NewModel(items)).Run(); err != nil {
			return errors.Wrap(err, "error running progress")
		}
	} else {
		fmt.Println("No new items to process.")
	}

	return nil
}
