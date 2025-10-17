package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.toml")

	// Create a sample config file
	config := Config{
		DB:                filepath.Join(tempDir, "test.db"),
		GeminiApiKey:      "test-key",
		SpeakingRate:      1.5,
		SaveAudioData:     true,
		RequireConfirm:    true,
		UseGeminiTTS:      true,
		EnableEnvOverride: true,
	}

	// Write the config to the temporary file
	file, err := os.Create(configPath)
	require.NoError(t, err)
	defer func() { _ = file.Close() }()

	encoder := toml.NewEncoder(file)
	err = encoder.Encode(config)
	require.NoError(t, err)

	// Test loading the config
	loadedConfig, err := LoadConfig(configPath)
	require.NoError(t, err)

	assert.Equal(t, config.DB, loadedConfig.DB)
	assert.Equal(t, config.GeminiApiKey, loadedConfig.GeminiApiKey)
	assert.Equal(t, config.SpeakingRate, loadedConfig.SpeakingRate)
	assert.Equal(t, config.SaveAudioData, loadedConfig.SaveAudioData)
	assert.Equal(t, config.RequireConfirm, loadedConfig.RequireConfirm)
	assert.Equal(t, config.UseGeminiTTS, loadedConfig.UseGeminiTTS)
	assert.Equal(t, config.EnableEnvOverride, loadedConfig.EnableEnvOverride)
	assert.Equal(t, configPath, loadedConfig.SourcePath)
}

func TestLoadConfig_WithDefaults(t *testing.T) {
	// Create a temporary config file with minimal settings
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.toml")

	// Create a minimal config file
	minimalConfig := map[string]interface{}{
		"db": "test.db",
	}

	file, err := os.Create(configPath)
	require.NoError(t, err)
	defer func() { _ = file.Close() }()

	encoder := toml.NewEncoder(file)
	err = encoder.Encode(minimalConfig)
	require.NoError(t, err)

	// Test loading the config
	loadedConfig, err := LoadConfig(configPath)
	require.NoError(t, err)

	// Check default SpeakingRate
	assert.Equal(t, 1.3, loadedConfig.SpeakingRate)

	// Check that SourcePath is set correctly
	assert.Equal(t, configPath, loadedConfig.SourcePath)
}

func TestLoadConfig_WithInvalidPath(t *testing.T) {
	_, err := LoadConfig("/non/existent/path.toml")
	assert.Error(t, err)
}
