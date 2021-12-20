package yiigo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRedisOption(t *testing.T) {
	opt := &RedisOptions{
		ConnTimeout:  10 * time.Second,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		PoolSize:     10,
		IdleTimeout:  5 * time.Minute,
	}

	opt.rebuild(&RedisOptions{
		Username:    "root",
		Password:    "root@123456",
		Database:    9,
		ConnTimeout: -1,
		IdleTimeout: 60 * time.Second,
	})

	assert.Equal(t, &RedisOptions{
		Username:     "root",
		Password:     "root@123456",
		Database:     9,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		PoolSize:     10,
		IdleTimeout:  60 * time.Second,
	}, opt)
}
