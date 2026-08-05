package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		cfg, err := LoadConfig()
		require.NoError(t, err)
		require.NotNil(t, cfg)
		assert.NotEmpty(t, cfg.Server.StreamServiceAddr)
		assert.NotEmpty(t, cfg.Redis.Addr)
		assert.NotEmpty(t, cfg.MinIO.Endpoint)
		assert.NotEmpty(t, cfg.MinIO.AccessKeyID)
		assert.NotEmpty(t, cfg.MinIO.SecretAccessKey)
		assert.NotEmpty(t, cfg.MinIO.BucketName)
		assert.NotEmpty(t, cfg.MinIO.Region)
	})

	t.Run("reads from env", func(t *testing.T) {
		t.Setenv("REDIS_ADDR", "redis:6379")
		t.Setenv("MINIO_BUCKET_NAME", "custom-bucket")
		t.Setenv("WORKER_CONCURRENCY", "4")

		cfg, err := LoadConfig()
		require.NoError(t, err)
		assert.Equal(t, "redis:6379", cfg.Redis.Addr)
		assert.Equal(t, "custom-bucket", cfg.MinIO.BucketName)
		assert.Equal(t, 4, cfg.Worker.Concurrency)
	})
}
