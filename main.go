package main

import (
	"context"
	"log/slog"

	"github.com/alecthomas/kong"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"github.com/mopemope/quicknews/cmd"
	"github.com/mopemope/quicknews/ent"
	_ "github.com/mopemope/quicknews/pkg/log"
)

// CLI represents the command-line interface.
type CLI struct {
	Add   cmd.AddCmd   `cmd:"" aliases:"a" help:"Add a new RSS feed."`
	Fetch cmd.FetchCmd `cmd:"" aliases:"f" help:"Fetch articles from RSS feeds."`
	Read  cmd.ReadCmd  `cmd:"" aliases:"r" help:"Start read feeds."`
	Play  cmd.PlayCmd  `cmd:"" aliases:"p" help:"Read aloud unlistend feeds."`

	// Global flags
	DbPath string `name:"db" type:"path" default:"~/quicknews.db" help:"Path to the SQLite database file."`
}

func main() {
	var cli CLI
	kctx := kong.Parse(&cli,
		kong.Name("quicknews"),
		kong.Description("A simple RSS reader."),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
	)

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
