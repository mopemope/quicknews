package mcpserver

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/cockroachdb/errors"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/mopemope/quicknews/ent"
	articlemodel "github.com/mopemope/quicknews/models/article"
)

const snippetLimit = 240

type ArticleSearcher interface {
	Search(ctx context.Context, options articlemodel.SearchOptions) (ent.Articles, error)
}

type Server struct {
	articles ArticleSearcher
}

type SearchArticlesInput struct {
	Query  string `json:"query" jsonschema:"Keyword query. Multiple words are treated as AND conditions."`
	Limit  int    `json:"limit,omitempty" jsonschema:"Maximum number of articles to return. Defaults to 10 and is capped at 50."`
	Offset int    `json:"offset,omitempty" jsonschema:"Number of matching articles to skip. Defaults to 0."`
}

type SearchArticlesOutput struct {
	Query   string                `json:"query"`
	Count   int                   `json:"count"`
	Limit   int                   `json:"limit"`
	Offset  int                   `json:"offset"`
	Results []SearchArticleResult `json:"results"`
}

type SearchArticleResult struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	URL           string   `json:"url"`
	FeedTitle     string   `json:"feed_title"`
	PublishedAt   string   `json:"published_at,omitempty"`
	Snippet       string   `json:"snippet,omitempty"`
	MatchedFields []string `json:"matched_fields"`
	Readed        bool     `json:"readed"`
	Listened      bool     `json:"listened"`
}

func New(articleSearcher ArticleSearcher) *Server {
	return &Server{articles: articleSearcher}
}

func RunStdio(ctx context.Context, articleSearcher ArticleSearcher) error {
	return New(articleSearcher).MCPServer().Run(ctx, &mcp.StdioTransport{})
}

func (s *Server) MCPServer() *mcp.Server {
	openWorld := false
	destructive := false
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "quicknews",
		Version: "0.0.1",
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_articles",
		Title:       "Search quicknews articles",
		Description: "Search saved quicknews articles by keyword across article text and summaries.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:    true,
			DestructiveHint: &destructive,
			OpenWorldHint:   &openWorld,
		},
	}, s.SearchArticles)

	return server
}

func (s *Server) SearchArticles(ctx context.Context, _ *mcp.CallToolRequest, input SearchArticlesInput) (*mcp.CallToolResult, SearchArticlesOutput, error) {
	query := strings.TrimSpace(input.Query)
	if query == "" {
		return nil, SearchArticlesOutput{}, errors.New("query is required")
	}

	options := normalizeSearchOptions(query, input.Limit, input.Offset)
	articles, err := s.articles.Search(ctx, options)
	if err != nil {
		return nil, SearchArticlesOutput{}, err
	}

	output := SearchArticlesOutput{
		Query:   query,
		Count:   len(articles),
		Limit:   options.Limit,
		Offset:  options.Offset,
		Results: make([]SearchArticleResult, 0, len(articles)),
	}
	terms := strings.Fields(query)
	for _, article := range articles {
		output.Results = append(output.Results, formatArticleResult(article, terms))
	}

	return nil, output, nil
}

func normalizeSearchOptions(query string, limit, offset int) articlemodel.SearchOptions {
	if limit <= 0 {
		limit = articlemodel.DefaultSearchLimit
	}
	if limit > articlemodel.MaxSearchLimit {
		limit = articlemodel.MaxSearchLimit
	}
	if offset < 0 {
		offset = 0
	}
	return articlemodel.SearchOptions{
		Query:  query,
		Limit:  limit,
		Offset: offset,
	}
}

func formatArticleResult(article *ent.Article, terms []string) SearchArticleResult {
	result := SearchArticleResult{
		ID:            article.ID.String(),
		Title:         article.Title,
		URL:           article.URL,
		Snippet:       buildSnippet(article, terms),
		MatchedFields: matchedFields(article, terms),
	}

	if !article.PublishedAt.IsZero() {
		result.PublishedAt = article.PublishedAt.UTC().Format("2006-01-02T15:04:05Z07:00")
	}
	if article.Edges.Feed != nil {
		result.FeedTitle = article.Edges.Feed.Title
	}
	if article.Edges.Summary != nil {
		result.Readed = article.Edges.Summary.Readed
		result.Listened = article.Edges.Summary.Listened
	}
	return result
}

func matchedFields(article *ent.Article, terms []string) []string {
	fields := make([]string, 0, 5)
	addIfMatched := func(name, value string) {
		if containsAnyTerm(value, terms) {
			fields = append(fields, name)
		}
	}

	addIfMatched("article.title", article.Title)
	addIfMatched("article.description", article.Description)
	addIfMatched("article.content", article.Content)
	if article.Edges.Summary != nil {
		addIfMatched("summary.title", article.Edges.Summary.Title)
		addIfMatched("summary.summary", article.Edges.Summary.Summary)
	}
	return fields
}

func buildSnippet(article *ent.Article, terms []string) string {
	candidates := make([]string, 0, 5)
	if article.Edges.Summary != nil {
		candidates = append(candidates, article.Edges.Summary.Summary, article.Edges.Summary.Title)
	}
	candidates = append(candidates, article.Description, article.Content, article.Title)

	for _, candidate := range candidates {
		if containsAnyTerm(candidate, terms) {
			return truncate(candidate, snippetLimit)
		}
	}
	for _, candidate := range candidates {
		if strings.TrimSpace(candidate) != "" {
			return truncate(candidate, snippetLimit)
		}
	}
	return ""
}

func containsAnyTerm(value string, terms []string) bool {
	lower := strings.ToLower(value)
	for _, term := range terms {
		if strings.Contains(lower, strings.ToLower(term)) {
			return true
		}
	}
	return false
}

func truncate(value string, limit int) string {
	value = strings.Join(strings.Fields(value), " ")
	if utf8.RuneCountInString(value) <= limit {
		return value
	}

	runes := []rune(value)
	if limit <= 3 {
		return string(runes[:limit])
	}
	return string(runes[:limit-3]) + "..."
}
