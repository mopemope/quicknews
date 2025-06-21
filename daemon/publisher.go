package daemon

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	configPkg "github.com/mopemope/quicknews/config"
	"github.com/mopemope/quicknews/ent"
	"github.com/mopemope/quicknews/models/article"
	"github.com/mopemope/quicknews/models/feed"
	summaryModel "github.com/mopemope/quicknews/models/summary"
	"github.com/mopemope/quicknews/org"
	"github.com/mopemope/quicknews/rss"
	"github.com/mopemope/quicknews/storage"
	"github.com/mopemope/quicknews/tts"
)

// DaemonPublisher handles automated publishing functionality within the daemon
type DaemonPublisher struct {
	feedRepos    feed.FeedRepository
	articleRepos article.ArticleRepository
	summaryRepos summaryModel.SummaryRepository
	rssGenerator *rss.RSS
	r2Client     *storage.R2Storage
	config       *configPkg.Config
	stats        *PublishStatistics
	
	// Configuration
	enabled              bool
	schedule             string
	scheduleTime         time.Time
	weekday              time.Weekday
	rangeDays            int
	autoGenerateAudio    bool
	cleanupTempFiles     bool
	maxFileSizeMB        int
	parallelUploads      int
	retryAttempts        int
	retryDelay           time.Duration
}

// PublishStatistics holds daemon publish runtime statistics
type PublishStatistics struct {
	mu                    sync.RWMutex
	TotalPublishes        int64     `json:"total_publishes"`
	SuccessfulPublishes   int64     `json:"successful_publishes"`
	FailedPublishes       int64     `json:"failed_publishes"`
	LastPublishTime       time.Time `json:"last_publish_time"`
	LastPublishDuration   float64   `json:"last_publish_duration_seconds"`
	AudioFilesGenerated   int64     `json:"audio_files_generated"`
	AudioFilesUploaded    int64     `json:"audio_files_uploaded"`
	TotalUploadSize       int64     `json:"total_upload_size_bytes"`
	LastError             string    `json:"last_error,omitempty"`
}

// NewDaemonPublisher creates a new daemon publisher
func NewDaemonPublisher(ctx context.Context, client *ent.Client, config *configPkg.Config) (*DaemonPublisher, error) {
	if ctx == nil {
		return nil, errors.New("context cannot be nil")
	}
	if client == nil {
		return nil, errors.New("database client cannot be nil")
	}
	if config == nil {
		return nil, errors.New("config cannot be nil")
	}

	// Check if publish functionality is properly configured
	if config.AudioPath == nil || config.Podcast == nil || config.Cloudflare == nil {
		return nil, errors.New("publish functionality requires AudioPath, Podcast, and Cloudflare configuration")
	}
	
	// Initialize repositories
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
	
	// Initialize RSS generator
	rssGenerator := rss.NewRSS(config.Podcast)
	if rssGenerator == nil {
		return nil, errors.New("failed to create RSS generator")
	}
	
	// Initialize R2 client
	r2Client, err := storage.NewR2Storage(ctx, config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize R2 storage client")
	}
	if r2Client == nil {
		return nil, errors.New("R2 storage client creation returned nil")
	}
	
	// Parse daemon configuration
	daemonConfig := config.Daemon
	if daemonConfig == nil {
		daemonConfig = &configPkg.DaemonConfig{
			PublishEnabled:   false,
			PublishSchedule:  "daily",
			PublishTime:      "06:00",
			PublishRangeDays: 1,
		}
	}
	
	// Parse schedule time
	scheduleTime, err := time.Parse("15:04", daemonConfig.PublishTime)
	if err != nil {
		return nil, errors.Wrap(err, "invalid publish_time format, expected HH:MM")
	}
	
	// Parse weekday
	var weekday time.Weekday
	if daemonConfig.PublishSchedule == "weekly" {
		switch daemonConfig.PublishWeekday {
		case "sunday":
			weekday = time.Sunday
		case "monday":
			weekday = time.Monday
		case "tuesday":
			weekday = time.Tuesday
		case "wednesday":
			weekday = time.Wednesday
		case "thursday":
			weekday = time.Thursday
		case "friday":
			weekday = time.Friday
		case "saturday":
			weekday = time.Saturday
		default:
			weekday = time.Sunday // Default to Sunday
		}
	}
	
	// Parse retry delay
	retryDelay := 30 * time.Second
	if daemonConfig.Publish != nil && daemonConfig.Publish.RetryDelay != "" {
		if parsed, err := time.ParseDuration(daemonConfig.Publish.RetryDelay); err == nil {
			retryDelay = parsed
		}
	}
	
	// Set default values for publish config
	publishConfig := daemonConfig.Publish
	if publishConfig == nil {
		publishConfig = &configPkg.DaemonPublishConfig{
			AutoGenerateMissingAudio: true,
			CleanupTempFiles:         true,
			MaxFileSizeMB:           100,
			ParallelUploads:         3,
			RetryAttempts:           3,
		}
	}
	
	return &DaemonPublisher{
		feedRepos:         feedRepos,
		articleRepos:      articleRepos,
		summaryRepos:      summaryRepos,
		rssGenerator:      rssGenerator,
		r2Client:          r2Client,
		config:            config,
		enabled:           daemonConfig.PublishEnabled,
		schedule:          daemonConfig.PublishSchedule,
		scheduleTime:      scheduleTime,
		weekday:           weekday,
		rangeDays:         daemonConfig.PublishRangeDays,
		autoGenerateAudio: publishConfig.AutoGenerateMissingAudio,
		cleanupTempFiles:  publishConfig.CleanupTempFiles,
		maxFileSizeMB:     publishConfig.MaxFileSizeMB,
		parallelUploads:   publishConfig.ParallelUploads,
		retryAttempts:     publishConfig.RetryAttempts,
		retryDelay:        retryDelay,
		stats: &PublishStatistics{
			TotalPublishes:      0,
			SuccessfulPublishes: 0,
			FailedPublishes:     0,
		},
	}, nil
}

