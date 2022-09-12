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

// httpSetting http request setting
type httpSetting struct {
	headers map[string]string
	cookies []*http.Cookie
	close   bool
}

// HTTPOption configures how we set up the http request.
type HTTPOption func(s *httpSetting)

// WithHTTPHeader specifies the header to http request.
func WithHTTPHeader(key, value string) HTTPOption {
	return func(s *httpSetting) {
		s.headers[key] = value
	}
}

// WithHTTPCookies specifies the cookies to http request.
func WithHTTPCookies(cookies ...*http.Cookie) HTTPOption {
	return func(s *httpSetting) {
		s.cookies = cookies
	}
}

// WithHTTPClose specifies close the connection after
// replying to this request (for servers) or after sending this
// request and reading its response (for clients).
func WithHTTPClose() HTTPOption {
	return func(s *httpSetting) {
		s.close = true
	}
}

// UploadForm is the interface for http upload.
type UploadForm interface {
	// Write writes fields to multipart writer
	Write(w *multipart.Writer) error
}

// FormFileFunc writes file content to multipart writer.
type FormFileFunc func(w io.Writer) error

type formfile struct {
	fieldname string
	filename  string
	filefunc  FormFileFunc
}

type uploadform struct {
	formfiles  []*formfile
	formfields map[string]string
}

func (f *uploadform) Write(w *multipart.Writer) error {
	if len(f.formfiles) == 0 {
		return errors.New("empty file field")
	}

	for _, v := range f.formfiles {
		part, err := w.CreateFormFile(v.fieldname, v.filename)

		if err != nil {
			return err
		}

		if err = v.filefunc(part); err != nil {
			return err
		}
	}

	for name, value := range f.formfields {
		if err := w.WriteField(name, value); err != nil {
			return err
		}
	}

	return nil
}

// UploadField configures how we set up the upload from.
type UploadField func(f *uploadform)

// WithFormFile specifies the file field to upload from.
func WithFormFile(fieldname, filename string, fn FormFileFunc) UploadField {
	return func(f *uploadform) {
		f.formfiles = append(f.formfiles, &formfile{
			fieldname: fieldname,
			filename:  filename,
			filefunc:  fn,
		})
	}
}

// WithFormField specifies the form field to upload from.
func WithFormField(fieldname, fieldvalue string) UploadField {
	return func(u *uploadform) {
		u.formfields[fieldname] = fieldvalue
	}
}

// NewUploadForm returns an upload form
func NewUploadForm(fields ...UploadField) UploadForm {
	form := &uploadform{
		formfiles:  make([]*formfile, 0),
		formfields: make(map[string]string),
	}

	for _, f := range fields {
		f(form)
	}

	return form
}

// HTTPClient is the interface for a http client.
type HTTPClient interface {
	// Do sends an HTTP request and returns an HTTP response.
	// Should use context to specify the timeout for request.
	Do(ctx context.Context, method, reqURL string, body []byte, options ...HTTPOption) (*http.Response, error)

	// Upload issues a UPLOAD to the specified URL.
	// Should use context to specify the timeout for request.
	Upload(ctx context.Context, reqURL string, form UploadForm, options ...HTTPOption) (*http.Response, error)
}

type httpclient struct {
	client *http.Client
}

func (c *httpclient) Do(ctx context.Context, method, reqURL string, body []byte, options ...HTTPOption) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, reqURL, bytes.NewBuffer(body))

	if err != nil {
		return nil, err
	}

	setting := new(httpSetting)

	if len(options) != 0 {
		setting.headers = make(map[string]string)

		for _, f := range options {
			f(setting)
		}
	}

	// headers
	if len(setting.headers) != 0 {
		for k, v := range setting.headers {
			req.Header.Set(k, v)
		}
	}

	// cookies
	if len(setting.cookies) != 0 {
		for _, v := range setting.cookies {
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

func (c *httpclient) Upload(ctx context.Context, reqURL string, form UploadForm, options ...HTTPOption) (*http.Response, error) {
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

// NewDefaultHTTPClient returns a new client with default http.Client
func NewDefaultHTTPClient() HTTPClient {
	return &httpclient{
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

// NewHTTPClient returns a new client with http.Client
func NewHTTPClient(client *http.Client) HTTPClient {
	return &httpclient{
		client: client,
	}
}

var defaultHTTPClient = NewDefaultHTTPClient()

// HTTPGet issues a GET to the specified URL.
func HTTPGet(ctx context.Context, reqURL string, options ...HTTPOption) (*http.Response, error) {
	return defaultHTTPClient.Do(ctx, http.MethodGet, reqURL, nil, options...)
}

// HTTPPost issues a POST to the specified URL.
func HTTPPost(ctx context.Context, reqURL string, body []byte, options ...HTTPOption) (*http.Response, error) {
	return defaultHTTPClient.Do(ctx, http.MethodPost, reqURL, body, options...)
}

// HTTPPostForm issues a POST to the specified URL, with data's keys and values URL-encoded as the request body.
func HTTPPostForm(ctx context.Context, reqURL string, data url.Values, options ...HTTPOption) (*http.Response, error) {
	options = append(options, WithHTTPHeader("Content-Type", "application/x-www-form-urlencoded"))

	return defaultHTTPClient.Do(ctx, http.MethodPost, reqURL, []byte(data.Encode()), options...)
}

// HTTPUpload issues a UPLOAD to the specified URL.
func HTTPUpload(ctx context.Context, reqURL string, form UploadForm, options ...HTTPOption) (*http.Response, error) {
	return defaultHTTPClient.Upload(ctx, reqURL, form, options...)
}

// HTTPDo sends an HTTP request and returns an HTTP response
func HTTPDo(ctx context.Context, method, reqURL string, body []byte, options ...HTTPOption) (*http.Response, error) {
	return defaultHTTPClient.Do(ctx, method, reqURL, body, options...)
}
