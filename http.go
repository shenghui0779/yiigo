package yiigo

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"path/filepath"
	"time"
)

// defaultHTTPTimeout default http request timeout
const defaultHTTPTimeout = 10 * time.Second

// httpOptions http request options
type httpOptions struct {
	headers map[string]string
	cookies []*http.Cookie
	close   bool
	timeout time.Duration
}

// HTTPOption configures how we set up the http request
type HTTPOption interface {
	apply(*httpOptions)
}

// funcHTTPOption implements request option
type funcHTTPOption struct {
	f func(*httpOptions)
}

func (fo *funcHTTPOption) apply(o *httpOptions) {
	fo.f(o)
}

func newFuncHTTPOption(f func(*httpOptions)) *funcHTTPOption {
	return &funcHTTPOption{f: f}
}

// WithHTTPHeader specifies the header to http request.
func WithHTTPHeader(key, value string) HTTPOption {
	return newFuncHTTPOption(func(o *httpOptions) {
		o.headers[key] = value
	})
}

// WithHTTPCookies specifies the cookies to http request.
func WithHTTPCookies(cookies ...*http.Cookie) HTTPOption {
	return newFuncHTTPOption(func(o *httpOptions) {
		o.cookies = cookies
	})
}

// WithHTTPClose specifies close the connection after
// replying to this request (for servers) or after sending this
// request and reading its response (for clients).
func WithHTTPClose() HTTPOption {
	return newFuncHTTPOption(func(o *httpOptions) {
		o.close = true
	})
}

// WithHTTPTimeout specifies the timeout to http request.
func WithHTTPTimeout(d time.Duration) HTTPOption {
	return newFuncHTTPOption(func(o *httpOptions) {
		o.timeout = d
	})
}

// UploadForm upload form
type UploadForm struct {
	fieldname   string
	filename    string
	body        func() ([]byte, error)
	extraFields map[string]string
}

// NewUploadForm returns new upload form
func NewUploadForm(fieldname, filename string, extraFields map[string]string) *UploadForm {
	return &UploadForm{
		fieldname: fieldname,
		filename:  filename,
		body: func() ([]byte, error) {
			path, err := filepath.Abs(filename)

			if err != nil {
				return nil, err
			}

			return ioutil.ReadFile(path)
		},
		extraFields: extraFields,
	}
}

// HTTPClient http client
type HTTPClient struct {
	client  *http.Client
	timeout time.Duration
}

func (h *HTTPClient) Do(ctx context.Context, req *http.Request, options ...HTTPOption) ([]byte, error) {
	o := &httpOptions{
		headers: make(map[string]string),
		timeout: h.timeout,
	}

	if len(options) > 0 {
		for _, option := range options {
			option.apply(o)
		}
	}

	// headers
	if len(o.headers) > 0 {
		for k, v := range o.headers {
			req.Header.Set(k, v)
		}
	}

	// cookies
	if len(o.cookies) > 0 {
		for _, v := range o.cookies {
			req.AddCookie(v)
		}
	}

	if o.close {
		req.Close = true
	}

	// timeout
	ctx, cancel := context.WithTimeout(ctx, o.timeout)

	defer cancel()

	resp, err := h.client.Do(req.WithContext(ctx))

	if err != nil {
		// If the context has been canceled, the context's error is probably more useful.
		select {
		case <-ctx.Done():
			err = ctx.Err()
		default:
		}

		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		io.Copy(ioutil.Discard, resp.Body)

		return nil, fmt.Errorf("error http code: %d", resp.StatusCode)
	}

	b, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	return b, nil
}

// Get http get request
func (h *HTTPClient) Get(ctx context.Context, url string, options ...HTTPOption) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, err
	}

	return h.Do(ctx, req, options...)
}

// Post http post request
func (h *HTTPClient) Post(ctx context.Context, url string, body []byte, options ...HTTPOption) ([]byte, error) {
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))

	if err != nil {
		return nil, err
	}

	return h.Do(ctx, req, options...)
}

// Upload http upload media
func (h *HTTPClient) Upload(ctx context.Context, url string, form *UploadForm, options ...HTTPOption) ([]byte, error) {
	media, err := form.body()

	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)

	fw, err := w.CreateFormFile(form.fieldname, form.filename)

	if err != nil {
		return nil, err
	}

	if _, err = io.Copy(fw, bytes.NewReader(media)); err != nil {
		return nil, err
	}

	// add extra fields
	if len(form.extraFields) != 0 {
		for k, v := range form.extraFields {
			w.WriteField(k, v)
		}
	}

	options = append(options, WithHTTPHeader("Content-Type", w.FormDataContentType()))

	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	w.Close()

	req, err := http.NewRequest("POST", url, bytes.NewReader([]byte(buf.String())))

	if err != nil {
		return nil, err
	}

	return h.Do(ctx, req, options...)
}

// defaultHTTPClient default http client
var defaultHTTPClient = &HTTPClient{
	client: &http.Client{
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
	},
	timeout: defaultHTTPTimeout,
}

// NewHTTPClient returns a new http client
func NewHTTPClient(client *http.Client, defaultTimeout ...time.Duration) *HTTPClient {
	c := &HTTPClient{
		client:  client,
		timeout: defaultHTTPTimeout,
	}

	if len(defaultTimeout) != 0 {
		c.timeout = defaultTimeout[0]
	}

	return c
}

// HTTPGet http get request
func HTTPGet(ctx context.Context, url string, options ...HTTPOption) ([]byte, error) {
	return defaultHTTPClient.Get(ctx, url, options...)
}

// HTTPPost http post request
func HTTPPost(ctx context.Context, url string, body []byte, options ...HTTPOption) ([]byte, error) {
	return defaultHTTPClient.Post(ctx, url, body, options...)
}

// HTTPUpload http upload media
func HTTPUpload(ctx context.Context, url string, form *UploadForm, options ...HTTPOption) ([]byte, error) {
	return defaultHTTPClient.Upload(ctx, url, form, options...)
}
