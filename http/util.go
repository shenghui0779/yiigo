package http

import (
	"context"
	"net/http"
	"net/url"
)

const MaxFormMemory = 32 << 20

const (
	HeaderAccept        = "Accept"
	HeaderAuthorization = "Authorization"
	HeaderContentType   = "Content-Type"
)

const (
	ContentText          = "text/plain;charset=utf-8"
	ContentJSON          = "application/json;charset=utf-8"
	ContentForm          = "application/x-www-form-urlencoded"
	ContentStream        = "application/octet-stream"
	ContentFormMultipart = "multipart/form-data"
)

var defaultCli = NewDefaultClient()

// Get 发送GET请求
func Get(ctx context.Context, reqURL string, options ...Option) (*http.Response, error) {
	return defaultCli.Do(ctx, http.MethodGet, reqURL, nil, options...)
}

// Post 发送POST请求
func Post(ctx context.Context, reqURL string, body []byte, options ...Option) (*http.Response, error) {
	return defaultCli.Do(ctx, http.MethodPost, reqURL, body, options...)
}

// PostJSON 发送POST请求(json数据)
func PostJSON(ctx context.Context, reqURL string, body []byte, options ...Option) (*http.Response, error) {
	options = append(options, WithHeader(HeaderContentType, ContentJSON))
	return defaultCli.Do(ctx, http.MethodPost, reqURL, body, options...)
}

// PostForm 发送POST表单请求
func PostForm(ctx context.Context, reqURL string, data url.Values, options ...Option) (*http.Response, error) {
	options = append(options, WithHeader(HeaderContentType, ContentForm))
	return defaultCli.Do(ctx, http.MethodPost, reqURL, []byte(data.Encode()), options...)
}

// Upload 文件上传
func Upload(ctx context.Context, reqURL string, form UploadForm, options ...Option) (*http.Response, error) {
	return defaultCli.Upload(ctx, reqURL, form, options...)
}

// Do 发送HTTP请求
func Do(ctx context.Context, method, reqURL string, body []byte, options ...Option) (*http.Response, error) {
	return defaultCli.Do(ctx, method, reqURL, body, options...)
}
