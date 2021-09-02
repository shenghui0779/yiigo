package yiigo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPOption(t *testing.T) {
	setting := &httpSetting{
		headers: make(map[string]string),
		close:   false,
	}

	options := []HTTPOption{
		WithHTTPHeader("Content-Type", "application/x-www-form-urlencoded"),
		WithHTTPClose(),
	}

	for _, f := range options {
		f(setting)
	}

	assert.Equal(t, &httpSetting{
		headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		close: true,
	}, setting)
}
