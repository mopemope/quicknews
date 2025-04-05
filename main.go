package main

import (
	"context"
	"log/slog"

	"github.com/alecthomas/kong"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"github.com/mopemope/quicknews/cmd"
	"github.com/mopemope/quicknews/ent"
	"github.com/mopemope/quicknews/pkg/log" // Import log package
)

var version = "0.0.1"

// CLI represents the command-line interface.
type CLI struct {
	Add    cmd.AddCmd    `cmd:"" aliases:"a" help:"Add a new RSS feed."`
	Fetch  cmd.FetchCmd  `cmd:"" aliases:"f" help:"Fetch articles from RSS feeds."`
	Read   cmd.ReadCmd   `cmd:"" aliases:"r" help:"Start read feeds."`
	Play   cmd.PlayCmd   `cmd:"" aliases:"p" help:"Read aloud unlistend feeds."`
	Import cmd.ImportCmd `cmd:"" help:"Import feeds from an OPML file."`

	// Global flags
	DbPath  string `name:"db" type:"path" default:"~/quicknews.db" help:"Path to the SQLite database file."`
	LogPath string `name:"log" type:"path" default:"quicknews.log"  help:"Path to the log file. If not specified, logs to stdout."`

	// Version flag
	Version kong.VersionFlag `short:"V" help:"Show version information."`
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
	if err := log.InitializeLogger(cli.LogPath); err != nil {
		slog.Error("failed to initialize logger", "error", err)
		return
	}
	// Initialize database client
	client, err := ent.Open("sqlite3", cli.DbPath+"??cache=shared&_fk=1")
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
	}

	kctx.Bind(client)

	// Call the Run() method of the selected parsed command.
	err = kctx.Run()
	kctx.FatalIfErrorf(err)
}
