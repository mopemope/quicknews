package cmd

import (
	"context"
	"log/slog"

	"github.com/cockroachdb/errors"
	"github.com/gilliek/go-opml/opml"
	"github.com/mopemope/quicknews/ent"
	"github.com/mopemope/quicknews/models/feed"
)

// ImportCmd represents the import command.
type ImportCmd struct {
	OpmlPath string `arg:"" name:"opmlfile" type:"path" help:"Path to the OPML file."`

	feedRepos feed.FeedRepository `kong:"-"`
}

// Run executes the import command.
func (cmd *ImportCmd) Run(client *ent.Client) error {
	ctx := context.Background()
	cmd.feedRepos = feed.NewFeedRepository(client)

	doc, err := opml.NewOPMLFromFile(cmd.OpmlPath)
	if err != nil {
		return errors.Wrap(err, "failed to parse OPML file")
	}

	var feedsToSave []*feed.FeedInput
	for _, outline := range doc.Body.Outlines {
		feedsToSave = append(feedsToSave, cmd.extractFeeds(&outline)...)
	}

	if len(feedsToSave) == 0 {
		slog.Info("No new feeds found in the OPML file.")
		return nil
	}

	slog.Info("Importing feeds...", "count", len(feedsToSave))
	if err := cmd.feedRepos.SaveFeeds(ctx, feedsToSave); err != nil {
		return errors.Wrap(err, "failed to save feeds")
	}

	slog.Info("Successfully imported feeds.", "count", len(feedsToSave))
	return nil
}

// extractFeeds recursively extracts feed information from OPML outlines.
func (cmd *ImportCmd) extractFeeds(outline *opml.Outline) []*feed.FeedInput {
	var feeds []*feed.FeedInput

	// If it's a feed entry
	if outline.XMLURL != "" {
		// Check if feed already exists
		exists, err := cmd.feedRepos.Exist(context.Background(), outline.XMLURL) // Use background context for check
		if err != nil {
			slog.Error("Failed to check feed existence, skipping", "url", outline.XMLURL, "error", err)
		} else if !exists {
			feeds = append(feeds, &feed.FeedInput{
				URL:   outline.XMLURL,
				Title: outline.Title, // Use Title if Text is empty
				// Description and Link might not be present in OPML outline, set defaults or leave empty
				Description: "",
				Link:        outline.HTMLURL, // Use HTMLURL for Link if available
			})
			if outline.Text != "" { // Prefer Text over Title if available
				feeds[len(feeds)-1].Title = outline.Text
			}

		} else {
			slog.Info("Feed already exists, skipping", "url", outline.XMLURL)
		}
	}

	// Recursively process sub-outlines
	for _, subOutline := range outline.Outlines {
		feeds = append(feeds, cmd.extractFeeds(&subOutline)...)
	}

	return feeds
}
