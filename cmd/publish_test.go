package cmd

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mopemope/quicknews/config"
	"github.com/mopemope/quicknews/ent"
	"github.com/mopemope/quicknews/models/article"
	"github.com/mopemope/quicknews/models/summary"
	"github.com/mopemope/quicknews/rss"
	"github.com/stretchr/testify/require"
)

type fakePublishArticleRepository struct {
	articles ent.Articles
}

func (r *fakePublishArticleRepository) GetById(context.Context, uuid.UUID) (*ent.Article, error) {
	return nil, nil
}

func (r *fakePublishArticleRepository) GetByFeed(context.Context, uuid.UUID) (ent.Articles, error) {
	return nil, nil
}

func (r *fakePublishArticleRepository) GetByUnreaded(context.Context, uuid.UUID) (ent.Articles, error) {
	return nil, nil
}

func (r *fakePublishArticleRepository) GetFromURL(context.Context, string) (*ent.Article, error) {
	return nil, nil
}

func (r *fakePublishArticleRepository) GetByDate(context.Context, uuid.UUID, string) (ent.Articles, error) {
	return r.articles, nil
}

func (r *fakePublishArticleRepository) Search(context.Context, article.SearchOptions) (ent.Articles, error) {
	return nil, nil
}

func (r *fakePublishArticleRepository) Save(context.Context, *ent.Article) (*ent.Article, error) {
	return nil, nil
}

func (r *fakePublishArticleRepository) SaveAll(context.Context, ent.Articles) error {
	return nil
}

func (r *fakePublishArticleRepository) Delete(context.Context, string) error {
	return nil
}

type fakePublishSummaryRepository struct {
	updatedAudio []string
	updateErr    error
}

func (r *fakePublishSummaryRepository) GetAll(context.Context) ([]*ent.Summary, error) {
	return nil, nil
}

func (r *fakePublishSummaryRepository) GetFromURL(context.Context, string) (*ent.Summary, error) {
	return nil, nil
}

func (r *fakePublishSummaryRepository) Save(context.Context, *ent.Summary) (*ent.Summary, error) {
	return nil, nil
}

func (r *fakePublishSummaryRepository) GetUnlistened(context.Context, *string) ([]*ent.Summary, error) {
	return nil, nil
}

func (r *fakePublishSummaryRepository) UpdateListened(context.Context, *ent.Summary) error {
	return nil
}

func (r *fakePublishSummaryRepository) UpdateReaded(context.Context, *ent.Summary) error {
	return nil
}

func (r *fakePublishSummaryRepository) UpdateAudioFile(_ context.Context, _ uuid.UUID, filename string) error {
	if r.updateErr != nil {
		return r.updateErr
	}
	r.updatedAudio = append(r.updatedAudio, filename)
	return nil
}

func (r *fakePublishSummaryRepository) Delete(context.Context, uuid.UUID) error {
	return nil
}

type fakeObjectStorage struct {
	uploads []string
}

func (s *fakeObjectStorage) Upload(_ context.Context, key string, _ io.Reader, _ string) error {
	s.uploads = append(s.uploads, key)
	return nil
}

var _ article.ArticleRepository = (*fakePublishArticleRepository)(nil)
var _ summary.SummaryRepository = (*fakePublishSummaryRepository)(nil)

func TestResolvePublishDates_DefaultBaseDate(t *testing.T) {
	now := time.Date(2026, 4, 22, 8, 0, 0, 0, time.UTC)

	dates, err := resolvePublishDates("", 3, now)
	require.NoError(t, err)
	require.Equal(t, []string{"2026-04-22", "2026-04-21", "2026-04-20"}, dates)
}

func TestResolvePublishDates_CustomDate(t *testing.T) {
	now := time.Date(2026, 4, 22, 8, 0, 0, 0, time.UTC)

	dates, err := resolvePublishDates("2026-01-05", 2, now)
	require.NoError(t, err)
	require.Equal(t, []string{"2026-01-05", "2026-01-04"}, dates)
}

