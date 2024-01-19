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

// HTTPGet 发送GET请求
func HTTPGet(ctx context.Context, reqURL string, options ...Option) (*http.Response, error) {
	return defaultCli.Do(ctx, http.MethodGet, reqURL, nil, options...)
}

// HTTPPost 发送POST请求
func HTTPPost(ctx context.Context, reqURL string, body []byte, options ...Option) (*http.Response, error) {
	return defaultCli.Do(ctx, http.MethodPost, reqURL, body, options...)
}

// HTTPPostJSON 发送POST请求(json数据)
func HTTPPostJSON(ctx context.Context, reqURL string, body []byte, options ...Option) (*http.Response, error) {
	options = append(options, WithHeader(HeaderContentType, ContentJSON))
	return defaultCli.Do(ctx, http.MethodPost, reqURL, body, options...)
}

// HTTPPostForm 发送POST表单请求
func HTTPPostForm(ctx context.Context, reqURL string, data url.Values, options ...Option) (*http.Response, error) {
	options = append(options, WithHeader(HeaderContentType, ContentForm))
	return defaultCli.Do(ctx, http.MethodPost, reqURL, []byte(data.Encode()), options...)
}

// HTTPUpload 文件上传
func HTTPUpload(ctx context.Context, reqURL string, form UploadForm, options ...Option) (*http.Response, error) {
	return defaultCli.Upload(ctx, reqURL, form, options...)
}

// HTTPDo 发送HTTP请求
func HTTPDo(ctx context.Context, method, reqURL string, body []byte, options ...Option) (*http.Response, error) {
	return defaultCli.Do(ctx, method, reqURL, body, options...)
}
