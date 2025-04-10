package org

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"strings"

	"github.com/mopemope/quicknews/config"
	"github.com/mopemope/quicknews/ent"
	"github.com/pkg/errors"
)

// ExportOrg exports the summary to an Org file.
func ExportOrg(config *config.Config, sum *ent.Summary) error {
	dst := config.ExportOrg
	if dst == "" {
		slog.Info("EXPORT_ORG is not set")
		return nil
	}
	if sum.Edges.Feed == nil || sum.Edges.Article == nil {
		slog.Info("Feed or Article is nil")
		return nil
	}
	feed := sum.Edges.Feed
	article := sum.Edges.Article

	dst = path.Join(dst, convertPathName(feed.Title))
	if err := os.MkdirAll(dst, os.ModePerm); err != nil {
		return errors.Wrap(err, "failed to create directory")
	}

	timestamp := sum.CreatedAt.Format("20060102150405")
	orgFile := timestamp + "-" + convertPathName(sum.Title) + ".org"
	dst = path.Join(dst, orgFile)

	contentTemplate := `:PROPERTIES:
:ID:       %s
:FEEDURL:  %s
:FEED:     %s
:LINK:     %s
:TITLE:    %s
:END:
#+TITLE:   %s
#+TAGS: feed
#+STARTUP: overview
#+STARTUP: inlineimages
#+OPTIONS: ^:nil

# [[%s][%s]]

%s
`
	content := fmt.Sprintf(contentTemplate,
		sum.ID,
		feed.URL,
		feed.Title,
		article.URL,
		article.Title,
		sum.Title,
		sum.URL,
		sum.Title,
		sum.Summary)
	return os.WriteFile(dst, []byte(content), os.ModePerm)
}

func convertPathName(name string) string {
	return strings.ReplaceAll(name, " ", "_")
}
