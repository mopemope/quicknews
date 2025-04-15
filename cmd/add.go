package cmd

import (
	"context"
	"log/slog"

	"github.com/mmcdole/gofeed"
	"github.com/mopemope/quicknews/ent"
	"github.com/mopemope/quicknews/models/feed"
)

// AddCmd represents the add command.
type AddCmd struct {
	URLs []string `arg:"" name:"url" help:"URLs of the RSS feeds to add." required:""`
}

// Run executes the add command.
func (a *AddCmd) Run(client *ent.Client) error {

	ctx := context.Background()
	fp := gofeed.NewParser()
	repo := feed.NewRepository(client)

	for _, url := range a.URLs {
		// Check if the feed already exists
		exists, err := repo.Exist(ctx, url)
		if err != nil {
			slog.Error("Error checking url", "url", url, "error", err)
			continue // Skip this URL on error
		}
		if exists {
			slog.Info("Feed already exists", "url", url)
			continue
		}

		// Fetch feed information
		parsedFeed, err := fp.ParseURL(url)
		if err != nil {
			slog.Error("Error parsing", "url", url, "error", err)
			continue // Skip this URL on error
		}

		// Create feed in the database
		input := &feed.FeedInput{
			URL:         url,
			Title:       parsedFeed.Title,
			Description: parsedFeed.Description,
			Link:        parsedFeed.Link,
		}
		err = repo.Save(ctx, input, false)
		if err != nil {
			slog.Error("Error saving feed", "url", url, "error", err)
			continue // Skip this URL on error
		}

		slog.Info("Successfully added feed", "title", input.Title, "url", input.URL)
	}

	slog.Info("Add command finished.")
	return nil
}
