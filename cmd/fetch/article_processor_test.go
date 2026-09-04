package fetch

import (
	"context"
	stderrors "errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mmcdole/gofeed"
	"github.com/mopemope/quicknews/config"
	"github.com/mopemope/quicknews/ent"
	"github.com/mopemope/quicknews/gemini"
	"github.com/mopemope/quicknews/models/article"
	"github.com/mopemope/quicknews/models/summary"
	"github.com/stretchr/testify/require"
)

type fakeArticleRepository struct {
	existing     *ent.Article
	savedArticle *ent.Article
	saveCalls    int
	getErr       error
	saveErr      error
}

func (r *fakeArticleRepository) GetById(context.Context, uuid.UUID) (*ent.Article, error) {
	return nil, nil
}

func (r *fakeArticleRepository) GetByFeed(context.Context, uuid.UUID) (ent.Articles, error) {
	return nil, nil
}

func (r *fakeArticleRepository) GetByUnreaded(context.Context, uuid.UUID) (ent.Articles, error) {
	return nil, nil
}

func (r *fakeArticleRepository) GetFromURL(context.Context, string) (*ent.Article, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	return r.existing, nil
}

func (r *fakeArticleRepository) GetByDate(context.Context, uuid.UUID, string) (ent.Articles, error) {
	return nil, nil
}

func (r *fakeArticleRepository) Search(context.Context, article.SearchOptions) (ent.Articles, error) {
	return nil, nil
}

func (r *fakeArticleRepository) Save(_ context.Context, article *ent.Article) (*ent.Article, error) {
	if r.saveErr != nil {
		return nil, r.saveErr
	}
	r.saveCalls++
	article.ID = uuid.New()
	r.savedArticle = article
	return article, nil
}

func (r *fakeArticleRepository) SaveAll(context.Context, ent.Articles) error {
	return nil
}

func (r *fakeArticleRepository) Delete(context.Context, string) error {
	return nil
}

type fakeSummaryRepository struct {
	savedSummary *ent.Summary
	saveCalls    int
}

func (r *fakeSummaryRepository) GetAll(context.Context) ([]*ent.Summary, error) {
	return nil, nil
}

func (r *fakeSummaryRepository) GetFromURL(context.Context, string) (*ent.Summary, error) {
	return nil, nil
}

func (r *fakeSummaryRepository) Save(_ context.Context, sum *ent.Summary) (*ent.Summary, error) {
	r.saveCalls++
	sum.ID = uuid.New()
	r.savedSummary = sum
	return sum, nil
}

func (r *fakeSummaryRepository) GetUnlistened(context.Context, *string) ([]*ent.Summary, error) {
	return nil, nil
}

func (r *fakeSummaryRepository) UpdateListened(context.Context, *ent.Summary) error {
	return nil
}

func (r *fakeSummaryRepository) UpdateReaded(context.Context, *ent.Summary) error {
	return nil
}

func (r *fakeSummaryRepository) SetReaded(context.Context, uuid.UUID, bool) error {
	return nil
}

func (r *fakeSummaryRepository) MarkFeedReaded(context.Context, uuid.UUID) (int, error) {
	return 0, nil
}

func (r *fakeSummaryRepository) UpdateAudioFile(context.Context, uuid.UUID, string) error {
	return nil
}

func (r *fakeSummaryRepository) Delete(context.Context, uuid.UUID) error {
	return nil
}

type fakeSummarizer struct {
	result     *gemini.PageSummary
	calls      int
	closeCalls int
}

func (s *fakeSummarizer) Summarize(context.Context, string) (*gemini.PageSummary, error) {
	s.calls++
	return s.result, nil
}

func (s *fakeSummarizer) Close() error {
	s.closeCalls++
	return nil
}

var _ article.ArticleRepository = (*fakeArticleRepository)(nil)
var _ summary.SummaryRepository = (*fakeSummaryRepository)(nil)
var _ gemini.Summarizer = (*fakeSummarizer)(nil)

