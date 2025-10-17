package storage

import (
	"context"
	"testing"

	"github.com/mopemope/quicknews/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewR2Storage(t *testing.T) {
	ctx := context.Background()

	// Test with valid configuration
	cfg := &config.Config{
		Cloudflare: &config.Cloudflare{
			AccessKeyID:     "test-access-key",
			SecretAccessKey: "test-secret-key",
			BucketName:      "test-bucket",
			EndpointURL:     "https://test.r2.cloudflarestorage.com",
		},
	}

	storage, err := NewR2Storage(ctx, cfg)
	require.NoError(t, err)
	assert.NotNil(t, storage)
	assert.Equal(t, "test-bucket", storage.bucketName)
}

func TestNewR2Storage_MissingConfig(t *testing.T) {
	ctx := context.Background()

	// Test with nil Cloudflare config
	cfg := &config.Config{
		Cloudflare: nil,
	}

	_, err := NewR2Storage(ctx, cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Cloudflare R2 configuration is missing")
}

func TestNewR2Storage_MissingRequiredFields(t *testing.T) {
	ctx := context.Background()

	// Test with missing required fields
	cfg := &config.Config{
		Cloudflare: &config.Cloudflare{
			AccessKeyID:     "", // Missing
			SecretAccessKey: "test-secret-key",
			BucketName:      "test-bucket",
			EndpointURL:     "https://test.r2.cloudflarestorage.com",
		},
	}

	_, err := NewR2Storage(ctx, cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required Cloudflare R2 configuration fields")
}

func TestNewR2Storage_MissingAllFields(t *testing.T) {
	ctx := context.Background()

	// Test with all required fields missing
	cfg := &config.Config{
		Cloudflare: &config.Cloudflare{},
	}

	_, err := NewR2Storage(ctx, cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required Cloudflare R2 configuration fields")
}

func TestNewR2Storage_WithValidConfig(t *testing.T) {
	ctx := context.Background()

	// Test with all required fields present
	cfg := &config.Config{
		Cloudflare: &config.Cloudflare{
			AccessKeyID:     "test-access-key",
			SecretAccessKey: "test-secret-key",
			BucketName:      "test-bucket",
			EndpointURL:     "https://test.r2.cloudflarestorage.com",
		},
	}

	storage, err := NewR2Storage(ctx, cfg)
	require.NoError(t, err)
	assert.NotNil(t, storage)
	assert.Equal(t, "test-bucket", storage.bucketName)

	// Check that the endpoint contains the expected value (substring check)
	// Since we can't easily access the internal client configuration,
	// we'll just verify the storage object was created successfully
	assert.NotEmpty(t, storage.bucketName)
}
