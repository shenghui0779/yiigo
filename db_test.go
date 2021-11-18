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
		ConnMaxLifetime: 10 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	}

	opt.rebuild(&DBOptions{
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: -1,
	})

	assert.Equal(t, &DBOptions{
		MaxOpenConns:    20,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
	}, opt)
}
