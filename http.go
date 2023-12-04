package yiigo

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"time"
)

type httpSetting struct {
	header http.Header
	cookie []*http.Cookie
	close  bool
}

// HTTPOption HTTP请求选项
type HTTPOption func(s *httpSetting)

// WithHTTPHeader 设置HTTP请求头
func WithHTTPHeader(key string, vals ...string) HTTPOption {
	return func(s *httpSetting) {
		if len(vals) == 1 {
			s.header.Set(key, vals[0])

			return
		}

		for _, v := range vals {
			s.header.Add(key, v)
		}
	}
}

// WithHTTPCookies 设置HTTP请求Cookie
func WithHTTPCookies(cookies ...*http.Cookie) HTTPOption {
	return func(s *httpSetting) {
		s.cookie = cookies
	}
}

// WithHTTPClose 请求结束后关闭请求
func WithHTTPClose() HTTPOption {
	return func(s *httpSetting) {
		s.close = true
	}
}

// UploadForm HTTP文件上传表单
type UploadForm interface {
	// Field 返回表单普通字段
	Field(name string) string

	// Write 将表单文件写入流
	Write(w *multipart.Writer) error
}

// FormFileFunc 将文件写入表单流
type FormFileFunc func(w io.Writer) error

type formfile struct {
	fieldname string
	filename  string
	filefunc  FormFileFunc
}

type uploadform struct {
	files  []*formfile
	fields V
}

func (form *uploadform) Field(name string) string {
	return form.fields.Get(name)
}

func (form *uploadform) Write(w *multipart.Writer) error {
	if len(form.files) == 0 {
		return errors.New("empty file field")
	}

	for _, v := range form.files {
		part, err := w.CreateFormFile(v.fieldname, v.filename)
		if err != nil {
			return err
		}

		if err = v.filefunc(part); err != nil {
			return err
		}
	}

	for name, value := range form.fields {
		if err := w.WriteField(name, value); err != nil {
			return err
		}
	}

	return nil
}

// UploadField 文件上传表单字段
type UploadField func(form *uploadform)

// WithFormFile 设置表单文件字段
func WithFormFile(fieldname, filename string, fn FormFileFunc) UploadField {
	return func(form *uploadform) {
		form.files = append(form.files, &formfile{
			fieldname: fieldname,
			filename:  filename,
			filefunc:  fn,
		})
	}
}

// WithFormField 设置表单普通字段
func WithFormField(name, value string) UploadField {
	return func(form *uploadform) {
		form.fields.Set(name, value)
	}
}

// NewUploadForm 生成一个文件上传表单
func NewUploadForm(fields ...UploadField) UploadForm {
	form := &uploadform{
		files:  make([]*formfile, 0),
		fields: make(V),
	}

	for _, f := range fields {
		f(form)
	}

	return form
}

// HTTPClient HTTP客户端
type HTTPClient interface {
	// Do 发送HTTP请求
	// 注意：应该使用Context设置请求超时时间
	Do(ctx context.Context, method, reqURL string, body []byte, options ...HTTPOption) (*http.Response, error)

	// Upload 上传文件
	// 注意：应该使用Context设置请求超时时间
	Upload(ctx context.Context, reqURL string, form UploadForm, options ...HTTPOption) (*http.Response, error)
}

type httpCli struct {
	client *http.Client
}

func (c *httpCli) Do(ctx context.Context, method, reqURL string, body []byte, options ...HTTPOption) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, reqURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	setting := new(httpSetting)

	if len(options) != 0 {
		setting.header = http.Header{}

		for _, f := range options {
			f(setting)
		}
	}

	// header
	if len(setting.header) != 0 {
		req.Header = setting.header
	}

	// cookie
	if len(setting.cookie) != 0 {
		for _, v := range setting.cookie {
			req.AddCookie(v)
		}
	}

	if setting.close {
		req.Close = true
	}

	resp, err := c.client.Do(req)
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

func (c *httpCli) Upload(ctx context.Context, reqURL string, form UploadForm, options ...HTTPOption) (*http.Response, error) {
	buf := bytes.NewBuffer(make([]byte, 0, 20<<10)) // 20kb
	w := multipart.NewWriter(buf)

	if err := form.Write(w); err != nil {
		return nil, err
	}

	options = append(options, WithHTTPHeader("Content-Type", w.FormDataContentType()))

	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	if err := w.Close(); err != nil {
		return nil, err
	}

	return c.Do(ctx, http.MethodPost, reqURL, buf.Bytes(), options...)
}

// NewDefaultHTTPClient 生成一个默认的HTTP客户端
func NewDefaultHTTPClient() HTTPClient {
	return &httpCli{
		client: &http.Client{
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

// NewHTTPClient 通过官方 `http.Client` 生成一个HTTP客户端
func NewHTTPClient(client *http.Client) HTTPClient {
	return &httpCli{
		client: client,
	}
}

var defaultHttpCli = NewDefaultHTTPClient()

// HTTPGet 发送GET请求
func HTTPGet(ctx context.Context, reqURL string, options ...HTTPOption) (*http.Response, error) {
	return defaultHttpCli.Do(ctx, http.MethodGet, reqURL, nil, options...)
}

// HTTPPost 发送POST请求
func HTTPPost(ctx context.Context, reqURL string, body []byte, options ...HTTPOption) (*http.Response, error) {
	return defaultHttpCli.Do(ctx, http.MethodPost, reqURL, body, options...)
}

// HTTPPostJSON 发送POST请求(json数据)
func HTTPPostJSON(ctx context.Context, reqURL string, body []byte, options ...HTTPOption) (*http.Response, error) {
	options = append(options, WithHTTPHeader(HeaderContentType, ContentForm))

	return defaultHttpCli.Do(ctx, http.MethodPost, reqURL, body, options...)
}

// HTTPPostForm 发送POST表单请求
func HTTPPostForm(ctx context.Context, reqURL string, data url.Values, options ...HTTPOption) (*http.Response, error) {
	options = append(options, WithHTTPHeader(HeaderContentType, ContentForm))

	return defaultHttpCli.Do(ctx, http.MethodPost, reqURL, []byte(data.Encode()), options...)
}

// HTTPUpload 文件上传
func HTTPUpload(ctx context.Context, reqURL string, form UploadForm, options ...HTTPOption) (*http.Response, error) {
	return defaultHttpCli.Upload(ctx, reqURL, form, options...)
}

// HTTPDo 发送HTTP请求
func HTTPDo(ctx context.Context, method, reqURL string, body []byte, options ...HTTPOption) (*http.Response, error) {
	return defaultHttpCli.Do(ctx, method, reqURL, body, options...)
}
