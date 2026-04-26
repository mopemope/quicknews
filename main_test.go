package main

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRun_ReturnsErrorWhenConfigMissing(t *testing.T) {
	missingConfig := filepath.Join(t.TempDir(), "missing.toml")

	err := run([]string{"--config", missingConfig, "config", "--format", "json"})
	require.Error(t, err)
	require.ErrorContains(t, err, "failed to decode config file")
}
