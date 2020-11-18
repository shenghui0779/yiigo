package yiigo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHTTPOption(t *testing.T) {
	options := []HTTPOption{
		WithHTTPHeader("Accept-Language", "zh-CN,zh;q=0.9"),
		WithHTTPHeader("Content-Type", "text/xml; charset=utf-8"),
		WithHTTPClose(),
		WithHTTPTimeout(5 * time.Second),
	}

	o := &httpOptions{
		headers: make(map[string]string),
		timeout: defaultHTTPTimeout,
	}

	for _, option := range options {
		option.apply(o)
	}

	assert.Equal(t, map[string]string{
		"Accept-Language": "zh-CN,zh;q=0.9",
		"Content-Type":    "text/xml; charset=utf-8",
	}, o.headers)
	assert.True(t, o.close)
	assert.Equal(t, 5*time.Second, o.timeout)
}
