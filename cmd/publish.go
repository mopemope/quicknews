package cmd

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/mopemope/quicknews/config"
	"github.com/mopemope/quicknews/ent"
	"github.com/mopemope/quicknews/models/article"
	"github.com/mopemope/quicknews/models/feed"
	"github.com/mopemope/quicknews/rss"
	"github.com/mopemope/quicknews/storage"
	"github.com/mopemope/quicknews/tts"
)

type PublishCmd struct {
	Date string `arg:"" help:"Date to publish the articles in YYYY-MM-DD format."`
	// Output string `short:"o" help:"Output file path for the joined audio."`
}

type publisher struct {
	FeedRepository    feed.FeedRepository
	ArticleRepository article.ArticleRepository
	RSSFeed           *rss.RSS
	R2Client          *storage.R2Storage
	Config            *config.Config
}

func NewPublisher(ctx context.Context, client *ent.Client, config *config.Config) (*publisher, error) {

	feedRepos := feed.NewRepository(client)
	articleRepos := article.NewRepository(client)
	rssFeed := rss.NewRSS(config.Podcast)
	r2client, err := storage.NewR2Storage(ctx, config)
	if err != nil {
		return nil, err
	}

	return &publisher{
		FeedRepository:    feedRepos,
		ArticleRepository: articleRepos,
		RSSFeed:           rssFeed,
		R2Client:          r2client,
		Config:            config,
	}, nil
}

func (c *PublishCmd) Run(client *ent.Client, config *config.Config) error {
	if config.AudioPath == nil || config.Podcast == nil {
		return errors.New("Not support publish. Please set AudioPath and Podcast in config")
	}

	dateRange := 3
	targetDate := c.Date
	if targetDate == "" {
		targetDate = time.Now().Format("2006-01-02")
	}
	targetDateTime, err := time.Parse("2006-01-02", targetDate)
	if err != nil {
		return errors.Wrap(err, "failed to parse target date")
	}
	ctx := context.Background()
	pb, err := NewPublisher(ctx, client, config)
	if err != nil {
		return err
	}

	feedList, err := pb.FeedRepository.All(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get feeds")
	}

	for i := range dateRange {

		tmpDate := targetDateTime.AddDate(0, 0, -i)
		pubDate := tmpDate.Format("2006-01-02")
		for _, f := range feedList {
			if err := pb.processFeed(ctx, f, pubDate); err != nil {
				return err
			}
		}
	}

	if err := pb.publishRSS(ctx); err != nil {
		return err
	}

	return nil
}

func (pb *publisher) processFeed(ctx context.Context, f *ent.Feed, pubDate string) error {
	feedID := f.ID
	feedName := f.Title

	articles, err := pb.ArticleRepository.GetByDate(ctx, feedID, pubDate)
	if err != nil {
		return errors.Wrap(err, "failed to get articles by date")
	}

	infiles := make([]string, 0)
	for _, article := range articles {
		if article.Edges.Summary != nil && article.Edges.Summary.AudioFile != "" {
			infile := filepath.Join(*pb.Config.AudioPath, article.Edges.Summary.AudioFile)
			infiles = append(infiles, infile)
		}
	}

	if len(infiles) == 0 {
		fmt.Printf("No audio files found for feed %s on %s, skipping.\n", feedName, pubDate)
		return nil
	}

	outputFilename := convertPathName(pubDate+"_"+feedName) + ".mp3"
	output := filepath.Join(os.TempDir(), outputFilename)
	defer func() {
		if err := os.Remove(output); err != nil {
			slog.Warn("Failed to remove temporary file", "path", output, "error", err)
		}
	}()

	if err := tts.MergeMP3(output, infiles); err != nil {
		return errors.Wrap(err, "failed to merge mp3 files")
	}

	// アップロード処理
	meta, err := os.Stat(output)
	if err != nil {
		return errors.Wrap(err, "failed to get file info")
	}
	fileSize := meta.Size()

	fileReader, err := os.Open(output)
	if err != nil {
		return errors.Wrap(err, "failed to read output file")
	}
	defer func() {
		if err := fileReader.Close(); err != nil {
			slog.Warn("Failed to close file reader", "path", output, "error", err)
		}
	}()

	if err := pb.R2Client.Upload(ctx, outputFilename, fileReader, "audio/mpeg"); err != nil {
		return errors.Wrap(err, "failed to upload audio file")
	}

	// Add item to RSS feed
	pubdate, err := time.Parse("2006-01-02", pubDate)
	if err != nil {
		return errors.Wrap(err, "failed to parse date")
	}
	podcastConfig := pb.Config.Podcast
	pb.RSSFeed.AddItem(rss.RSSItem{
		Title:       fmt.Sprintf("%s %s Podcast", pubDate, feedName),
		Link:        podcastConfig.PublishURL + "/" + outputFilename,
		Guid:        podcastConfig.PublishURL + "/" + outputFilename,
		PubDate:     pubdate.UTC().Format(time.RFC1123),
		Description: fmt.Sprintf("This is %s %s podcast", pubDate, feedName),
		AudioURL:    podcastConfig.PublishURL + "/" + outputFilename,
		Length:      fmt.Sprintf("%d", fileSize),
		MimeType:    "audio/mpeg",
	})

	return nil
}

func (pb *publisher) publishRSS(ctx context.Context) error {
	rssOutput := filepath.Join(os.TempDir(), "rss.xml")
	defer func() {
		if err := os.Remove(rssOutput); err != nil {
			slog.Warn("Failed to remove temporary RSS file", "path", rssOutput, "error", err)
		}
	}()

	if err := pb.RSSFeed.WriteToFile(rssOutput); err != nil {
		return errors.Wrap(err, "failed to write RSS to file")
	}

	rssFile, err := os.Open(rssOutput)
	if err != nil {
		return errors.Wrap(err, "failed to read RSS file")
	}
	defer func() {
		if err := rssFile.Close(); err != nil {
			slog.Warn("Failed to close RSS file", "path", rssOutput, "error", err)
		}
	}()

	if err := pb.R2Client.Upload(ctx, "rss.xml", rssFile, "application/rss+xml"); err != nil {
		return errors.Wrap(err, "failed to upload RSS file")
	}

	fmt.Println("Successfully published RSS feed.")
	return nil
}

func convertPathName(name string) string {
	h := sha256.New()
	h.Write([]byte(name))
	x := h.Sum(nil)
	return fmt.Sprintf("%x", x)
}
