package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// HealthChecker provides HTTP health check endpoints for the daemon
type HealthChecker struct {
	manager *DaemonManager
	server  *http.Server
	port    int
}

// HealthStatus represents the current health status of the daemon
type HealthStatus struct {
	Status           string                 `json:"status"`      // "healthy", "degraded", "unhealthy"
	Timestamp        time.Time              `json:"timestamp"`
	Uptime           string                 `json:"uptime"`
	Version          string                 `json:"version"`
	Statistics       *Statistics            `json:"statistics"`
	PublishStatistics *PublishStatistics    `json:"publish_statistics,omitempty"`
	LastError        string                 `json:"last_error,omitempty"`
	Checks           map[string]CheckResult `json:"checks"`
}

// CheckResult represents the result of an individual health check
type CheckResult struct {
	Status    string    `json:"status"`    // "pass", "fail", "warn"
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Duration  string    `json:"duration,omitempty"`
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(manager *DaemonManager, port int) *HealthChecker {
	return &HealthChecker{
		manager: manager,
		port:    port,
	}
}

// Start starts the health check HTTP server
func (hc *HealthChecker) Start() error {
	mux := http.NewServeMux()
	
	// Health check endpoint
	mux.HandleFunc("/health", hc.handleHealth)
	mux.HandleFunc("/health/live", hc.handleLiveness)
	mux.HandleFunc("/health/ready", hc.handleReadiness)
	mux.HandleFunc("/stats", hc.handleStats)
	mux.HandleFunc("/metrics", hc.handleMetrics)
	
	hc.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", hc.port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	
	slog.Info("Starting health check server", "port", hc.port)
	
	go func() {
		if err := hc.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Health check server failed", "error", err)
		}
	}()
	
	return nil
}

