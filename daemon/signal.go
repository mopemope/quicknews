package daemon

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

// SignalHandler manages OS signals for the daemon process
type SignalHandler struct {
	manager *DaemonManager
	signals chan os.Signal
	done    chan struct{}
}

// NewSignalHandler creates a new signal handler
func NewSignalHandler(manager *DaemonManager) *SignalHandler {
	return &SignalHandler{
		manager: manager,
		signals: make(chan os.Signal, 1),
		done:    make(chan struct{}),
	}
}

// Start begins listening for OS signals
func (sh *SignalHandler) Start() {
	// Register signal notifications
	signal.Notify(sh.signals,
		syscall.SIGTERM, // Termination request
		syscall.SIGINT,  // Interrupt (Ctrl+C)
		syscall.SIGHUP,  // Hangup (reload config)
		syscall.SIGUSR1, // User-defined signal 1 (stats dump)
		syscall.SIGUSR2, // User-defined signal 2 (toggle debug)
	)

	slog.Info("Signal handler started")

	go sh.handleSignals()
}

// Stop stops the signal handler
func (sh *SignalHandler) Stop() {
	signal.Stop(sh.signals)
	close(sh.done)
	slog.Info("Signal handler stopped")
}

// handleSignals processes incoming signals
func (sh *SignalHandler) handleSignals() {
	for {
		select {
		case sig := <-sh.signals:
			sh.processSignal(sig)
		case <-sh.done:
			return
		}
	}
}

// processSignal handles individual signals
func (sh *SignalHandler) processSignal(sig os.Signal) {
	slog.Info("Received signal", "signal", sig.String())

	switch sig {
	case syscall.SIGTERM:
		slog.Info("Received SIGTERM, initiating graceful shutdown")
		sh.manager.GracefulShutdown()

	case syscall.SIGINT:
		slog.Info("Received SIGINT (Ctrl+C), initiating graceful shutdown")
		sh.manager.GracefulShutdown()

	case syscall.SIGHUP:
		slog.Info("Received SIGHUP, reloading configuration")
		if err := sh.manager.ReloadConfig(); err != nil {
			slog.Error("Failed to reload configuration", "error", err)
		} else {
			slog.Info("Configuration reloaded successfully")
		}

	case syscall.SIGUSR1:
		slog.Info("Received SIGUSR1, dumping statistics")
		sh.manager.DumpStatistics()

	case syscall.SIGUSR2:
		slog.Info("Received SIGUSR2, toggling debug mode")
		sh.manager.ToggleDebugMode()

	default:
		slog.Warn("Received unhandled signal", "signal", sig.String())
	}
}