func TestResolvePublishDates_InvalidRange(t *testing.T) {
	now := time.Date(2026, 4, 22, 8, 0, 0, 0, time.UTC)

	_, err := resolvePublishDates("", 0, now)
	require.Error(t, err)
}

func TestResolvePublishDates_InvalidDate(t *testing.T) {
	now := time.Date(2026, 4, 22, 8, 0, 0, 0, time.UTC)

	_, err := resolvePublishDates("2026-99-99", 1, now)
	require.Error(t, err)
}

func TestPublishCmd_RunRequiresNonEmptyAudioPath(t *testing.T) {
	audioPath := ""
	cmd := &PublishCmd{}
	cfg := &config.Config{
		AudioPath: &audioPath,
		Podcast:   &config.Podcast{},
	}

	err := cmd.Run(nil, cfg)
	require.Error(t, err)
	require.ErrorContains(t, err, "AudioPath")
}

func TestProcessFeed_SkipsWhenNoAudioFiles(t *testing.T) {
	audioDir := t.TempDir()
	cfg := testPublishConfig(audioDir)
	storage := &fakeObjectStorage{}

	pb := &publisher{
		ArticleRepository: &fakePublishArticleRepository{
			articles: ent.Articles{
				{
					Edges: ent.ArticleEdges{},
				},
			},
		},
		SummaryRepository: &fakePublishSummaryRepository{},
		RSSFeed:           rss.NewRSS(cfg.Podcast),
		R2Client:          storage,
		Config:            cfg,
		mergeAudio: func(string, []string) error {
			return nil
		},
		saveAudioData: summary.SaveAudioData,
	}

	err := pb.processFeed(context.Background(), &ent.Feed{ID: uuid.New(), Title: "Tech"}, "2026-04-22")
	require.NoError(t, err)
	require.Empty(t, storage.uploads)
	require.Empty(t, pb.RSSFeed.Channel.Items)
}

func TestProcessFeed_UploadsMergedAudioAndAddsRSSItem(t *testing.T) {
	audioDir := t.TempDir()
	cfg := testPublishConfig(audioDir)
	storage := &fakeObjectStorage{}
	mergeCalled := false

	audioFile := filepath.Join(audioDir, "summary.mp3")
	require.NoError(t, os.WriteFile(audioFile, []byte("input-mp3"), 0o644))

	sum := &ent.Summary{
		ID:        uuid.New(),
		Title:     "Summary title",
		Summary:   "Summary body",
		AudioFile: "summary.mp3",
	}
	articleItem := &ent.Article{
		Edges: ent.ArticleEdges{
			Summary: sum,
		},
	}

	pb := &publisher{
		ArticleRepository: &fakePublishArticleRepository{
			articles: ent.Articles{articleItem},
		},
		SummaryRepository: &fakePublishSummaryRepository{},
		RSSFeed:           rss.NewRSS(cfg.Podcast),
		R2Client:          storage,
		Config:            cfg,
		mergeAudio: func(out string, in []string) error {
			mergeCalled = true
			require.Len(t, in, 1)
			return os.WriteFile(out, []byte("merged-mp3"), 0o644)
		},
		saveAudioData: summary.SaveAudioData,
	}

	err := pb.processFeed(context.Background(), &ent.Feed{ID: uuid.New(), Title: "Tech"}, "2026-04-22")
	require.NoError(t, err)
	require.True(t, mergeCalled)
	require.Equal(t, []string{"2026-04-22_Tech.mp3"}, storage.uploads)
	require.Len(t, pb.RSSFeed.Channel.Items, 1)
	require.Equal(t, "2026-04-22 Tech Podcast", pb.RSSFeed.Channel.Items[0].Title)
}

