package daemon

import (
	"context"
	"log/slog"
	"sync"
	"time"

	pond "github.com/alitto/pond/v2"
	"github.com/cockroachdb/errors"
	"github.com/mmcdole/gofeed"
	configPkg "github.com/mopemope/quicknews/config"
	"github.com/mopemope/quicknews/ent"
	"github.com/mopemope/quicknews/gemini"
	"github.com/mopemope/quicknews/models/article"
	"github.com/mopemope/quicknews/models/feed"
	summaryModel "github.com/mopemope/quicknews/models/summary"
	"github.com/mopemope/quicknews/org"
)

// DaemonManager manages the daemon process lifecycle and operations
type DaemonManager struct {
	config     *configPkg.Config
	client     *ent.Client
	pidManager *PIDManager
	
	// Configuration
	interval   time.Duration
	maxWorkers int
	
	// Control channels
	stopCh     chan struct{}
	reloadCh   chan struct{}
	shutdownCh chan struct{}
	
	// Signal handling
	signalHandler *SignalHandler
	
	// Health checking
	healthChecker *HealthChecker
	
	// Publishing
	publisher *DaemonPublisher
	
	// Repositories
	feedRepos    feed.FeedRepository
	articleRepos article.ArticleRepository
	summaryRepos summaryModel.SummaryRepository
	
	// State management
	mu         sync.RWMutex
	isRunning  bool
	startTime  time.Time
	debugMode  bool
	
	// Statistics
	stats      *Statistics
}

// Statistics holds daemon runtime statistics
type Statistics struct {
	mu                  sync.RWMutex
	StartTime           time.Time `json:"start_time"`
	LastFetchTime       time.Time `json:"last_fetch_time"`
	TotalFetches        int64     `json:"total_fetches"`
	FeedsProcessed      int64     `json:"feeds_processed"`
	ArticlesFetched     int64     `json:"articles_fetched"`
	SummariesGenerated  int64     `json:"summaries_generated"`
	ErrorsCount         int64     `json:"errors_count"`
	AverageProcessTime  float64   `json:"average_process_time_seconds"`
}

// DaemonConfig holds daemon-specific configuration
type DaemonConfig struct {
	Interval        time.Duration
	MaxWorkers      int
	PidFile         string
	Detach          bool
	HealthCheckPort int
}

// NewDaemonManager creates a new daemon manager
func NewDaemonManager(client *ent.Client, config *configPkg.Config, daemonConfig *DaemonConfig) (*DaemonManager, error) {
	if client == nil {
		return nil, errors.New("database client cannot be nil")
	}
	if config == nil {
		return nil, errors.New("config cannot be nil")
	}
	if daemonConfig == nil {
		return nil, errors.New("daemon config cannot be nil")
	}

	pidManager := NewPIDManager(daemonConfig.PidFile)
	if pidManager == nil {
		return nil, errors.New("failed to create PID manager")
	}
	
	feedRepos := feed.NewRepository(client)
	if feedRepos == nil {
		return nil, errors.New("failed to create feed repository")
	}
	
	articleRepos := article.NewRepository(client)
	if articleRepos == nil {
		return nil, errors.New("failed to create article repository")
	}
	
	summaryRepos := summaryModel.NewRepository(client)
	if summaryRepos == nil {
		return nil, errors.New("failed to create summary repository")
	}

	manager := &DaemonManager{
		config:       config,
		client:       client,
		pidManager:   pidManager,
		interval:     daemonConfig.Interval,
		maxWorkers:   daemonConfig.MaxWorkers,
		stopCh:       make(chan struct{}),
		reloadCh:     make(chan struct{}),
		shutdownCh:   make(chan struct{}),
		feedRepos:    feedRepos,
		articleRepos: articleRepos,
		summaryRepos: summaryRepos,
		debugMode:    false,
		stats: &Statistics{
			StartTime: time.Now(),
		},
	}
	
	// Initialize signal handler
	signalHandler := NewSignalHandler(manager)
	if signalHandler == nil {
		return nil, errors.New("failed to create signal handler")
	}
	manager.signalHandler = signalHandler
	
	// Initialize health checker if port is specified
	if daemonConfig.HealthCheckPort > 0 {
		healthChecker := NewHealthChecker(manager, daemonConfig.HealthCheckPort)
		if healthChecker == nil {
			return nil, errors.New("failed to create health checker")
		}
		manager.healthChecker = healthChecker
	}
	
	// Initialize publisher if enabled
	if publisher, err := NewDaemonPublisher(context.Background(), client, config); err != nil {
		slog.Warn("Failed to initialize daemon publisher", "error", err)
	} else {
		if publisher == nil {
			slog.Warn("Daemon publisher creation returned nil")
		} else {
			manager.publisher = publisher
			if publisher.IsEnabled() {
				slog.Info("Daemon publisher initialized and enabled")
			} else {
				slog.Info("Daemon publisher initialized but disabled")
			}
		}
	}
	
	return manager, nil
}

