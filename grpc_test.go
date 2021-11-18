package yiigo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPoolOption(t *testing.T) {
	opt := &PoolOptions{
		PoolSize:    10,
		IdleTimeout: 5 * time.Minute,
	}

	opt.rebuild(&PoolOptions{
		PoolPrefill: 1,
		IdleTimeout: -1,
	})

	assert.Equal(t, &PoolOptions{
		PoolSize:    10,
		PoolPrefill: 1,
	}, opt)
}
