package cmd

import (
	"github.com/mopemope/quicknews/ent"
	"github.com/mopemope/quicknews/mcpserver"
	"github.com/mopemope/quicknews/models/article"
)

type MCPCmd struct{}

func (c *MCPCmd) Run(client *ent.Client) error {
	repo := article.NewRepository(client)
	return mcpserver.RunStdio(RunContext(), repo)
}