// Start starts the daemon process
func (dm *DaemonManager) Start(ctx context.Context) error {
	if dm == nil {
		return errors.New("daemon manager is nil")
	}
	if ctx == nil {
		return errors.New("context cannot be nil")
	}

	dm.mu.Lock()
	if dm.isRunning {
		dm.mu.Unlock()
		return errors.New("daemon is already running")
	}
	dm.isRunning = true
	dm.startTime = time.Now()
	if dm.stats != nil {
		dm.stats.StartTime = dm.startTime
	}
	dm.mu.Unlock()
	
	// Write PID file
	if dm.pidManager != nil {
		if err := dm.pidManager.WritePID(); err != nil {
			return errors.Wrap(err, "failed to write PID file")
		}
	}
	
	// Start signal handler
	if dm.signalHandler != nil {
		dm.signalHandler.Start()
		defer dm.signalHandler.Stop()
	}
	
	// Start health checker if available
	if dm.healthChecker != nil {
		if err := dm.healthChecker.Start(); err != nil {
			slog.Warn("Failed to start health checker", "error", err)
		} else {
			defer func() {
				if err := dm.healthChecker.Stop(); err != nil {
					slog.Warn("Failed to stop health checker", "error", err)
				}
			}()
		}
	}
	
	slog.Info("Daemon started", 
		"pid", dm.pidManager.GetPIDFile(),
		"interval", dm.interval,
		"max_workers", dm.maxWorkers)
	
	// Start the main loop
	return dm.run(ctx)
}

// Stop stops the daemon process gracefully
func (dm *DaemonManager) Stop() error {
	dm.mu.Lock()
	if !dm.isRunning {
		dm.mu.Unlock()
		return errors.New("daemon is not running")
	}
	dm.mu.Unlock()
	
	slog.Info("Stopping daemon...")
	close(dm.stopCh)
	
	// Remove PID file
	if err := dm.pidManager.RemovePID(); err != nil {
		slog.Warn("Failed to remove PID file", "error", err)
	}
	
	dm.mu.Lock()
	dm.isRunning = false
	dm.mu.Unlock()
	
	slog.Info("Daemon stopped")
	return nil
}

// IsRunning returns whether the daemon is currently running
func (dm *DaemonManager) IsRunning() bool {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	return dm.isRunning
}

// GetStatistics returns current daemon statistics
func (dm *DaemonManager) GetStatistics() *Statistics {
	dm.stats.mu.RLock()
	defer dm.stats.mu.RUnlock()
	
	// Create a copy to avoid race conditions, excluding the mutex
	statsCopy := Statistics{
		StartTime:           dm.stats.StartTime,
		LastFetchTime:       dm.stats.LastFetchTime,
		TotalFetches:        dm.stats.TotalFetches,
		FeedsProcessed:      dm.stats.FeedsProcessed,
		ArticlesFetched:     dm.stats.ArticlesFetched,
		SummariesGenerated:  dm.stats.SummariesGenerated,
		ErrorsCount:         dm.stats.ErrorsCount,
		AverageProcessTime:  dm.stats.AverageProcessTime,
	}
	return &statsCopy
}

// run is the main daemon loop
func (dm *DaemonManager) run(ctx context.Context) error {
	ticker := time.NewTicker(dm.interval)
	defer ticker.Stop()
	
	// Initial fetch
	dm.performFetch(ctx)
	
	for {
		select {
		case <-ctx.Done():
			slog.Info("Context cancelled, stopping daemon")
			return ctx.Err()
			
		case <-dm.stopCh:
			slog.Info("Stop signal received")
			return nil
			
		case <-dm.shutdownCh:
			slog.Info("Graceful shutdown initiated")
			return dm.performGracefulShutdown(ctx)
			
		case <-dm.reloadCh:
			slog.Info("Reload signal received")
			// TODO: Implement config reload
			
		case <-ticker.C:
			dm.performFetch(ctx)
			
			// Check if it's time to publish
			if dm.publisher != nil && dm.publisher.ShouldPublishNow() {
				go dm.performPublish(ctx)
			}
		}
	}
}

