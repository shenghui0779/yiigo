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

	settings := &httpSettings{
		headers: make(map[string]string),
		timeout: defaultHTTPTimeout,
	}

	for _, f := range options {
		f(settings)
	}

	assert.Equal(t, map[string]string{
		"Accept-Language": "zh-CN,zh;q=0.9",
		"Content-Type":    "text/xml; charset=utf-8",
	}, settings.headers)
	assert.True(t, settings.close)
	assert.Equal(t, 5*time.Second, settings.timeout)
}

// func TestUploadOption(t *testing.T) {
// 	options := []UploadOption{
// 		WithResourceURL("https://img.test.com/test.jpg"),
// 		WithExtraField("title", "TITLE"),
// 		WithExtraField("introduction", "INTRODUCTION"),
// 	}

// 	upload := &httpUpload{extraFields: make(map[string]string)}

// 	for _, f := range options {
// 		f(upload)
// 	}

// 	assert.Equal(t, "https://img.test.com/test.jpg", upload.resourceURL)
// 	assert.Equal(t, map[string]string{
// 		"title":        "TITLE",
// 		"introduction": "INTRODUCTION",
// 	}, upload.extraFields)
// }
