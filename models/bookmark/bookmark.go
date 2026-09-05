package bookmark

import (
	"context"
	"log/slog"
	"sync"

	"github.com/cockroachdb/errors"
	"github.com/google/uuid"
	"github.com/mopemope/quicknews/clock"
	"github.com/mopemope/quicknews/config"
	"github.com/mopemope/quicknews/database"
	"github.com/mopemope/quicknews/ent"
	"github.com/mopemope/quicknews/ent/article"
	"github.com/mopemope/quicknews/ent/feed"
	"github.com/mopemope/quicknews/models/summary"
	"github.com/mopemope/quicknews/org"
	"github.com/mopemope/quicknews/scraper"
	"github.com/mopemope/quicknews/summarizer"
)

type Repository interface {
	AddBookmark(ctx context.Context, url string) error
	GetBookmarkFeed(ctx context.Context) (*ent.Feed, error)
}

type RepositoryImpl struct {
	client           *ent.Client
	config           *config.Config
	summarizerClient summarizer.Summarizer
	initMu           sync.Mutex
}

func NewRepository(ctx context.Context, client *ent.Client, cfg *config.Config) (Repository, error) {
	if client == nil {
		return nil, errors.New("ent client is required")
	}
	if cfg == nil {
		cfg = &config.Config{}
	}
	return &RepositoryImpl{
		client: client,
		config: cfg,
	}, nil
}

func (r *RepositoryImpl) GetBookmarkFeed(ctx context.Context) (*ent.Feed, error) {
	feed, err := r.client.Feed.
		Query().
		Where(feed.IsBookmarkEQ(true)).
		Only(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get bookmark feeds")
	}
	return feed, nil
}

func (r *RepositoryImpl) AddBookmark(ctx context.Context, url string) error {
	bookmarkFeed, err := r.GetBookmarkFeed(ctx)
	if err != nil {
		return err
	}

	existArticle, err := r.client.Article.
		Query().
		Where(article.URL(url)).
		WithFeed().
		WithSummary().
		Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return errors.Wrap(err, "failed to query existing article")
	}

	if existArticle != nil {
		// Existing article: only its feed association changes, no network I/O needed.
		return database.WithTx(ctx, r.client, func(tx *ent.Tx) error {
			return r.handleExistingArticle(ctx, tx, existArticle, bookmarkFeed)
		})
	}

	// New article: prepare page content and summary outside the transaction so
	// long-running network calls never hold the SQLite write lock.
	prepared, err := r.prepareNewBookmark(ctx, url, bookmarkFeed)
	if err != nil {
		return err
	}

	return database.WithTx(ctx, r.client, func(tx *ent.Tx) error {
		return r.saveNewBookmark(ctx, tx, prepared, bookmarkFeed)
	})
}

// preparedBookmark holds everything needed to persist a new bookmark.
type preparedBookmark struct {
	summaryID   uuid.UUID
	title       string
	pageSummary *summarizer.PageSummary
	audioFile   *string
}

// prepareNewBookmark performs all network I/O (page fetch, summarization,
// audio synthesis) before any transaction starts.
func (r *RepositoryImpl) prepareNewBookmark(ctx context.Context, url string, bookmarkFeed *ent.Feed) (*preparedBookmark, error) {
	page, err := scraper.GetPageContent(ctx, url)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get page content")
	}
	if page.Title == "" {
		return nil, errors.Wrap(&scraper.PermanentError{Err: errors.Newf("could not find title tag on %s", url)}, "failed to get title")
	}

	pageSummary, err := r.summarizeContent(ctx, page)
	if err != nil {
		// Log the error but proceed to create the summary entry without the AI summary
		slog.Error("failed to summarize page, creating summary entry without AI summary", slog.Any("url", url), slog.Any("error", err))
		pageSummary = &summarizer.PageSummary{
			URL:     url,
			Title:   page.Title, // Use scraped title as fallback
			Summary: "",         // Empty summary
		}
	}

	prepared := &preparedBookmark{
		summaryID:   uuid.New(),
		title:       page.Title,
		pageSummary: pageSummary,
	}

	if r.config.SaveAudioData {
		prepared.audioFile = r.prepareAudio(ctx, prepared, bookmarkFeed)
	}
	return prepared, nil
}

// prepareAudio synthesizes audio for the summary when it fits the limits.
// The summary ID is pre-generated so the audio filename matches the record
// that will be persisted inside the transaction. The feed edge is required
// by GetAudioData to build the TTS text. Failures are logged; the bookmark
// is still saved without audio.
func (r *RepositoryImpl) prepareAudio(ctx context.Context, prepared *preparedBookmark, bookmarkFeed *ent.Feed) *string {
	sum := &ent.Summary{
		ID:      prepared.summaryID,
		Title:   prepared.pageSummary.Title,
		Summary: prepared.pageSummary.Summary,
		Edges:   ent.SummaryEdges{Feed: bookmarkFeed},
	}
	if len(sum.Summary)+len(sum.Title) > summary.MaxAudioTextLength {
		slog.Warn("Skip summary because it is too long", slog.Any("title", sum.Title))
		return nil
	}
	filename, err := summary.SaveAudioData(ctx, sum, r.config)
	if err != nil {
		slog.Error("failed to save audio data", slog.Any("error", err))
		return nil
	}
	return filename
}

