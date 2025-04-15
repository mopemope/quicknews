package main

import (
	"context"
	"log/slog"

	"github.com/alecthomas/kong"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"github.com/mopemope/quicknews/cmd"
	"github.com/mopemope/quicknews/config"
	"github.com/mopemope/quicknews/ent"
	"github.com/mopemope/quicknews/log" // Import log package
	"github.com/mopemope/quicknews/models/feed"
)

var version = "0.0.1"

// CLI represents the command-line interface.
type CLI struct {
	Add         cmd.AddCmd         `cmd:"" aliases:"a" help:"Add a new RSS feed."`
	Fetch       cmd.FetchCmd       `cmd:"" aliases:"f" help:"Fetch articles from RSS feeds."`
	Read        cmd.ReadCmd        `cmd:"" aliases:"r" help:"Start read feeds."`
	Play        cmd.PlayCmd        `cmd:"" aliases:"p" help:"Read aloud unlistend feeds."`
	Import      cmd.ImportCmd      `cmd:"" help:"Import feeds from an OPML file."`
	Bookmark    cmd.BookmarkCmd    `cmd:"" aliases:"b" help:"Add a new bookmark."`
	ExportAudio cmd.ExportAudioCmd `cmd:""  help:"Export audio files."`
	Publish     cmd.PublishCmd     `cmd:"" help:"Publish articles."`

	// Global flags
	ConfigPath string           `name:"config" type:"path" default:"~/.config/quicknews/config.toml" help:"Path to the config file."`
	LogPath    string           `name:"log" type:"path" default:"~/quicknews.log"  help:"Path to the log file. If not specified, logs to stdout."`
	Version    kong.VersionFlag `short:"V" help:"Show version information."`
	Debug      bool             `short:"d" help:"Enable debug logging."`

	// Internal fields
	config *config.Config
}

func main() {
	var cli CLI
	kctx := kong.Parse(&cli,
		kong.Name("quicknews"),
		kong.Description("RSS reader."),
		kong.UsageOnError(),
		kong.Vars{"version": version}, // Pass version variable for --version flag
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
	)

	cfg, err := config.LoadConfig(cli.ConfigPath)
	if err != nil {
		slog.Error("failed to load config", "error", err)
		return
	}
	cli.config = cfg

	if err := log.InitializeLogger(cli.LogPath, cli.Debug); err != nil {
		slog.Error("failed to initialize logger", "error", err)
		return
	}
	// Initialize database client
	client, err := ent.Open("sqlite3", cfg.DB+"?cache=shared&_fk=1")
	if err != nil {
		slog.Error("failed opening connection to sqlite", "error", err)
		return
	}
	defer func() {
		if err := client.Close(); err != nil {
			slog.Error("failed closing connection to sqlite", "error", err)
		}
	}()

	ctx := context.Background()
	// Run the auto migration tool.
	if err := client.Schema.Create(ctx); err != nil {
		slog.Error("failed creating schema resources", "error", err)
		return
	}

	if err := setup(ctx, client); err != nil {
		slog.Error("failed to setup initial data", "error", err)
		return
	}
	kctx.Bind(client, cli.config)

	// Call the Run() method of the selected parsed command.
	err = kctx.Run()
	kctx.FatalIfErrorf(err)
}

func setup(ctx context.Context, cilent *ent.Client) error {
	repo := feed.NewRepository(cilent)
	exist, err := repo.ExistBookmarkFeed(ctx)
	if err != nil {
		return err
	}
	if !exist {
		input := &feed.FeedInput{
			URL:         "https://quicknews.org/bookmark/rss",
			Title:       "Bookmark",
			Description: "Bookmark",
			Link:        "https://quicknews.org/bookmark/rss",
		}
		if err := repo.Save(ctx, input, true); err != nil {
			return err
		}
	}
	return nil
}
