package yiigo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDBOptions(t *testing.T) {
	opt := &DBOptions{
		MaxOpenConns:    20,
		MaxIdleConns:    10,
		ConnMaxLifetime: 60 * time.Second,
		ConnMaxIdleTime: 5 * time.Minute,
	}

	opt.rebuild(&DBOptions{
		MaxIdleConns:    5,
		ConnMaxLifetime: -1,
		ConnMaxIdleTime: 60 * time.Second,
	})

	assert.Equal(t, &DBOptions{
		MaxOpenConns:    20,
		MaxIdleConns:    5,
		ConnMaxLifetime: 0,
		ConnMaxIdleTime: 60 * time.Second,
	}, opt)
}
