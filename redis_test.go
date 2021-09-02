package yiigo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRedisOption(t *testing.T) {
	setting := &redisSetting{
		pool: new(poolSetting),
	}

	options := []RedisOption{
		WithRedisDatabase(1),
		WithRedisConnTimeout(10 * time.Second),
		WithRedisReadTimeout(10 * time.Second),
		WithRedisWriteTimeout(10 * time.Second),
		WithRedisPool(
			WithPoolSize(10),
			WithPoolLimit(20),
			WithPoolIdleTimeout(60*time.Second),
			WithPoolPrefill(2),
		),
	}

	for _, f := range options {
		f(setting)
	}

	assert.Equal(t, &redisSetting{
		database:     1,
		connTimeout:  10 * time.Second,
		readTimeout:  10 * time.Second,
		writeTimeout: 10 * time.Second,
		pool: &poolSetting{
			size:        10,
			limit:       20,
			idleTimeout: 60 * time.Second,
			prefill:     2,
		},
	}, setting)
}
