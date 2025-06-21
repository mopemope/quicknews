package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"syscall"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/mopemope/quicknews/config"
	"github.com/mopemope/quicknews/daemon"
	"github.com/mopemope/quicknews/ent"
)

// DaemonCmd represents the daemon command group
type DaemonCmd struct {
	Start   DaemonStartCmd   `cmd:"" help:"Start daemon process"`
	Stop    DaemonStopCmd    `cmd:"" help:"Stop daemon process"`
	Restart DaemonRestartCmd `cmd:"" help:"Restart daemon process"`
	Status  DaemonStatusCmd  `cmd:"" help:"Show daemon status"`
}

// DaemonStartCmd represents the daemon start command
type DaemonStartCmd struct {
	Interval        time.Duration `short:"i" default:"1h" help:"Fetch interval (e.g., 30m, 1h, 2h)"`
	PidFile         string        `default:"~/quicknews.pid" help:"PID file path"`
	MaxWorkers      int           `default:"5" help:"Maximum number of concurrent workers"`
	Detach          bool          `short:"d" default:"false" help:"Run as detached daemon process"`
	HealthCheckPort int           `default:"8080" help:"Health check HTTP server port (0 to disable)"`
}

// DaemonStopCmd represents the daemon stop command
type DaemonStopCmd struct {
	PidFile string `default:"~/quicknews.pid" help:"PID file path"`
	Force   bool   `short:"f" help:"Force stop (SIGKILL instead of SIGTERM)"`
}

// DaemonRestartCmd represents the daemon restart command
type DaemonRestartCmd struct {
	Interval        time.Duration `short:"i" default:"1h" help:"Fetch interval (e.g., 30m, 1h, 2h)"`
	PidFile         string        `default:"~/quicknews.pid" help:"PID file path"`
	MaxWorkers      int           `default:"5" help:"Maximum number of concurrent workers"`
	Detach          bool          `short:"d" default:"false" help:"Run as detached daemon process"`
	HealthCheckPort int           `default:"8080" help:"Health check HTTP server port (0 to disable)"`
}

// DaemonStatusCmd represents the daemon status command
type DaemonStatusCmd struct {
	PidFile string `default:"~/quicknews.pid" help:"PID file path"`
	Verbose bool   `short:"v" help:"Show detailed statistics"`
}

// Run starts the daemon process
func (cmd *DaemonStartCmd) Run(client *ent.Client, config *config.Config) error {
	pidManager := daemon.NewPIDManager(cmd.PidFile)
	
	// Check if daemon is already running
	if running, pid, err := pidManager.IsRunning(); err == nil && running {
		return errors.Errorf("daemon is already running with PID %d", pid)
	}
	
	// Validate interval
	if cmd.Interval < time.Minute {
		return errors.New("interval must be at least 1 minute")
	}
	
	// Create daemon configuration
	daemonConfig := &daemon.DaemonConfig{
		Interval:        cmd.Interval,
		MaxWorkers:      cmd.MaxWorkers,
		PidFile:         cmd.PidFile,
		Detach:          cmd.Detach,
		HealthCheckPort: cmd.HealthCheckPort,
	}
	
	// Create daemon manager
	manager, err := daemon.NewDaemonManager(client, config, daemonConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create daemon manager")
	}
	
	if cmd.Detach {
		return cmd.startDetached(manager)
	} else {
		return cmd.startForeground(manager)
	}
}

// startForeground starts the daemon in foreground mode
func (cmd *DaemonStartCmd) startForeground(manager *daemon.DaemonManager) error {
	ctx := context.Background()
	
	fmt.Printf("Starting quicknews daemon in foreground mode...\n")
	fmt.Printf("Interval: %v\n", cmd.Interval)
	fmt.Printf("Max Workers: %d\n", cmd.MaxWorkers)
	fmt.Printf("PID File: %s\n", cmd.PidFile)
	if cmd.HealthCheckPort > 0 {
		fmt.Printf("Health Check Port: %d\n", cmd.HealthCheckPort)
	}
	fmt.Printf("Press Ctrl+C to stop\n\n")
	
	return manager.Start(ctx)
}

// startDetached starts the daemon in detached mode
func (cmd *DaemonStartCmd) startDetached(manager *daemon.DaemonManager) error {
	// For now, we'll implement a simple detached mode
	// In a production environment, you might want to use a proper daemonization library
	
	fmt.Printf("Starting quicknews daemon in detached mode...\n")
	fmt.Printf("Interval: %v\n", cmd.Interval)
	fmt.Printf("Max Workers: %d\n", cmd.MaxWorkers)
	fmt.Printf("PID File: %s\n", cmd.PidFile)
	if cmd.HealthCheckPort > 0 {
		fmt.Printf("Health Check Port: %d\n", cmd.HealthCheckPort)
	}
	
	// Start in background
	ctx := context.Background()
	go func() {
		if err := manager.Start(ctx); err != nil {
			slog.Error("Daemon failed", "error", err)
		}
	}()
	
	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)
	
	// Check if it started successfully
	pidManager := daemon.NewPIDManager(cmd.PidFile)
	if running, pid, err := pidManager.IsRunning(); err == nil && running {
		fmt.Printf("Daemon started successfully with PID %d\n", pid)
		return nil
	} else {
		return errors.New("failed to start daemon")
	}
}