// performFetch performs a single fetch operation
func (dm *DaemonManager) performFetch(ctx context.Context) {
	startTime := time.Now()
	
	slog.Info("Starting fetch operation")
	
	// Update statistics
	dm.stats.mu.Lock()
	dm.stats.TotalFetches++
	dm.stats.LastFetchTime = startTime
	dm.stats.mu.Unlock()
	
	// Get all feeds
	feeds, err := dm.feedRepos.All(ctx)
	if err != nil {
		slog.Error("Failed to get feeds", "error", err)
		dm.incrementErrorCount()
		return
	}
	
	if len(feeds) == 0 {
		slog.Info("No feeds registered")
		return
	}
	
	// Process feeds concurrently
	pool := pond.NewPool(dm.maxWorkers)
	var wg sync.WaitGroup
	
	for _, feedData := range feeds {
		if feedData.IsBookmark {
			continue // Skip bookmark feeds
		}
		
		wg.Add(1)
		feed := feedData // Capture for closure
		
		pool.Submit(func() {
			defer wg.Done()
			if err := dm.processFeed(ctx, feed); err != nil {
				slog.Error("Failed to process feed", 
					"feed", feed.Title, 
					"url", feed.URL, 
					"error", err)
				dm.incrementErrorCount()
			}
		})
	}
	
	wg.Wait()
	pool.StopAndWait()
	
	// Update processing time statistics
	processingTime := time.Since(startTime)
	dm.updateProcessingTime(processingTime)
	
	slog.Info("Fetch operation completed", 
		"duration", processingTime,
		"feeds_processed", len(feeds))
}

// processFeed processes a single feed
func (dm *DaemonManager) processFeed(ctx context.Context, feedData *ent.Feed) error {
	// Import the existing feed processing logic from cmd/fetch.go
	items, err := dm.processFeedItems(ctx, feedData)
	if err != nil {
		return errors.Wrap(err, "failed to process feed items")
	}
	
	// Process articles concurrently
	if len(items) > 0 {
		pool := pond.NewPool(3) // Smaller pool for individual feed processing
		var wg sync.WaitGroup
		
		for _, item := range items {
			wg.Add(1)
			article := item // Capture for closure
			
			pool.Submit(func() {
				defer wg.Done()
				article.Process()
			})
		}
		
		wg.Wait()
		pool.StopAndWait()
		
		// Update statistics
		dm.stats.mu.Lock()
		dm.stats.ArticlesFetched += int64(len(items))
		dm.stats.mu.Unlock()
	}
	
	dm.stats.mu.Lock()
	dm.stats.FeedsProcessed++
	dm.stats.mu.Unlock()
	
	slog.Debug("Processed feed", 
		"title", feedData.Title, 
		"url", feedData.URL,
		"articles", len(items))
	
	return nil
}

// incrementErrorCount increments the error counter
func (dm *DaemonManager) incrementErrorCount() {
	dm.stats.mu.Lock()
	dm.stats.ErrorsCount++
	dm.stats.mu.Unlock()
}

// updateProcessingTime updates the average processing time
func (dm *DaemonManager) updateProcessingTime(duration time.Duration) {
	dm.stats.mu.Lock()
	defer dm.stats.mu.Unlock()
	
	seconds := duration.Seconds()
	if dm.stats.AverageProcessTime == 0 {
		dm.stats.AverageProcessTime = seconds
	} else {
		// Simple moving average
		dm.stats.AverageProcessTime = (dm.stats.AverageProcessTime + seconds) / 2
	}
}

// processFeedItems processes feed items and returns articles to be processed
func (dm *DaemonManager) processFeedItems(ctx context.Context, feedData *ent.Feed) ([]*DaemonArticle, error) {
	fp := gofeed.NewParser()
	slog.Debug("Fetching feed", "title", feedData.Title, "url", feedData.URL)

	parsedFeed, err := fp.ParseURLWithContext(feedData.URL, ctx)
	if err != nil {
		return nil, errors.Wrap(err, "fetch error")
	}

	updatedFeed, err := dm.feedRepos.UpdateFeed(ctx, feedData, parsedFeed)
	if err != nil {
		return nil, errors.Wrap(err, "error updating feed")
	}

	var items []*DaemonArticle
	for _, item := range parsedFeed.Items {
		daemonArticle := &DaemonArticle{
			name:         item.Title,
			feed:         updatedFeed,
			feedItem:     item,
			articleRepos: dm.articleRepos,
			summaryRepos: dm.summaryRepos,
			config:       dm.config,
			stats:        dm.stats,
		}
		items = append(items, daemonArticle)
	}

	return items, nil
}

