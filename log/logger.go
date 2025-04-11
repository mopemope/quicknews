package log

import (
	"io"
	"log/slog"
	"os"
)

// InitializeLogger initializes the global slog logger.
// If logPath is empty, it logs to os.Stdout.
// Otherwise, it logs to the specified file path.
// If debug is true, the log level is set to Debug.
func InitializeLogger(logPath string, debug bool) error {
	var output io.Writer = os.Stdout
	var err error

	if logPath != "" {
		output, err = os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			// Fallback to stdout if file opening fails
			slog.Error("failed to open log file, falling back to stdout", "path", logPath, "error", err)
			// output = os.Stdout // This assignment is ineffectual because of the return below.
			// Return the error so the caller knows initialization partially failed
			return err
		}
		// Note: We don't defer file close here as the logger needs the file open for the application's lifetime.
		// The OS will close the file descriptor on process exit.
	}

	logLevel := slog.LevelInfo
	if debug {
		logLevel = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	logger := slog.New(slog.NewJSONHandler(output, opts))
	slog.SetDefault(logger)
	return nil
}
