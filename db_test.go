package yiigo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDBOption(t *testing.T) {
	setting := new(dbSetting)

	options := []DBOption{
		WithDBMaxOpenConns(20),
		WithDBMaxIdleConns(10),
		WithDBConnMaxIdleTime(60 * time.Second),
		WithDBConnMaxLifetime(10 * time.Minute),
	}

	for _, f := range options {
		f(setting)
	}

	assert.Equal(t, &dbSetting{
		maxOpenConns:    20,
		maxIdleConns:    10,
		connMaxIdleTime: 60 * time.Second,
		connMaxLifetime: 10 * time.Minute,
	}, setting)
}
