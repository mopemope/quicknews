package log

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitializeLogger_WithFile(t *testing.T) {
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "test.log")

	err := InitializeLogger(logPath, false)
	require.NoError(t, err)

	// Check if the file was created
	_, err = os.Stat(logPath)
	assert.NoError(t, err)
}

func TestInitializeLogger_WithInvalidPath(t *testing.T) {
	// Try to initialize logger with a path that's not writable
	invalidPath := "/invalid/path/log.txt"

	err := InitializeLogger(invalidPath, false)
	assert.Error(t, err)
}

func TestInitializeLogger_WithDebug(t *testing.T) {
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "test_debug.log")

	err := InitializeLogger(logPath, true)
	require.NoError(t, err)

	// Check if the file was created
	_, err = os.Stat(logPath)
	assert.NoError(t, err)
}

func TestInitializeLogger_WithEmptyPath(t *testing.T) {
	// Test with empty path (should use stdout)
	err := InitializeLogger("", false)
	require.NoError(t, err)
}