// IsEnabled returns whether publish functionality is enabled
func (dp *DaemonPublisher) IsEnabled() bool {
	return dp.enabled
}

// ShouldPublishNow checks if it's time to publish based on the schedule
func (dp *DaemonPublisher) ShouldPublishNow() bool {
	if !dp.enabled {
		return false
	}
	
	now := time.Now()
	
	switch dp.schedule {
	case "daily":
		// Check if it's the right time of day
		currentTime := time.Date(0, 1, 1, now.Hour(), now.Minute(), 0, 0, time.UTC)
		scheduleTime := time.Date(0, 1, 1, dp.scheduleTime.Hour(), dp.scheduleTime.Minute(), 0, 0, time.UTC)
		
		// Allow a 1-minute window for execution
		diff := currentTime.Sub(scheduleTime)
		return diff >= 0 && diff < time.Minute
		
	case "weekly":
		// Check if it's the right day and time
		if now.Weekday() != dp.weekday {
			return false
		}
		
		currentTime := time.Date(0, 1, 1, now.Hour(), now.Minute(), 0, 0, time.UTC)
		scheduleTime := time.Date(0, 1, 1, dp.scheduleTime.Hour(), dp.scheduleTime.Minute(), 0, 0, time.UTC)
		
		diff := currentTime.Sub(scheduleTime)
		return diff >= 0 && diff < time.Minute
		
	case "manual":
		return false // Manual mode never auto-publishes
		
	default:
		return false
	}
}

// PublishScheduled performs a scheduled publish operation
func (dp *DaemonPublisher) PublishScheduled(ctx context.Context) error {
	startTime := time.Now()
	
	slog.Info("Starting scheduled publish operation",
		"schedule", dp.schedule,
		"range_days", dp.rangeDays)
	
	// Update statistics
	dp.stats.mu.Lock()
	dp.stats.TotalPublishes++
	dp.stats.mu.Unlock()
	
	// Determine target dates
	targetDate := time.Now()
	if dp.schedule == "weekly" {
		// For weekly, process the last week
		targetDate = targetDate.AddDate(0, 0, -7)
	}
	
	// Process multiple days if configured
	for i := 0; i < dp.rangeDays; i++ {
		processDate := targetDate.AddDate(0, 0, -i)
		dateStr := processDate.Format("2006-01-02")
		
		slog.Info("Processing date for publish", "date", dateStr)
		
		if err := dp.publishForDate(ctx, dateStr); err != nil {
			slog.Error("Failed to publish for date", "date", dateStr, "error", err)
			
			dp.stats.mu.Lock()
			dp.stats.FailedPublishes++
			dp.stats.LastError = err.Error()
			dp.stats.mu.Unlock()
			
			return errors.Wrap(err, fmt.Sprintf("failed to publish for date %s", dateStr))
		}
	}
	
	// Generate and upload RSS feed
	if err := dp.publishRSS(ctx); err != nil {
		slog.Error("Failed to publish RSS feed", "error", err)
		
		dp.stats.mu.Lock()
		dp.stats.FailedPublishes++
		dp.stats.LastError = err.Error()
		dp.stats.mu.Unlock()
		
		return errors.Wrap(err, "failed to publish RSS feed")
	}
	
	// Update success statistics
	duration := time.Since(startTime)
	dp.stats.mu.Lock()
	dp.stats.SuccessfulPublishes++
	dp.stats.LastPublishTime = startTime
	dp.stats.LastPublishDuration = duration.Seconds()
	dp.stats.LastError = "" // Clear last error on success
	dp.stats.mu.Unlock()
	
	slog.Info("Scheduled publish completed successfully",
		"duration", duration,
		"processed_days", dp.rangeDays)
	
	return nil
}

