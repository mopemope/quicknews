package mcpserver

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/mopemope/quicknews/ent"
	articlemodel "github.com/mopemope/quicknews/models/article"
	"github.com/stretchr/testify/require"
)

type fakeArticleSearcher struct {
	options  articlemodel.SearchOptions
	articles ent.Articles
}

func (s *fakeArticleSearcher) Search(_ context.Context, options articlemodel.SearchOptions) (ent.Articles, error) {
	s.options = options
	return s.articles, nil
}

func TestSearchArticles(t *testing.T) {
	searcher := &fakeArticleSearcher{
		articles: ent.Articles{
			{
				ID:          uuid.MustParse("00000000-0000-0000-0000-000000000001"),
				Title:       "Go MCP Integration",
				URL:         "https://example.com/go-mcp",
				Description: "Use quicknews with AI tools.",
				Content:     "Long article body",
				PublishedAt: time.Date(2026, 4, 22, 10, 0, 0, 0, time.FixedZone("JST", 9*60*60)),
				Edges: ent.ArticleEdges{
					Feed: &ent.Feed{Title: "Tech Feed"},
					Summary: &ent.Summary{
						Title:    "MCP summary",
						Summary:  "Quicknews articles can be searched from other AI tools.",
						Readed:   true,
						Listened: false,
					},
				},
			},
		},
	}

	_, output, err := New(searcher).SearchArticles(context.Background(), nil, SearchArticlesInput{
		Query:  " mcp ai ",
		Limit:  100,
		Offset: -5,
	})
	require.NoError(t, err)
	require.Equal(t, "mcp ai", output.Query)
	require.Equal(t, articlemodel.MaxSearchLimit, output.Limit)
	require.Equal(t, 0, output.Offset)
	require.Equal(t, articlemodel.SearchOptions{
		Query:  "mcp ai",
		Limit:  articlemodel.MaxSearchLimit,
		Offset: 0,
	}, searcher.options)
	require.Len(t, output.Results, 1)

	result := output.Results[0]
	require.Equal(t, "00000000-0000-0000-0000-000000000001", result.ID)
	require.Equal(t, "Go MCP Integration", result.Title)
	require.Equal(t, "https://example.com/go-mcp", result.URL)
	require.Equal(t, "Tech Feed", result.FeedTitle)
	require.Equal(t, "2026-04-22T01:00:00Z", result.PublishedAt)
	require.Equal(t, "Quicknews articles can be searched from other AI tools.", result.Snippet)
	require.ElementsMatch(t, []string{"article.title", "article.description", "summary.title", "summary.summary"}, result.MatchedFields)
	require.True(t, result.Readed)
	require.False(t, result.Listened)
}

func TestSearchArticlesRequiresQuery(t *testing.T) {
	_, _, err := New(&fakeArticleSearcher{}).SearchArticles(context.Background(), nil, SearchArticlesInput{})
	require.Error(t, err)
	require.ErrorContains(t, err, "query is required")
}

func TestMCPServerRegistersSearchArticles(t *testing.T) {
	ctx := context.Background()
	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	searcher := &fakeArticleSearcher{articles: ent.Articles{}}
	serverSession, err := New(searcher).MCPServer().Connect(ctx, serverTransport, nil)
	require.NoError(t, err)

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)

	result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: "search_articles",
		Arguments: map[string]any{
			"query": "mcp",
		},
	})
	require.NoError(t, err)
	require.False(t, result.IsError)
	require.NotEmpty(t, result.Content)
	require.NotNil(t, result.StructuredContent)

	clientSession.Close()
	serverSession.Wait()
}

func TestTruncate(t *testing.T) {
	value := strings.Repeat("a", snippetLimit+10)
	got := truncate(value, snippetLimit)
	require.Len(t, []rune(got), snippetLimit)
	require.True(t, strings.HasSuffix(got, "..."))
}
