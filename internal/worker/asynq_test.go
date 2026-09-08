package worker

import (
	"testing"
	"time"

	"github.com/mrhumster/thumbnail-service/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisOpt(t *testing.T) {
	opt := RedisOpt(config.Redis{
		Addr:     "redis:6379",
		Password: "secret",
		DB:       3,
	})
	assert.Equal(t, "redis:6379", opt.Addr)
	assert.Equal(t, "secret", opt.Password)
	assert.Equal(t, 3, opt.DB)
}

func TestAsynqConfig(t *testing.T) {
	cfg := &config.Config{
		Worker: config.Worker{
			Concurrency:     4,
			ShutdownTimeout: 30 * time.Second,
		},
	}
	asynqCfg := asynqConfig(cfg, nil)
	require.NotNil(t, asynqCfg.Queues)
	assert.Equal(t, 4, asynqCfg.Concurrency)
	assert.Equal(t, 30*time.Second, asynqCfg.ShutdownTimeout)
	assert.Contains(t, asynqCfg.Queues, "thumbsnails")
	assert.NotNil(t, asynqCfg.ErrorHandler)
}
