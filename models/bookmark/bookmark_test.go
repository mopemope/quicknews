package bookmark

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mopemope/quicknews/config"
	"github.com/mopemope/quicknews/ent/article"
	"github.com/mopemope/quicknews/ent/enttest"
	"github.com/mopemope/quicknews/ent/summary"
	"github.com/mopemope/quicknews/tts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBookmarkRepository(t *testing.T) {
	client := enttest.Open(t, dialect.SQLite, "file:ent?mode=memory&cache=shared&_fk=1")
	defer func() { _ = client.Close() }()

	// Create a minimal config for testing
	config := &config.Config{
		SaveAudioData: false,
	}

	// First, create a bookmark feed (this is needed for bookmark functionality)
	bookmarkFeed, err := client.Feed.Create().
		SetURL("https://quicknews.org/bookmark/rss").
		SetTitle("Bookmark").
		SetDescription("Bookmark").
		SetLink("https://quicknews.org/bookmark/rss").
		SetUpdatedAt(time.Now()).
		SetIsBookmark(true).
		Save(context.Background())
	require.NoError(t, err)

	// Create the bookmark repository
	repo, err := NewRepository(context.Background(), client, config)
	require.NoError(t, err)

	ctx := context.Background()

	// Get the bookmark feed
	fetchedBookmarkFeed, err := repo.GetBookmarkFeed(ctx)
	require.NoError(t, err)
	assert.Equal(t, bookmarkFeed.ID, fetchedBookmarkFeed.ID)
}

func TestBookmarkRepository_AddBookmark(t *testing.T) {
	client := enttest.Open(t, dialect.SQLite, "file:ent?mode=memory&cache=shared&_fk=1")
	defer func() { _ = client.Close() }()

	// Create a minimal config for testing
	config := &config.Config{
		SaveAudioData: false,
	}

	// First, create a bookmark feed (this is needed for bookmark functionality)
	bookmarkFeed, err := client.Feed.Create().
		SetURL("https://quicknews.org/bookmark/rss").
		SetTitle("Bookmark").
		SetDescription("Bookmark").
		SetLink("https://quicknews.org/bookmark/rss").
		SetUpdatedAt(time.Now()).
		SetIsBookmark(true).
		Save(context.Background())
	require.NoError(t, err)

	feed, err := client.Feed.Create().
		SetURL("https://quicknews.org/feed").
		SetTitle("Source Feed").
		SetDescription("Source Feed").
		SetLink("https://quicknews.org/feed").
		SetUpdatedAt(time.Now()).
		Save(context.Background())
	require.NoError(t, err)

	url := "https://example.com/article"
	articleEntity, err := client.Article.Create().
		SetTitle("Example").
		SetURL(url).
		SetDescription("desc").
		SetContent("content").
		SetCreatedAt(time.Now()).
		SetPublishedAt(time.Now()).
		SetFeed(feed).
		Save(context.Background())
	require.NoError(t, err)

	_, err = client.Summary.Create().
		SetTitle("Example Summary").
		SetSummary("summary").
		SetURL(url).
		SetCreatedAt(time.Now()).
		SetArticle(articleEntity).
		SetFeed(feed).
		Save(context.Background())
	require.NoError(t, err)

	// Create the bookmark repository
	repo, err := NewRepository(context.Background(), client, config)
	require.NoError(t, err)

	ctx := context.Background()

	require.NoError(t, repo.AddBookmark(ctx, url))

	updatedArticle, err := client.Article.Query().
		Where(article.URL(url)).
		WithFeed().
		Only(ctx)
	require.NoError(t, err)
	assert.Equal(t, bookmarkFeed.ID, updatedArticle.Edges.Feed.ID)
	updatedSummary, err := client.Summary.Query().
		Where(summary.URL(url)).
		WithFeed().
		Only(ctx)
	require.NoError(t, err)
	require.NotNil(t, updatedSummary.Edges.Feed)
	assert.True(t, updatedSummary.Edges.Feed.IsBookmark)
}

func TestNewRepository_SucceedsWithoutGeminiKey(t *testing.T) {
	client := enttest.Open(t, dialect.SQLite, "file:ent?mode=memory&cache=shared&_fk=1")
	defer func() { _ = client.Close() }()

	_, err := NewRepository(context.Background(), client, &config.Config{})
	require.NoError(t, err)
}

// TestBookmarkRepository_PrepareAudioPassesFeedEdge guards the regression
// where prepareAudio built a summary without the Feed edge, making
// GetAudioData fail with "summary feed edge is not loaded" whenever
// save_audio_data was enabled. It uses the VoiceVox engine against an
// httptest server so no external credentials are needed.
func TestBookmarkRepository_PrepareAudioPassesFeedEdge(t *testing.T) {
	client := enttest.Open(t, dialect.SQLite, "file:ent?mode=memory&cache=shared&_fk=1")
	defer func() { _ = client.Close() }()

	audioDir := t.TempDir()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/article":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(`<html><head><title>テスト記事</title></head><body><p>` + strings.Repeat("本文です。", 40) + `</p></body></html>`))
		case "/speakers":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[{"name":"test","speaker_uuid":"u","styles":[{"id":1,"name":"normal"}]}]`))
		case "/audio_query":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"accent_phrases":[],"speedScale":1,"pitchScale":0,"intonationScale":1,"volumeScale":1,"prePhonemeLength":0,"postPhonemeLength":0,"outputSamplingRate":24000,"outputStereo":false,"kana":""}`))
		case "/synthesis":
			// Minimal WAV payload; the save path only writes the bytes.
			w.Header().Set("Content-Type", "audio/wav")
			_, _ = w.Write([]byte("RIFF\x24\x00\x00\x00WAVEfmt \x10\x00\x00\x00\x01\x00\x01\x00\x80\xbb\x00\x00\x00\x77\x01\x00\x02\x00\x10\x00data\x00\x00\x00\x00"))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	// Force the openai provider without a key so summarization fails
	// deterministically (fallback empty summary) and never touches the network.
	cfg := &config.Config{
		SaveAudioData:     true,
		AudioPath:         &audioDir,
		SummarizeProvider: config.SummarizeProviderOpenAI,
		VoiceVox: &config.VoiceVox{
			Speaker:  0,
			Style:    0,
			Endpoint: server.URL,
		},
	}
	t.Setenv("OPENAI_API_KEY", "")

	origEngine := tts.SpeachOpt.Engine
	tts.SpeachOpt.Engine = "voicevox"
	defer func() { tts.SpeachOpt.Engine = origEngine }()

	bookmarkFeed, err := client.Feed.Create().
		SetURL("https://quicknews.org/bookmark/rss").
		SetTitle("Bookmark").
		SetDescription("Bookmark").
		SetLink("https://quicknews.org/bookmark/rss").
		SetUpdatedAt(time.Now()).
		SetIsBookmark(true).
		Save(context.Background())
	require.NoError(t, err)

	repo := &RepositoryImpl{client: client, config: cfg}
	prepared, err := repo.prepareNewBookmark(context.Background(), server.URL+"/article", bookmarkFeed)
	require.NoError(t, err)
	require.NotNil(t, prepared.audioFile, "audio must be generated when save_audio_data is enabled")
	assert.FileExists(t, filepath.Join(audioDir, *prepared.audioFile))
}