// saveNewBookmark persists the article, summary and audio reference inside a
// transaction. Network I/O must have been completed by prepareNewBookmark.
func (r *RepositoryImpl) saveNewBookmark(ctx context.Context, tx *ent.Tx, prepared *preparedBookmark, bookmarkFeed *ent.Feed) error {
	// Re-check inside the transaction: a concurrent AddBookmark may have
	// created the article after the outer check.
	existArticle, err := tx.Article.
		Query().
		Where(article.URL(prepared.pageSummary.URL)).
		WithFeed().
		WithSummary().
		Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return errors.Wrap(err, "failed to query existing article")
	}
	if existArticle != nil {
		return r.handleExistingArticle(ctx, tx, existArticle, bookmarkFeed)
	}

	now := clock.Now()
	createdArticle, err := tx.Article.Create().
		SetTitle(prepared.title).
		SetURL(prepared.pageSummary.URL).
		SetDescription("").
		SetContent("").
		SetCreatedAt(now).
		SetPublishedAt(now).
		SetFeed(bookmarkFeed).
		Save(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to create article")
	}
	createdArticle.Edges.Feed = bookmarkFeed

	createdSummary, err := tx.Summary.
		Create().
		SetID(prepared.summaryID).
		SetTitle(prepared.pageSummary.Title).
		SetSummary(prepared.pageSummary.Summary).
		SetURL(prepared.pageSummary.URL).
		SetNillableAudioFile(prepared.audioFile).
		SetCreatedAt(now).
		SetArticle(createdArticle).
		SetFeed(bookmarkFeed).
		Save(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to save summary")
	}
	createdSummary.Edges.Article = createdArticle
	createdSummary.Edges.Feed = bookmarkFeed

	if err := org.ExportOrg(r.config, createdSummary); err != nil {
		// Log the error but don't fail the transaction
		slog.Error("failed to export org", slog.Any("summary_id", createdSummary.ID), slog.Any("error", err))
	}
	return nil
}

// handleExistingArticle handles the case where the article already exists.
func (r *RepositoryImpl) handleExistingArticle(ctx context.Context, tx *ent.Tx, existArticle *ent.Article, bookmarkFeed *ent.Feed) error {
	if existArticle.Edges.Feed.IsBookmark {
		// already bookmarked
		slog.Warn("already bookmarked", slog.Any("url", existArticle.URL))
		return nil
	}

	// Update the article and summary to point to the bookmark feed
	if err := tx.Article.
		UpdateOneID(existArticle.ID).
		SetFeedID(bookmarkFeed.ID).
		Exec(ctx); err != nil {
		return errors.Wrap(err, "failed to update article")
	}

	if existArticle.Edges.Summary == nil {
		// Summary might not exist if it failed previously or was deleted
		slog.Warn("summary not found for existing article, skipping summary update", slog.Any("article_id", existArticle.ID))
		return nil
	}

	if err := tx.Summary.
		UpdateOneID(existArticle.Edges.Summary.ID).
		SetFeedID(bookmarkFeed.ID).
		Exec(ctx); err != nil {
		// Log the error but don't fail the whole transaction,
		// as the article itself was successfully moved.
		slog.Error("failed to update summary feed", slog.Any("summary_id", existArticle.Edges.Summary.ID), slog.Any("error", err))
	}
	return nil
}

// summarizeContent summarizes the pre-fetched page. Content-aware providers
// (OpenAI) reuse the page without another fetch; others (Gemini) fetch the
// URL themselves via search grounding.
func (r *RepositoryImpl) summarizeContent(ctx context.Context, page *scraper.PageContent) (*summarizer.PageSummary, error) {
	if err := r.ensureSummarizer(ctx); err != nil {
		return nil, errors.Wrap(err, "failed to initialize summarizer")
	}

	pageSummary, err := summarizer.SummarizeContentWithRetry(ctx, r.summarizerClient, page, summarizer.DefaultRetryWait)
	if err != nil {
		return nil, errors.Wrap(err, "failed to summarize page")
	}
	return pageSummary, nil
}

func (r *RepositoryImpl) ensureSummarizer(ctx context.Context) error {
	r.initMu.Lock()
	defer r.initMu.Unlock()

	if r.summarizerClient != nil {
		return nil
	}

	client, err := summarizer.New(ctx, r.config)
	if err != nil {
		return errors.Wrap(err, "failed to create summarizer client")
	}
	r.summarizerClient = client
	return nil
}