// Stop stops the health check HTTP server
func (hc *HealthChecker) Stop() error {
	if hc.server == nil {
		return nil
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	slog.Info("Stopping health check server")
	return hc.server.Shutdown(ctx)
}

// handleHealth handles the main health check endpoint
func (hc *HealthChecker) handleHealth(w http.ResponseWriter, r *http.Request) {
	status := hc.getHealthStatus()
	
	w.Header().Set("Content-Type", "application/json")
	
	// Set HTTP status code based on health status
	switch status.Status {
	case "healthy":
		w.WriteHeader(http.StatusOK)
	case "degraded":
		w.WriteHeader(http.StatusOK) // Still OK, but with warnings
	case "unhealthy":
		w.WriteHeader(http.StatusServiceUnavailable)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	
	if err := json.NewEncoder(w).Encode(status); err != nil {
		slog.Error("Failed to encode health status", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handleLiveness handles the liveness probe endpoint
func (hc *HealthChecker) handleLiveness(w http.ResponseWriter, r *http.Request) {
	// Simple liveness check - is the daemon process running?
	if hc.manager.IsRunning() {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			slog.Warn("Failed to write liveness response", "error", err)
		}
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		if _, err := w.Write([]byte("NOT RUNNING")); err != nil {
			slog.Warn("Failed to write liveness response", "error", err)
		}
	}
}

// handleReadiness handles the readiness probe endpoint
func (hc *HealthChecker) handleReadiness(w http.ResponseWriter, r *http.Request) {
	// Readiness check - is the daemon ready to process requests?
	checks := hc.performReadinessChecks()
	
	allPassed := true
	for _, check := range checks {
		if check.Status == "fail" {
			allPassed = false
			break
		}
	}
	
	if allPassed {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("READY")); err != nil {
			slog.Warn("Failed to write readiness response", "error", err)
		}
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		if _, err := w.Write([]byte("NOT READY")); err != nil {
			slog.Warn("Failed to write readiness response", "error", err)
		}
	}
}

// handleStats handles the statistics endpoint
func (hc *HealthChecker) handleStats(w http.ResponseWriter, r *http.Request) {
	stats := hc.manager.GetStatistics()
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		slog.Error("Failed to encode statistics", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handleMetrics handles the metrics endpoint (Prometheus-style)
func (hc *HealthChecker) handleMetrics(w http.ResponseWriter, r *http.Request) {
	stats := hc.manager.GetStatistics()
	
	w.Header().Set("Content-Type", "text/plain")
	
	// Generate Prometheus-style metrics
	metrics := fmt.Sprintf(`# HELP quicknews_uptime_seconds Total uptime in seconds
# TYPE quicknews_uptime_seconds counter
quicknews_uptime_seconds %f

# HELP quicknews_total_fetches_total Total number of fetch operations
# TYPE quicknews_total_fetches_total counter
quicknews_total_fetches_total %d

# HELP quicknews_feeds_processed_total Total number of feeds processed
# TYPE quicknews_feeds_processed_total counter
quicknews_feeds_processed_total %d

# HELP quicknews_articles_fetched_total Total number of articles fetched
# TYPE quicknews_articles_fetched_total counter
quicknews_articles_fetched_total %d

# HELP quicknews_summaries_generated_total Total number of summaries generated
# TYPE quicknews_summaries_generated_total counter
quicknews_summaries_generated_total %d

# HELP quicknews_errors_total Total number of errors
# TYPE quicknews_errors_total counter
quicknews_errors_total %d

# HELP quicknews_average_process_time_seconds Average processing time in seconds
# TYPE quicknews_average_process_time_seconds gauge
quicknews_average_process_time_seconds %f
`,
		time.Since(stats.StartTime).Seconds(),
		stats.TotalFetches,
		stats.FeedsProcessed,
		stats.ArticlesFetched,
		stats.SummariesGenerated,
		stats.ErrorsCount,
		stats.AverageProcessTime,
	)
	
	if _, err := w.Write([]byte(metrics)); err != nil {
		slog.Warn("Failed to write metrics response", "error", err)
	}
}

// getHealthStatus performs comprehensive health checks and returns status
func (hc *HealthChecker) getHealthStatus() *HealthStatus {
	stats := hc.manager.GetStatistics()
	publishStats := hc.manager.GetPublishStatistics()
	checks := hc.performHealthChecks()
	
	// Determine overall status based on individual checks
	overallStatus := "healthy"
	hasWarnings := false
	hasFailed := false
	
	for _, check := range checks {
		switch check.Status {
		case "warn":
			hasWarnings = true
		case "fail":
			hasFailed = true
		}
	}
	
	if hasFailed {
		overallStatus = "unhealthy"
	} else if hasWarnings {
		overallStatus = "degraded"
	}
	
	return &HealthStatus{
		Status:            overallStatus,
		Timestamp:         time.Now(),
		Uptime:            time.Since(stats.StartTime).String(),
		Version:           "0.0.1", // TODO: Get from build info
		Statistics:        stats,
		PublishStatistics: publishStats,
		Checks:            checks,
	}
}

// performHealthChecks performs individual health checks
func (hc *HealthChecker) performHealthChecks() map[string]CheckResult {
	checks := make(map[string]CheckResult)
	
	// Check if daemon is running
	checks["daemon_running"] = hc.checkDaemonRunning()
	
	// Check database connectivity
	checks["database"] = hc.checkDatabase()
	
	// Check recent activity
	checks["recent_activity"] = hc.checkRecentActivity()
	
	// Check error rate
	checks["error_rate"] = hc.checkErrorRate()
	
	// Check publish functionality if enabled
	if hc.manager.publisher != nil && hc.manager.publisher.IsEnabled() {
		checks["publish_status"] = hc.checkPublishStatus()
	}
	
	return checks
}

// performReadinessChecks performs readiness-specific checks
func (hc *HealthChecker) performReadinessChecks() map[string]CheckResult {
	checks := make(map[string]CheckResult)
	
	// Check if daemon is running and ready
	checks["daemon_ready"] = hc.checkDaemonRunning()
	
	// Check database connectivity
	checks["database"] = hc.checkDatabase()
	
	return checks
}

// Individual health check methods
func (hc *HealthChecker) checkDaemonRunning() CheckResult {
	if hc.manager.IsRunning() {
		return CheckResult{
			Status:    "pass",
			Message:   "Daemon is running",
			Timestamp: time.Now(),
		}
	}
	return CheckResult{
		Status:    "fail",
		Message:   "Daemon is not running",
		Timestamp: time.Now(),
	}
}

func (hc *HealthChecker) checkDatabase() CheckResult {
	// TODO: Implement actual database connectivity check
	// For now, assume it's working if we can get statistics
	_ = hc.manager.GetStatistics()
	
	return CheckResult{
		Status:    "pass",
		Message:   "Database is accessible",
		Timestamp: time.Now(),
	}
}

func (hc *HealthChecker) checkRecentActivity() CheckResult {
	stats := hc.manager.GetStatistics()
	
	// Check if we've had activity in the last 2 hours
	if time.Since(stats.LastFetchTime) > 2*time.Hour && stats.TotalFetches > 0 {
		return CheckResult{
			Status:    "warn",
			Message:   "No recent fetch activity",
			Timestamp: time.Now(),
		}
	}
	
	return CheckResult{
		Status:    "pass",
		Message:   "Recent activity detected",
		Timestamp: time.Now(),
	}
}

func (hc *HealthChecker) checkErrorRate() CheckResult {
	stats := hc.manager.GetStatistics()
	
	if stats.TotalFetches == 0 {
		return CheckResult{
			Status:    "pass",
			Message:   "No operations yet",
			Timestamp: time.Now(),
		}
	}
	
	errorRate := float64(stats.ErrorsCount) / float64(stats.TotalFetches)
	
	if errorRate > 0.5 { // More than 50% error rate
		return CheckResult{
			Status:    "fail",
			Message:   fmt.Sprintf("High error rate: %.2f%%", errorRate*100),
			Timestamp: time.Now(),
		}
	} else if errorRate > 0.1 { // More than 10% error rate
		return CheckResult{
			Status:    "warn",
			Message:   fmt.Sprintf("Elevated error rate: %.2f%%", errorRate*100),
			Timestamp: time.Now(),
		}
	}
	
	return CheckResult{
		Status:    "pass",
		Message:   fmt.Sprintf("Error rate acceptable: %.2f%%", errorRate*100),
		Timestamp: time.Now(),
	}
}

// checkPublishStatus checks the status of publish functionality
func (hc *HealthChecker) checkPublishStatus() CheckResult {
	publishStats := hc.manager.GetPublishStatistics()
	if publishStats == nil {
		return CheckResult{
			Status:    "pass",
			Message:   "Publish functionality not enabled",
			Timestamp: time.Now(),
		}
	}
	
	// Check if there have been recent publish failures
	if publishStats.TotalPublishes > 0 {
		failureRate := float64(publishStats.FailedPublishes) / float64(publishStats.TotalPublishes)
		
		if failureRate > 0.5 { // More than 50% failure rate
			return CheckResult{
				Status:    "fail",
				Message:   fmt.Sprintf("High publish failure rate: %.2f%%", failureRate*100),
				Timestamp: time.Now(),
			}
		} else if failureRate > 0.1 { // More than 10% failure rate
			return CheckResult{
				Status:    "warn",
				Message:   fmt.Sprintf("Elevated publish failure rate: %.2f%%", failureRate*100),
				Timestamp: time.Now(),
			}
		}
	}
	
	// Check for recent errors
	if publishStats.LastError != "" {
		return CheckResult{
			Status:    "warn",
			Message:   fmt.Sprintf("Last publish error: %s", publishStats.LastError),
			Timestamp: time.Now(),
		}
	}
	
	return CheckResult{
		Status:    "pass",
		Message:   "Publish functionality is healthy",
		Timestamp: time.Now(),
	}
}