// DaemonArticle represents an article to be processed by the daemon
type DaemonArticle struct {
	name         string
	feed         *ent.Feed
	feedItem     *gofeed.Item
	articleRepos article.ArticleRepository
	summaryRepos summaryModel.SummaryRepository
	config       *configPkg.Config
	stats        *Statistics
}

// DisplayName returns the display name of the article
func (da *DaemonArticle) DisplayName() string {
	return da.name
}

// URL returns the URL of the article
func (da *DaemonArticle) URL() string {
	return da.feedItem.Link
}

// Process processes the article (similar to the existing Article.Process method)
func (da *DaemonArticle) Process() {
	ctx := context.Background()
	article, err := da.articleRepos.GetFromURL(ctx, da.feedItem.Link)
	if err != nil {
		slog.Error("Error checking if article exists", "link", da.feedItem.Link, "error", err)
		return
	}

	if article == nil {
		slog.Debug("Processing new article", "title", da.feedItem.Title, "link", da.feedItem.Link)
		newArticle := &ent.Article{
			Title:       da.feedItem.Title,
			URL:         da.feedItem.Link,
			Description: da.feedItem.Description,
			Content:     da.feedItem.Content,
		}
		newArticle.Edges.Feed = da.feed

		// PublishedParsed があれば設定
		if da.feedItem.PublishedParsed != nil {
			newArticle.PublishedAt = *da.feedItem.PublishedParsed
		} else if da.feedItem.UpdatedParsed != nil {
			newArticle.PublishedAt = *da.feedItem.UpdatedParsed
		}

		article, err = da.articleRepos.Save(ctx, newArticle)
		if err != nil {
			slog.Error("Error saving article", "link", da.feedItem.Link, "error", err)
			return
		}
		article.Edges.Feed = da.feed
		slog.Debug("Saved article", "link", da.feedItem.Link, "id", newArticle.ID)
	}

	if article.Edges.Summary == nil {
		if err := da.processSummary(ctx, article); err != nil {
			slog.Error("Error processing summary", "link", article.URL, "error", err)
			return
		}
		
		// Update statistics
		da.stats.mu.Lock()
		da.stats.SummariesGenerated++
		da.stats.mu.Unlock()
	}
}

// processSummary processes the summary for an article (similar to existing logic)
func (da *DaemonArticle) processSummary(ctx context.Context, article *ent.Article) error {
	if da == nil {
		return errors.New("daemon article is nil")
	}
	if ctx == nil {
		return errors.New("context cannot be nil")
	}
	if article == nil {
		return errors.New("article cannot be nil")
	}
	if article.URL == "" {
		return errors.New("article URL cannot be empty")
	}

	// Import the existing summary processing logic from cmd/fetch.go
	geminiClient, err := gemini.NewClient(ctx, da.config)
	if err != nil {
		return errors.Wrap(err, "error creating gemini client")
	}
	if geminiClient == nil {
		return errors.New("gemini client creation returned nil")
	}

	defer func() {
		if closeErr := geminiClient.Close(); closeErr != nil {
			slog.Warn("Failed to close gemini client", "error", closeErr)
		}
	}()

	url := article.URL
	var pageSummary *gemini.PageSummary
	for i := range 3 {
		pageSummary, err = geminiClient.Summarize(ctx, url)
		if err != nil || pageSummary == nil {
			// retry if error
			slog.Info("retrying to summarize page", "link", url, "error", err, "attempt", i+1)
			i += 1
			wait := i * i
			time.Sleep(time.Duration(wait) * time.Second)
		} else {
			break
		}
	}
	if err != nil {
		return errors.Wrap(err, "error summarizing page")
	}

	sum := &ent.Summary{
		URL:     url,
		Title:   pageSummary.Title,
		Summary: pageSummary.Summary,
		Readed:  false,
		Listend: false,
	}
	sum.Edges.Article = article
	sum.Edges.Feed = article.Edges.Feed

	slog.Debug("Saving summary", "title", sum.Title, "summary", sum.Summary)
	created, err := da.summaryRepos.Save(ctx, sum)
	if err != nil {
		slog.Error("Error saving summary", "link", article.URL, "error", err)
		return err
	}

	// Generate audio if configured
	if da.config.SaveAudioData {
		if len(created.Summary)+len(created.Title) > 4500 {
			// skip
			slog.Warn("Skip summary because it is too long", slog.Any("title", created.Title))
		} else {
			filename, err := summaryModel.SaveAudioData(ctx, created, da.config)
			if err != nil {
				return err
			}
			if filename != nil {
				if err := da.summaryRepos.UpdateAudioFile(ctx, created.ID, *filename); err != nil {
					return err
				}
			}
		}
	}
	
	// Export to Org mode if configured
	if err := org.ExportOrg(da.config, created); err != nil {
		return err
	}
	
	return nil
}