func TestProcessFeed_SaveAudioErrorIncludesContext(t *testing.T) {
	audioDir := t.TempDir()
	cfg := testPublishConfig(audioDir)
	sum := &ent.Summary{
		ID:      uuid.New(),
		Title:   "Summary title",
		Summary: "Summary body",
	}
	articleItem := &ent.Article{
		Edges: ent.ArticleEdges{
			Summary: sum,
		},
	}

	pb := &publisher{
		ArticleRepository: &fakePublishArticleRepository{
			articles: ent.Articles{articleItem},
		},
		SummaryRepository: &fakePublishSummaryRepository{},
		RSSFeed:           rss.NewRSS(cfg.Podcast),
		R2Client:          &fakeObjectStorage{},
		Config:            cfg,
		mergeAudio: func(string, []string) error {
			return nil
		},
		saveAudioData: func(context.Context, *ent.Summary, *config.Config) (*string, error) {
			return nil, errors.New("tts unavailable")
		},
	}

	err := pb.processFeed(context.Background(), &ent.Feed{ID: uuid.New(), Title: "Tech"}, "2026-04-22")
	require.Error(t, err)
	require.ErrorContains(t, err, "failed to save audio data for summary")
	require.ErrorContains(t, err, "Summary title")
	require.ErrorContains(t, err, "Tech")
	require.ErrorContains(t, err, "2026-04-22")
}

func TestProcessFeed_SkipsWhenSaveAudioReturnsNoFilename(t *testing.T) {
	audioDir := t.TempDir()
	cfg := testPublishConfig(audioDir)
	storage := &fakeObjectStorage{}
	sum := &ent.Summary{
		ID:      uuid.New(),
		Title:   "Summary title",
		Summary: "Summary body",
	}
	articleItem := &ent.Article{
		Edges: ent.ArticleEdges{
			Summary: sum,
		},
	}

	pb := &publisher{
		ArticleRepository: &fakePublishArticleRepository{
			articles: ent.Articles{articleItem},
		},
		SummaryRepository: &fakePublishSummaryRepository{},
		RSSFeed:           rss.NewRSS(cfg.Podcast),
		R2Client:          storage,
		Config:            cfg,
		mergeAudio: func(string, []string) error {
			t.Fatal("mergeAudio should not be called without input audio files")
			return nil
		},
		saveAudioData: func(context.Context, *ent.Summary, *config.Config) (*string, error) {
			return nil, nil
		},
	}

	err := pb.processFeed(context.Background(), &ent.Feed{ID: uuid.New(), Title: "Tech"}, "2026-04-22")
	require.NoError(t, err)
	require.Empty(t, storage.uploads)
	require.Empty(t, pb.RSSFeed.Channel.Items)
}

func TestProcessFeed_UpdateAudioErrorIncludesContext(t *testing.T) {
	audioDir := t.TempDir()
	cfg := testPublishConfig(audioDir)
	sum := &ent.Summary{
		ID:      uuid.New(),
		Title:   "Summary title",
		Summary: "Summary body",
	}
	articleItem := &ent.Article{
		Edges: ent.ArticleEdges{
			Summary: sum,
		},
	}

	pb := &publisher{
		ArticleRepository: &fakePublishArticleRepository{
			articles: ent.Articles{articleItem},
		},
		SummaryRepository: &fakePublishSummaryRepository{
			updateErr: errors.New("db write failed"),
		},
		RSSFeed:  rss.NewRSS(cfg.Podcast),
		R2Client: &fakeObjectStorage{},
		Config:   cfg,
		mergeAudio: func(string, []string) error {
			return nil
		},
		saveAudioData: func(context.Context, *ent.Summary, *config.Config) (*string, error) {
			filename := "generated.mp3"
			return &filename, nil
		},
	}

	err := pb.processFeed(context.Background(), &ent.Feed{ID: uuid.New(), Title: "Tech"}, "2026-04-22")
	require.Error(t, err)
	require.ErrorContains(t, err, "failed to update audio file for summary")
	require.ErrorContains(t, err, "Summary title")
	require.ErrorContains(t, err, "Tech")
	require.ErrorContains(t, err, "2026-04-22")
}

func testPublishConfig(audioDir string) *config.Config {
	return &config.Config{
		AudioPath: &audioDir,
		Podcast: &config.Podcast{
			ChannelTitle: "quicknews",
			ChannelLink:  "https://example.com/podcast",
			ChannelDesc:  "quicknews feed",
			Author:       "quicknews",
			PublishURL:   "https://example.com/podcast",
		},
	}
}