func TestArticleProcessor_Process_SavesArticleAndSummary(t *testing.T) {
	feedEntity := &ent.Feed{ID: uuid.New(), Title: "Tech", URL: "https://example.com/feed"}
	item := &gofeed.Item{
		Title:       "Article title",
		Link:        "https://example.com/article",
		Description: "Description",
		Content:     "Content",
	}
	articleRepo := &fakeArticleRepository{}
	summaryRepo := &fakeSummaryRepository{}
	summarizer := &fakeSummarizer{
		result: &gemini.PageSummary{
			Title:   "Summary title",
			Summary: "Summary body",
		},
	}

	processor := NewArticleProcessorWithSummarizer(
		feedEntity,
		item,
		articleRepo,
		summaryRepo,
		&config.Config{},
		func(context.Context, *config.Config) (gemini.Summarizer, error) {
			return summarizer, nil
		},
	)

	err := processor.Process(context.Background())
	require.NoError(t, err)
	require.Equal(t, 1, articleRepo.saveCalls)
	require.Equal(t, 1, summaryRepo.saveCalls)
	require.Equal(t, 1, summarizer.calls)
	require.Equal(t, 1, summarizer.closeCalls)
	require.NotNil(t, articleRepo.savedArticle)
	require.Equal(t, feedEntity, articleRepo.savedArticle.Edges.Feed)
	require.NotNil(t, summaryRepo.savedSummary)
	require.Equal(t, "Summary title", summaryRepo.savedSummary.Title)
	require.Equal(t, articleRepo.savedArticle, summaryRepo.savedSummary.Edges.Article)
	require.Equal(t, feedEntity, summaryRepo.savedSummary.Edges.Feed)
}

func TestArticleProcessor_Process_SkipsSummarizerWhenSummaryExists(t *testing.T) {
	feedEntity := &ent.Feed{ID: uuid.New(), Title: "Tech", URL: "https://example.com/feed"}
	item := &gofeed.Item{
		Title: "Article title",
		Link:  "https://example.com/article",
	}
	articleRepo := &fakeArticleRepository{
		existing: &ent.Article{
			ID: uuid.New(),
			Edges: ent.ArticleEdges{
				Summary: &ent.Summary{ID: uuid.New()},
			},
		},
	}
	summaryRepo := &fakeSummaryRepository{}
	summarizer := &fakeSummarizer{
		result: &gemini.PageSummary{
			Title:   "Summary title",
			Summary: "Summary body",
		},
	}

	processor := NewArticleProcessorWithSummarizer(
		feedEntity,
		item,
		articleRepo,
		summaryRepo,
		&config.Config{},
		func(context.Context, *config.Config) (gemini.Summarizer, error) {
			return summarizer, nil
		},
	)

	err := processor.Process(context.Background())
	require.NoError(t, err)
	require.Equal(t, 0, articleRepo.saveCalls)
	require.Equal(t, 0, summaryRepo.saveCalls)
	require.Equal(t, 0, summarizer.calls)
	require.Equal(t, 0, summarizer.closeCalls)
}

func TestArticleProcessor_Process_ErrorsWhenSummarizerReturnsNil(t *testing.T) {
	feedEntity := &ent.Feed{ID: uuid.New(), Title: "Tech", URL: "https://example.com/feed"}
	item := &gofeed.Item{
		Title: "Article title",
		Link:  "https://example.com/article",
	}
	articleRepo := &fakeArticleRepository{}
	summaryRepo := &fakeSummaryRepository{}
	summarizer := &fakeSummarizer{}

	processor := NewArticleProcessorWithSummarizer(
		feedEntity,
		item,
		articleRepo,
		summaryRepo,
		&config.Config{},
		func(context.Context, *config.Config) (gemini.Summarizer, error) {
			return summarizer, nil
		},
	)
	processor.retryWait = func(context.Context, time.Duration) error {
		return nil
	}

	err := processor.Process(context.Background())
	require.Error(t, err)
	require.ErrorContains(t, err, "summarizer returned nil summary")
	require.Equal(t, 3, summarizer.calls)
	require.Equal(t, 0, summaryRepo.saveCalls)
}

func TestQueueItemWrapper_Process_ReturnsArticleError(t *testing.T) {
	feedEntity := &ent.Feed{ID: uuid.New(), Title: "Tech", URL: "https://example.com/feed"}
	item := &gofeed.Item{
		Title: "Article title",
		Link:  "https://example.com/article",
	}
	processor := NewArticleProcessor(
		feedEntity,
		item,
		&fakeArticleRepository{getErr: stderrors.New("database unavailable")},
		&fakeSummaryRepository{},
		&config.Config{},
	)
	wrapper := &QueueItemWrapper{processor: processor, name: item.Title}

	err := wrapper.Process()
	require.Error(t, err)
	require.ErrorContains(t, err, "failed to process article")
	require.ErrorContains(t, err, "database unavailable")
}