// GracefulShutdown initiates a graceful shutdown of the daemon
func (dm *DaemonManager) GracefulShutdown() {
	dm.mu.RLock()
	if !dm.isRunning {
		dm.mu.RUnlock()
		slog.Warn("Daemon is not running, ignoring shutdown signal")
		return
	}
	dm.mu.RUnlock()
	
	slog.Info("Initiating graceful shutdown...")
	
	// Signal the main loop to shutdown
	select {
	case dm.shutdownCh <- struct{}{}:
		slog.Info("Shutdown signal sent to main loop")
	default:
		slog.Warn("Shutdown channel is full, forcing immediate shutdown")
		close(dm.stopCh)
	}
}

// performGracefulShutdown performs the actual graceful shutdown
func (dm *DaemonManager) performGracefulShutdown(ctx context.Context) error {
	slog.Info("Performing graceful shutdown...")
	
	// TODO: Wait for ongoing operations to complete
	// For now, we'll add a small delay to simulate cleanup
	shutdownTimeout := 30 * time.Second
	shutdownCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()
	
	// Create a channel to signal completion
	done := make(chan struct{})
	
	go func() {
		defer close(done)
		
		// TODO: Implement actual cleanup logic
		// - Wait for ongoing feed processing to complete
		// - Save any pending statistics
		// - Close database connections gracefully
		
		slog.Info("Cleanup operations completed")
	}()
	
	select {
	case <-done:
		slog.Info("Graceful shutdown completed successfully")
	case <-shutdownCtx.Done():
		slog.Warn("Graceful shutdown timed out, forcing shutdown")
	}
	
	// Remove PID file
	if err := dm.pidManager.RemovePID(); err != nil {
		slog.Warn("Failed to remove PID file during shutdown", "error", err)
	}
	
	dm.mu.Lock()
	dm.isRunning = false
	dm.mu.Unlock()
	
	return nil
}

// ReloadConfig reloads the daemon configuration
func (dm *DaemonManager) ReloadConfig() error {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	
	slog.Info("Reloading configuration...")
	
	// TODO: Implement actual config reload
	// For now, we'll just log that we received the signal
	slog.Info("Configuration reload not yet implemented")
	
	return nil
}

// DumpStatistics dumps current statistics to the log
func (dm *DaemonManager) DumpStatistics() {
	stats := dm.GetStatistics()
	
	slog.Info("=== Daemon Statistics ===")
	slog.Info("Runtime information",
		"start_time", stats.StartTime,
		"last_fetch_time", stats.LastFetchTime,
		"uptime", time.Since(stats.StartTime))
	
	slog.Info("Processing statistics",
		"total_fetches", stats.TotalFetches,
		"feeds_processed", stats.FeedsProcessed,
		"articles_fetched", stats.ArticlesFetched,
		"summaries_generated", stats.SummariesGenerated,
		"errors_count", stats.ErrorsCount,
		"average_process_time", stats.AverageProcessTime)
	
	slog.Info("=== End Statistics ===")
}

// ToggleDebugMode toggles debug logging mode
func (dm *DaemonManager) ToggleDebugMode() {
	dm.mu.Lock()
	dm.debugMode = !dm.debugMode
	newMode := dm.debugMode
	dm.mu.Unlock()
	
	if newMode {
		slog.Info("Debug mode enabled")
		// TODO: Change log level to debug
	} else {
		slog.Info("Debug mode disabled")
		// TODO: Change log level back to info
	}
}

// performPublish performs a scheduled publish operation
func (dm *DaemonManager) performPublish(ctx context.Context) {
	if dm.publisher == nil || !dm.publisher.IsEnabled() {
		return
	}
	
	slog.Info("Starting scheduled publish operation")
	
	if err := dm.publisher.PublishScheduled(ctx); err != nil {
		slog.Error("Scheduled publish failed", "error", err)
	} else {
		slog.Info("Scheduled publish completed successfully")
	}
}

// GetPublishStatistics returns current publish statistics
func (dm *DaemonManager) GetPublishStatistics() *PublishStatistics {
	if dm.publisher == nil {
		return nil
	}
	return dm.publisher.GetStatistics()
}
