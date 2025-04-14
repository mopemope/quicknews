package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/mopemope/quicknews/config"
	"github.com/mopemope/quicknews/ent"
	"github.com/mopemope/quicknews/models/article"
	"github.com/mopemope/quicknews/storage"
	"github.com/mopemope/quicknews/tts"
)

type PublishCmd struct {
	Date string `arg:"" help:"Date to publish the articles in YYYY-MM-DD format."`
	// Output string `short:"o" help:"Output file path for the joined audio."`
}

func (c *PublishCmd) Run(client *ent.Client, config *config.Config) error {
	if config.AudioPath == nil {
		fmt.Println("Audio path is not set in the config.")
		return nil
	}

	date := c.Date
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	repos := article.NewRepository(client)
	ctx := context.Background()
	articles, err := repos.GetByDate(ctx, date)
	if err != nil {
		fmt.Println("Error fetching articles:", err)
		return err
	}

	output := date + ".mp3"
	infiles := make([]string, 0)
	for _, article := range articles {
		if article.Edges.Summary != nil && article.Edges.Summary.AudioFile != "" {
			infile := filepath.Join(*config.AudioPath, article.Edges.Summary.AudioFile)
			infiles = append(infiles, infile)
		}
	}
	if err := tts.MergeMP3(output, infiles); err != nil {
		return err
	}

	fmt.Println("Merged audio files into:", output)

	r2client, err := storage.NewR2Storage(ctx, config)
	if err != nil {
		return err
	}

	f, err := os.Open(output)
	if err != nil {
		return errors.Wrap(err, "failed to read output file")
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println("Error closing file:", err) // Or handle error appropriately
		}
	}()

	if err := r2client.Upload(ctx, output, f); err != nil {
		return err
	}
	fmt.Println("Uploaded to R2:", output)
	return nil
}
