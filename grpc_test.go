package yiigo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPoolOption(t *testing.T) {
	setting := new(poolSetting)

	options := []PoolOption{
		WithPoolSize(10),
		WithPoolLimit(20),
		WithPoolIdleTimeout(60 * time.Second),
		WithPoolPrefill(2),
	}

	for _, f := range options {
		f(setting)
	}

	assert.Equal(t, &poolSetting{
		size:        10,
		limit:       20,
		idleTimeout: 60 * time.Second,
		prefill:     2,
	}, setting)
}
