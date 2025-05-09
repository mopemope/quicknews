package org

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/mopemope/quicknews/config"
	"github.com/mopemope/quicknews/ent"
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

	dst = path.Join(dst, ConvertPathName(feed.Title))
	if err := os.MkdirAll(dst, os.ModePerm); err != nil {
		return errors.Wrap(err, "failed to create directory")
	}

	timestamp := sum.CreatedAt.Format("20060102150405")
	orgFile := timestamp + "-" + ConvertPathName(sum.Title) + ".org"
	dst = path.Join(dst, orgFile)

	contentTemplate := `:PROPERTIES:
:ID:       %s
:FEEDURL:  %s
:FEED:     %s
:LINK:     %s
:TITLE:    %s [%s]
:END:
#+TITLE:   %s [%s]
#+TAGS: feed
#+STARTUP: overview
#+STARTUP: inlineimages
#+OPTIONS: ^:nil

# [[%s][%s %s]]

%s
`
	content := fmt.Sprintf(contentTemplate,
		sum.ID,
		feed.URL,
		feed.Title,
		article.URL,
		article.Title,
		article.URL,
		sum.Title,
		sum.URL,
		sum.URL,
		sum.Title,
		sum.URL,
		sum.Summary)
	return os.WriteFile(dst, []byte(content), os.ModePerm)
}

// ConvertPathName converts a string to a safe path name component
// by replacing spaces and other problematic characters with underscores.
func ConvertPathName(name string) string {
	s := strings.ReplaceAll(name, " ", "_")
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, ":", "_")
	s = strings.ReplaceAll(s, "?", "_")
	return s
}
