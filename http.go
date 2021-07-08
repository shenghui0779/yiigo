package yiigo

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// defaultHTTPTimeout default http request timeout
const defaultHTTPTimeout = 10 * time.Second

// httpSettings http request settings
type httpSettings struct {
	headers map[string]string
	cookies []*http.Cookie
	close   bool
	timeout time.Duration
}

// HTTPOption configures how we set up the http request.
type HTTPOption func(s *httpSettings)

// WithHTTPHeader specifies the header to http request.
func WithHTTPHeader(key, value string) HTTPOption {
	return func(s *httpSettings) {
		s.headers[key] = value
	}
}

// WithHTTPCookies specifies the cookies to http request.
func WithHTTPCookies(cookies ...*http.Cookie) HTTPOption {
	return func(s *httpSettings) {
		s.cookies = cookies
	}
}

// WithHTTPClose specifies close the connection after
// replying to this request (for servers) or after sending this
// request and reading its response (for clients).
func WithHTTPClose() HTTPOption {
	return func(s *httpSettings) {
		s.close = true
	}
}

// WithHTTPTimeout specifies the timeout to http request.
func WithHTTPTimeout(timeout time.Duration) HTTPOption {
	return func(s *httpSettings) {
		s.timeout = timeout
	}
}

type uploadmethod int

const (
	uploadbybyptes uploadmethod = iota
	uploadbypath
	uploadbyurl
)

// HTTPClient is the interface for an http client.
type HTTPClient interface {
	// Do sends an HTTP request and returns an HTTP response.
	Do(ctx context.Context, req *http.Request, options ...HTTPOption) (*http.Response, error)

	// Upload issues a UPLOAD to the specified URL.
	Upload(ctx context.Context, reqURL string, form UploadForm, options ...HTTPOption) (*http.Response, error)
}

type httpclient struct {
	client  *http.Client
	timeout time.Duration
}

func (c *httpclient) Do(ctx context.Context, req *http.Request, options ...HTTPOption) (*http.Response, error) {
	settings := &httpSettings{timeout: c.timeout}

	if len(options) != 0 {
		settings.headers = make(map[string]string)

		for _, f := range options {
			f(settings)
		}
	}

	// headers
	if len(settings.headers) != 0 {
		for k, v := range settings.headers {
			req.Header.Set(k, v)
		}
	}

	// cookies
	if len(settings.cookies) != 0 {
		for _, v := range settings.cookies {
			req.AddCookie(v)
		}
	}

	if settings.close {
		req.Close = true
	}

	// timeout
	ctx, cancel := context.WithTimeout(ctx, settings.timeout)

	defer cancel()

	resp, err := c.client.Do(req.WithContext(ctx))

	if err != nil {
		// If the context has been canceled, the context's error is probably more useful.
		select {
		case <-ctx.Done():
			err = ctx.Err()
		default:
		}

		return nil, err
	}

	return resp, err
}

func (c *httpclient) Upload(ctx context.Context, reqURL string, form UploadForm, options ...HTTPOption) (*http.Response, error) {
	buf := bytes.NewBuffer(make([]byte, 0, 4<<10)) // 4kb
	w := multipart.NewWriter(buf)

	if err := form.Write(w); err != nil {
		return nil, err
	}

	options = append(options, WithHTTPHeader("Content-Type", w.FormDataContentType()))

	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	w.Close()

	req, err := http.NewRequest(http.MethodPost, reqURL, buf)

	if err != nil {
		return nil, err
	}

	return c.Do(ctx, req, options...)
}

// NewHTTPClient returns a new http client
func NewHTTPClient(client *http.Client, defaultTimeout ...time.Duration) HTTPClient {
	c := &httpclient{
		client:  client,
		timeout: defaultHTTPTimeout,
	}

	if len(defaultTimeout) != 0 {
		c.timeout = defaultTimeout[0]
	}

	return c
}

// UploadForm is the interface for http upload
type UploadForm interface {
	// Write writes fields to multipart writer
	Write(w *multipart.Writer) error
}

type uploadFileField struct {
	fieldname string
	filename  string
	body      []byte
}

type httpUpload struct {
	filefield []*uploadFileField
	formfield map[string]string
}

func (u *httpUpload) Write(w *multipart.Writer) error {
	if len(u.filefield) == 0 {
		return errors.New("yiigo: empty file field")
	}

	for _, field := range u.filefield {
		part, err := w.CreateFormFile(field.fieldname, field.filename)

		if err != nil {
			return err
		}

		if _, err = part.Write(field.body); err != nil {
			return err
		}
	}

	for name, value := range u.formfield {
		if err := w.WriteField(name, value); err != nil {
			return err
		}
	}

	return nil
}

// UploadOption configures how we set up the upload from.
type UploadOption func(u *httpUpload)

// WithFieldField specifies the file field to upload from.
func WithFieldField(fieldname, filename string, body []byte) UploadOption {
	return func(u *httpUpload) {
		u.filefield = append(u.filefield, &uploadFileField{
			fieldname: fieldname,
			filename:  filename,
			body:      body,
		})
	}
}

// WithFormField specifies the form field to upload from.
func WithFormField(fieldname, fieldvalue string) UploadOption {
	return func(u *httpUpload) {
		u.formfield[fieldname] = fieldvalue
	}
}

// NewUploadForm returns an upload form
func NewUploadForm(options ...UploadOption) UploadForm {
	form := &httpUpload{
		filefield: make([]*uploadFileField, 0),
		formfield: make(map[string]string),
	}

	for _, f := range options {
		f(form)
	}

	return form
}

// defaultHTTPClient default http client
var defaultHTTPClient = NewHTTPClient(&http.Client{
	Transport: &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 60 * time.Second,
		}).DialContext,
		MaxIdleConns:          0,
		MaxIdleConnsPerHost:   1000,
		MaxConnsPerHost:       1000,
		IdleConnTimeout:       60 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	},
}, defaultHTTPTimeout)

// HTTPGet issues a GET to the specified URL.
func HTTPGet(ctx context.Context, reqURL string, options ...HTTPOption) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)

	if err != nil {
		return nil, err
	}

	return defaultHTTPClient.Do(ctx, req, options...)
}

// HTTPPost issues a POST to the specified URL.
func HTTPPost(ctx context.Context, reqURL string, body io.Reader, options ...HTTPOption) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, reqURL, body)

	if err != nil {
		return nil, err
	}

	return defaultHTTPClient.Do(ctx, req, options...)
}

// HTTPPostForm issues a POST to the specified URL, with data's keys and values URL-encoded as the request body.
func HTTPPostForm(ctx context.Context, reqURL string, data url.Values, options ...HTTPOption) (*http.Response, error) {
	options = append(options, WithHTTPHeader("Content-Type", "application/x-www-form-urlencoded"))

	req, err := http.NewRequest(http.MethodPost, reqURL, strings.NewReader(data.Encode()))

	if err != nil {
		return nil, err
	}

	return defaultHTTPClient.Do(ctx, req, options...)
}

// HTTPUpload issues a UPLOAD to the specified URL.
func HTTPUpload(ctx context.Context, reqURL string, form UploadForm, options ...HTTPOption) (*http.Response, error) {
	return defaultHTTPClient.Upload(ctx, reqURL, form, options...)
}

// HTTPDo sends an HTTP request and returns an HTTP response
func HTTPDo(ctx context.Context, req *http.Request, options ...HTTPOption) (*http.Response, error) {
	return defaultHTTPClient.Do(ctx, req, options...)
}