// Run stops the daemon process
func (cmd *DaemonStopCmd) Run(client *ent.Client, config *config.Config) error {
	pidManager := daemon.NewPIDManager(cmd.PidFile)
	
	// Check if daemon is running
	running, pid, err := pidManager.IsRunning()
	if err != nil {
		return errors.Wrap(err, "failed to check daemon status")
	}
	
	if !running {
		fmt.Println("Daemon is not running")
		return nil
	}
	
	// Choose signal based on force flag
	signal := syscall.SIGTERM
	if cmd.Force {
		signal = syscall.SIGKILL
		fmt.Printf("Force stopping daemon (PID %d)...\n", pid)
	} else {
		fmt.Printf("Stopping daemon (PID %d)...\n", pid)
	}
	
	// Send signal to process
	if err := pidManager.KillProcess(signal); err != nil {
		return errors.Wrap(err, "failed to stop daemon")
	}
	
	// Wait for process to stop (only for SIGTERM)
	if !cmd.Force {
		for i := 0; i < 30; i++ { // Wait up to 30 seconds
			time.Sleep(1 * time.Second)
			if running, _, _ := pidManager.IsRunning(); !running {
				break
			}
		}
	}
	
	// Clean up PID file
	if err := pidManager.RemovePID(); err != nil {
		slog.Warn("Failed to remove PID file", "error", err)
	}
	
	fmt.Println("Daemon stopped successfully")
	return nil
}

// Run restarts the daemon process
func (cmd *DaemonRestartCmd) Run(client *ent.Client, config *config.Config) error {
	// Stop the daemon first
	stopCmd := &DaemonStopCmd{
		PidFile: cmd.PidFile,
		Force:   false,
	}
	
	if err := stopCmd.Run(client, config); err != nil {
		return errors.Wrap(err, "failed to stop daemon")
	}
	
	// Wait a moment
	time.Sleep(1 * time.Second)
	
	// Start the daemon
	startCmd := &DaemonStartCmd{
		Interval:        cmd.Interval,
		PidFile:         cmd.PidFile,
		MaxWorkers:      cmd.MaxWorkers,
		Detach:          cmd.Detach,
		HealthCheckPort: cmd.HealthCheckPort,
	}
	
	return startCmd.Run(client, config)
}

// Run shows the daemon status
func (cmd *DaemonStatusCmd) Run(client *ent.Client, config *config.Config) error {
	pidManager := daemon.NewPIDManager(cmd.PidFile)
	
	// Check if daemon is running
	running, pid, err := pidManager.IsRunning()
	if err != nil {
		fmt.Printf("Status: Unknown (error: %v)\n", err)
		return nil
	}
	
	if !running {
		fmt.Println("Status: Stopped")
		return nil
	}
	
	fmt.Printf("Status: Running\n")
	fmt.Printf("PID: %d\n", pid)
	fmt.Printf("PID File: %s\n", cmd.PidFile)
	
	if cmd.Verbose {
		// Try to get detailed statistics via health check endpoint
		fmt.Println("\n=== Detailed Status ===")
		
		// Try common health check ports
		healthPorts := []int{8080, 8081, 8082}
		var healthData map[string]interface{}
		
		for _, port := range healthPorts {
			if data, err := fetchHealthData(port); err == nil {
				healthData = data
				fmt.Printf("Health Check Port: %d\n", port)
				break
			}
		}
		
		if healthData != nil {
			if status, ok := healthData["status"].(string); ok {
				fmt.Printf("Health Status: %s\n", status)
			}
			if uptime, ok := healthData["uptime"].(string); ok {
				fmt.Printf("Uptime: %s\n", uptime)
			}
			if stats, ok := healthData["statistics"].(map[string]interface{}); ok {
				fmt.Println("\n=== Statistics ===")
				if totalFetches, ok := stats["total_fetches"].(float64); ok {
					fmt.Printf("Total Fetches: %.0f\n", totalFetches)
				}
				if feedsProcessed, ok := stats["feeds_processed"].(float64); ok {
					fmt.Printf("Feeds Processed: %.0f\n", feedsProcessed)
				}
				if articlesProcessed, ok := stats["articles_fetched"].(float64); ok {
					fmt.Printf("Articles Fetched: %.0f\n", articlesProcessed)
				}
				if summariesGenerated, ok := stats["summaries_generated"].(float64); ok {
					fmt.Printf("Summaries Generated: %.0f\n", summariesGenerated)
				}
				if errorsCount, ok := stats["errors_count"].(float64); ok {
					fmt.Printf("Errors: %.0f\n", errorsCount)
				}
				if avgProcessTime, ok := stats["average_process_time_seconds"].(float64); ok {
					fmt.Printf("Average Process Time: %.2fs\n", avgProcessTime)
				}
			}
		} else {
			fmt.Println("Health check endpoint not accessible.")
			fmt.Println("The daemon may not have health checking enabled.")
		}
		
		fmt.Println("\n=== Control Commands ===")
		fmt.Printf("Stop daemon: quicknews daemon stop --pid-file=%s\n", cmd.PidFile)
		fmt.Println("Send SIGUSR1 for stats dump: kill -USR1", pid)
		fmt.Println("Send SIGHUP for config reload: kill -HUP", pid)
	}
	
	return nil
}

// fetchHealthData attempts to fetch health data from the daemon's health endpoint
func fetchHealthData(port int) (map[string]interface{}, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	
	url := fmt.Sprintf("http://localhost:%d/health", port)
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			slog.Warn("Failed to close response body", "error", closeErr)
		}
	}()
	
	var healthData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&healthData); err != nil {
		return nil, err
	}
	
	return healthData, nil
}
