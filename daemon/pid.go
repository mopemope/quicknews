package daemon

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/cockroachdb/errors"
)

// PIDManager handles PID file operations for daemon processes
type PIDManager struct {
	pidFile string
}

// NewPIDManager creates a new PID manager with the specified PID file path
func NewPIDManager(pidFile string) *PIDManager {
	// Expand tilde to home directory
	if strings.HasPrefix(pidFile, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			pidFile = filepath.Join(home, pidFile[2:])
		}
	}
	
	return &PIDManager{
		pidFile: pidFile,
	}
}

// WritePID writes the current process PID to the PID file
func (pm *PIDManager) WritePID() error {
	pid := os.Getpid()
	
	// Create directory if it doesn't exist
	dir := filepath.Dir(pm.pidFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return errors.Wrap(err, "failed to create PID file directory")
	}
	
	// Write PID to file with restricted permissions
	pidStr := strconv.Itoa(pid)
	if err := os.WriteFile(pm.pidFile, []byte(pidStr), 0600); err != nil {
		return errors.Wrap(err, "failed to write PID file")
	}
	
	return nil
}

// ReadPID reads the PID from the PID file
func (pm *PIDManager) ReadPID() (int, error) {
	data, err := os.ReadFile(pm.pidFile)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, errors.New("PID file does not exist")
		}
		return 0, errors.Wrap(err, "failed to read PID file")
	}
	
	pidStr := strings.TrimSpace(string(data))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return 0, errors.Wrap(err, "invalid PID in file")
	}
	
	return pid, nil
}

// IsRunning checks if the process with the PID from the file is running
func (pm *PIDManager) IsRunning() (bool, int, error) {
	pid, err := pm.ReadPID()
	if err != nil {
		return false, 0, err
	}
	
	// Check if process exists by sending signal 0
	process, err := os.FindProcess(pid)
	if err != nil {
		return false, pid, errors.Wrap(err, "failed to find process")
	}
	
	err = process.Signal(syscall.Signal(0))
	if err != nil {
		// Process doesn't exist or we don't have permission
		return false, pid, nil
	}
	
	return true, pid, nil
}

// RemovePID removes the PID file
func (pm *PIDManager) RemovePID() error {
	if err := os.Remove(pm.pidFile); err != nil && !os.IsNotExist(err) {
		return errors.Wrap(err, "failed to remove PID file")
	}
	return nil
}

// GetPIDFile returns the PID file path
func (pm *PIDManager) GetPIDFile() string {
	return pm.pidFile
}

// KillProcess attempts to kill the process with the PID from the file
func (pm *PIDManager) KillProcess(signal os.Signal) error {
	pid, err := pm.ReadPID()
	if err != nil {
		return err
	}
	
	process, err := os.FindProcess(pid)
	if err != nil {
		return errors.Wrap(err, "failed to find process")
	}
	
	if err := process.Signal(signal); err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to send signal %v to process %d", signal, pid))
	}
	
	return nil
}
