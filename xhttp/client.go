package xhttp

import (
	"bytes"
	"context"
	"crypto/tls"
	"mime/multipart"
	"net"
	"net/http"
	"time"
)

// Client HTTP客户端
type Client interface {
	// Do 发送HTTP请求
	// 注意：应该使用Context设置请求超时时间
	Do(ctx context.Context, method, reqURL string, body []byte, opts ...Option) (*http.Response, error)
	// Upload 上传文件
	// 注意：应该使用Context设置请求超时时间
	Upload(ctx context.Context, reqURL string, form UploadForm, opts ...Option) (*http.Response, error)
}

type client struct {
	cli *http.Client
}

func (c *client) Do(ctx context.Context, method, reqURL string, body []byte, opts ...Option) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, reqURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	o := new(options)
	if len(opts) != 0 {
		o.header = http.Header{}
		for _, fn := range opts {
			fn(o)
		}
	}

	// header
	if len(o.header) != 0 {
		req.Header = o.header
	}
	// cookie
	if len(o.cookie) != 0 {
		for _, v := range o.cookie {
			req.AddCookie(v)
		}
	}
	// close the connection after this request
	if o.close {
		req.Close = true
	}

	resp, err := c.cli.Do(req)
	if err != nil {
		// If the context has been canceled, the context's error is probably more useful.
		select {
		case <-ctx.Done():
			err = ctx.Err()
		default:
		}
		return nil, err
	}

	return resp, nil
}

func (c *client) Upload(ctx context.Context, reqURL string, form UploadForm, options ...Option) (*http.Response, error) {
	buf := bytes.NewBuffer(make([]byte, 0, 20<<10)) // 20kb
	w := multipart.NewWriter(buf)
	if err := form.Write(w); err != nil {
		return nil, err
	}

	options = append(options, WithHeader("Content-Type", w.FormDataContentType()))
	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	if err := w.Close(); err != nil {
		return nil, err
	}
	return c.Do(ctx, http.MethodPost, reqURL, buf.Bytes(), options...)
}

// NewDefaultClient 生成一个默认的HTTP客户端
func NewDefaultClient() Client {
	return &client{
		cli: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 60 * time.Second,
				}).DialContext,
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
				MaxIdleConns:          0,
				MaxIdleConnsPerHost:   1000,
				MaxConnsPerHost:       1000,
				IdleConnTimeout:       60 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: time.Second,
			},
		},
	}
}

// NewClient 通过官方 `http.Client` 生成一个HTTP客户端
func NewClient(c *http.Client) Client {
	return &client{
		cli: c,
	}
}
