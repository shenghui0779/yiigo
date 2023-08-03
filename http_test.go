package yiigo

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPOption(t *testing.T) {
	setting := &httpSetting{
		header: http.Header{},
		close:  false,
	}

	options := []HTTPOption{
		WithHTTPHeader("Content-Type", "application/x-www-form-urlencoded"),
		WithHTTPClose(),
	}

	for _, f := range options {
		f(setting)
	}

	assert.Equal(t, &httpSetting{
		header: http.Header{
			"Content-Type": []string{"application/x-www-form-urlencoded"},
		},
		close: true,
	}, setting)
}