// GetStatistics returns current publish statistics
func (dp *DaemonPublisher) GetStatistics() *PublishStatistics {
	dp.stats.mu.RLock()
	defer dp.stats.mu.RUnlock()
	
	// Create a copy to avoid race conditions, excluding the mutex
	statsCopy := PublishStatistics{
		TotalPublishes:        dp.stats.TotalPublishes,
		SuccessfulPublishes:   dp.stats.SuccessfulPublishes,
		FailedPublishes:       dp.stats.FailedPublishes,
		LastPublishTime:       dp.stats.LastPublishTime,
		LastPublishDuration:   dp.stats.LastPublishDuration,
		AudioFilesGenerated:   dp.stats.AudioFilesGenerated,
		AudioFilesUploaded:    dp.stats.AudioFilesUploaded,
		TotalUploadSize:       dp.stats.TotalUploadSize,
		LastError:             dp.stats.LastError,
	}
	return &statsCopy
}

// publishForDate publishes content for a specific date
func (dp *DaemonPublisher) publishForDate(ctx context.Context, dateStr string) error {
	// Get all feeds
	feeds, err := dp.feedRepos.All(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get feeds")
	}
	
	// Process each feed for the given date
	for _, feed := range feeds {
		if feed.IsBookmark {
			continue // Skip bookmark feeds
		}
		
		if err := dp.processFeedForDate(ctx, feed, dateStr); err != nil {
			slog.Error("Failed to process feed for date",
				"feed", feed.Title,
				"date", dateStr,
				"error", err)
			// Continue with other feeds even if one fails
		}
	}
	
	return nil
}

// processFeedForDate processes a single feed for a specific date
func (dp *DaemonPublisher) processFeedForDate(ctx context.Context, feedData *ent.Feed, dateStr string) error {
	feedID := feedData.ID
	feedName := feedData.Title
	
	slog.Debug("Processing feed for date",
		"feed", feedName,
		"date", dateStr)
	
	// Get articles for the specified date
	articles, err := dp.articleRepos.GetByDate(ctx, feedID, dateStr)
	if err != nil {
		return errors.Wrap(err, "failed to get articles by date")
	}
	
	if len(articles) == 0 {
		slog.Debug("No articles found for feed and date",
			"feed", feedName,
			"date", dateStr)
		return nil
	}
	
	// Collect audio files
	audioFiles := make([]string, 0)
	for _, article := range articles {
		audioFile, err := dp.processArticleAudio(ctx, article, feedData)
		if err != nil {
			slog.Error("Failed to process article audio",
				"article", article.Title,
				"error", err)
			continue // Skip this article but continue with others
		}
		
		if audioFile != "" {
			audioFiles = append(audioFiles, audioFile)
		}
	}
	
	if len(audioFiles) == 0 {
		slog.Info("No audio files found for feed and date",
			"feed", feedName,
			"date", dateStr)
		return nil
	}
	
	// Merge audio files and upload
	return dp.mergeAndUploadAudio(ctx, audioFiles, feedName, dateStr)
}

// processArticleAudio processes audio for a single article
func (dp *DaemonPublisher) processArticleAudio(ctx context.Context, article *ent.Article, feedData *ent.Feed) (string, error) {
	summary := article.Edges.Summary
	if summary == nil {
		slog.Debug("Article has no summary, skipping", "article", article.Title)
		return "", nil
	}
	
	// Set feed edge for summary
	summary.Edges.Feed = feedData
	
	// Check if audio file already exists
	if summary.AudioFile != "" {
		audioPath := filepath.Join(*dp.config.AudioPath, summary.AudioFile)
		if _, err := os.Stat(audioPath); err == nil {
			slog.Debug("Audio file already exists", "file", summary.AudioFile)
			return audioPath, nil
		}
	}
	
	// Generate audio if enabled and content is not too long
	if dp.autoGenerateAudio {
		contentLength := len(summary.Summary) + len(summary.Title)
		if contentLength > 4500 {
			slog.Warn("Skipping audio generation for long content",
				"title", summary.Title,
				"length", contentLength)
			return "", nil
		}
		
		slog.Info("Generating missing audio file", "title", summary.Title)
		
		filename, err := summaryModel.SaveAudioData(ctx, summary, dp.config)
		if err != nil {
			return "", errors.Wrap(err, "failed to generate audio")
		}
		
		if filename != nil {
			if err := dp.summaryRepos.UpdateAudioFile(ctx, summary.ID, *filename); err != nil {
				return "", errors.Wrap(err, "failed to update audio file path")
			}
			
			dp.stats.mu.Lock()
			dp.stats.AudioFilesGenerated++
			dp.stats.mu.Unlock()
			
			audioPath := filepath.Join(*dp.config.AudioPath, *filename)
			return audioPath, nil
		}
	}
	
	return "", nil
}

