package yiigo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoggerOption(t *testing.T) {
	setting := new(loggerSetting)

	options := []LoggerOption{
		WithLogMaxSize(100),
		WithLogMaxBackups(10),
		WithLogMaxAge(30),
		WithLogCompress(),
		WithLogStdErr(),
	}

	for _, f := range options {
		f(setting)
	}

	assert.Equal(t, &loggerSetting{
		maxSize:    100,
		maxBackups: 10,
		maxAge:     30,
		compress:   true,
		stderr:     true,
	}, setting)
}