// mergeAndUploadAudio merges audio files and uploads to R2
func (dp *DaemonPublisher) mergeAndUploadAudio(ctx context.Context, audioFiles []string, feedName, dateStr string) error {
	// Create output filename
	outputFilename := org.ConvertPathName(dateStr+"_"+feedName) + ".mp3"
	outputPath := filepath.Join(os.TempDir(), outputFilename)
	
	// Cleanup temp file
	if dp.cleanupTempFiles {
		defer func() {
			if err := os.Remove(outputPath); err != nil {
				slog.Warn("Failed to remove temporary file", "path", outputPath, "error", err)
			}
		}()
	}
	
	slog.Info("Merging audio files",
		"feed", feedName,
		"date", dateStr,
		"file_count", len(audioFiles),
		"output", outputFilename)
	
	// Merge MP3 files
	if err := tts.MergeMP3(outputPath, audioFiles); err != nil {
		return errors.Wrap(err, "failed to merge MP3 files")
	}
	
	// Get file info
	fileInfo, err := os.Stat(outputPath)
	if err != nil {
		return errors.Wrap(err, "failed to get file info")
	}
	
	fileSize := fileInfo.Size()
	
	// Check file size limit
	if dp.maxFileSizeMB > 0 && fileSize > int64(dp.maxFileSizeMB*1024*1024) {
		return errors.Errorf("merged file size (%d bytes) exceeds limit (%d MB)",
			fileSize, dp.maxFileSizeMB)
	}
	
	// Open file for upload
	file, err := os.Open(outputPath)
	if err != nil {
		return errors.Wrap(err, "failed to open merged file")
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			slog.Warn("Failed to close file", "error", closeErr, "path", outputPath)
		}
	}()
	
	// Upload to R2
	slog.Info("Uploading audio file to R2",
		"filename", outputFilename,
		"size_bytes", fileSize)
	
	if err := dp.r2Client.Upload(ctx, outputFilename, file, "audio/mpeg"); err != nil {
		return errors.Wrap(err, "failed to upload audio file")
	}
	
	// Update statistics
	dp.stats.mu.Lock()
	dp.stats.AudioFilesUploaded++
	dp.stats.TotalUploadSize += fileSize
	dp.stats.mu.Unlock()
	
	// Add item to RSS feed
	pubDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return errors.Wrap(err, "failed to parse date for RSS")
	}
	
	podcastConfig := dp.config.Podcast
	dp.rssGenerator.AddItem(rss.RSSItem{
		Title:       fmt.Sprintf("%s %s Podcast", dateStr, feedName),
		Link:        podcastConfig.PublishURL + "/" + outputFilename,
		Guid:        podcastConfig.PublishURL + "/" + outputFilename,
		PubDate:     pubDate.UTC().Format(time.RFC1123),
		Description: fmt.Sprintf("This is %s %s podcast", dateStr, feedName),
		AudioURL:    podcastConfig.PublishURL + "/" + outputFilename,
		Length:      fmt.Sprintf("%d", fileSize),
		MimeType:    "audio/mpeg",
	})
	
	slog.Info("Successfully processed feed for date",
		"feed", feedName,
		"date", dateStr,
		"output_file", outputFilename,
		"size_bytes", fileSize)
	
	return nil
}

// publishRSS generates and uploads the RSS feed
func (dp *DaemonPublisher) publishRSS(ctx context.Context) error {
	rssOutputPath := filepath.Join(os.TempDir(), "rss.xml")
	
	// Cleanup temp RSS file
	if dp.cleanupTempFiles {
		defer func() {
			if err := os.Remove(rssOutputPath); err != nil {
				slog.Warn("Failed to remove temporary RSS file", "path", rssOutputPath, "error", err)
			}
		}()
	}
	
	slog.Info("Generating RSS feed", "path", rssOutputPath)
	
	// Write RSS to file
	if err := dp.rssGenerator.WriteToFile(rssOutputPath); err != nil {
		return errors.Wrap(err, "failed to write RSS to file")
	}
	
	// Open RSS file for upload
	rssFile, err := os.Open(rssOutputPath)
	if err != nil {
		return errors.Wrap(err, "failed to open RSS file")
	}
	defer func() {
		if closeErr := rssFile.Close(); closeErr != nil {
			slog.Warn("Failed to close RSS file", "error", closeErr, "path", rssOutputPath)
		}
	}()
	
	// Upload RSS to R2
	slog.Info("Uploading RSS feed to R2")
	
	if err := dp.r2Client.Upload(ctx, "rss.xml", rssFile, "application/rss+xml"); err != nil {
		return errors.Wrap(err, "failed to upload RSS file")
	}
	
	slog.Info("Successfully published RSS feed")
	return nil
}
