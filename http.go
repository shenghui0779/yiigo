package yiigo

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

// defaultHTTPTimeout default http request timeout
const defaultHTTPTimeout = 10 * time.Second

// httpClientOptions http client options
type httpClientOptions struct {
	dialTimeout           time.Duration
	dialKeepAlive         time.Duration
	fallbackDelay         time.Duration
	maxIdleConns          int
	maxIdleConnsPerHost   int
	maxConnsPerHost       int
	idleConnTimeout       time.Duration
	tlsConfig             *tls.Config
	tlsHandshakeTimeout   time.Duration
	expectContinueTimeout time.Duration
	defaultTimeout        time.Duration
}

// HTTPClientOption configures how we set up the http client
type HTTPClientOption interface {
	apply(*httpClientOptions)
}

// funcHTTPClientOption implements http client option
type funcHTTPClientOption struct {
	f func(*httpClientOptions)
}

func (fo *funcHTTPClientOption) apply(o *httpClientOptions) {
	fo.f(o)
}

func newFuncHTTPOption(f func(*httpClientOptions)) *funcHTTPClientOption {
	return &funcHTTPClientOption{f: f}
}

// WithHTTPDialTimeout specifies the `DialTimeout` to net.Dialer.
func WithHTTPDialTimeout(d time.Duration) HTTPClientOption {
	return newFuncHTTPOption(func(o *httpClientOptions) {
		o.dialTimeout = d
	})
}

// WithHTTPDialKeepAlive specifies the `KeepAlive` to net.Dialer.
func WithHTTPDialKeepAlive(d time.Duration) HTTPClientOption {
	return newFuncHTTPOption(func(o *httpClientOptions) {
		o.dialKeepAlive = d
	})
}

// WithHTTPDialFallbackDelay specifies the `FallbackDelay` to net.Dialer.
func WithHTTPDialFallbackDelay(d time.Duration) HTTPClientOption {
	return newFuncHTTPOption(func(o *httpClientOptions) {
		o.fallbackDelay = d
	})
}

// WithHTTPMaxIdleConns specifies the `MaxIdleConns` to http client.
func WithHTTPMaxIdleConns(n int) HTTPClientOption {
	return newFuncHTTPOption(func(o *httpClientOptions) {
		o.maxIdleConns = n
	})
}

// WithHTTPMaxIdleConnsPerHost specifies the `MaxIdleConnsPerHost` to http client.
func WithHTTPMaxIdleConnsPerHost(n int) HTTPClientOption {
	return newFuncHTTPOption(func(o *httpClientOptions) {
		o.maxIdleConnsPerHost = n
	})
}

// WithHTTPMaxConnsPerHost specifies the `MaxConnsPerHost` to http client.
func WithHTTPMaxConnsPerHost(n int) HTTPClientOption {
	return newFuncHTTPOption(func(o *httpClientOptions) {
		o.maxConnsPerHost = n
	})
}

// WithHTTPIdleConnTimeout specifies the `IdleConnTimeout` to http client.
func WithHTTPIdleConnTimeout(d time.Duration) HTTPClientOption {
	return newFuncHTTPOption(func(o *httpClientOptions) {
		o.idleConnTimeout = d
	})
}

// WithHTTPTLSConfig specifies the `TLSClientConfig` to http client.
func WithHTTPTLSConfig(c *tls.Config) HTTPClientOption {
	return newFuncHTTPOption(func(o *httpClientOptions) {
		o.tlsConfig = c
	})
}

// WithHTTPTLSHandshakeTimeout specifies the `TLSHandshakeTimeout` to http client.
func WithHTTPTLSHandshakeTimeout(d time.Duration) HTTPClientOption {
	return newFuncHTTPOption(func(o *httpClientOptions) {
		o.tlsHandshakeTimeout = d
	})
}

// WithHTTPExpectContinueTimeout specifies the `ExpectContinueTimeout` to http client.
func WithHTTPExpectContinueTimeout(d time.Duration) HTTPClientOption {
	return newFuncHTTPOption(func(o *httpClientOptions) {
		o.expectContinueTimeout = d
	})
}

// WithHTTPDefaultTimeout specifies the `DefaultTimeout` to http client.
func WithHTTPDefaultTimeout(d time.Duration) HTTPClientOption {
	return newFuncHTTPOption(func(o *httpClientOptions) {
		o.defaultTimeout = d
	})
}

// httpRequestOptions http request options
type httpRequestOptions struct {
	headers        map[string]string
	cookies        []*http.Cookie
	close          bool
	timeout        time.Duration
	zipkinSpanTags map[string]string
}

// HTTPRequestOption configures how we set up the http request
type HTTPRequestOption interface {
	apply(*httpRequestOptions)
}

// funcHTTPRequestOption implements request option
type funcHTTPRequestOption struct {
	f func(*httpRequestOptions)
}

func (fo *funcHTTPRequestOption) apply(o *httpRequestOptions) {
	fo.f(o)
}

func newFuncHTTPRequestOption(f func(*httpRequestOptions)) *funcHTTPRequestOption {
	return &funcHTTPRequestOption{f: f}
}

// WithRequestHeader specifies the header to http request.
func WithRequestHeader(key, value string) HTTPRequestOption {
	return newFuncHTTPRequestOption(func(o *httpRequestOptions) {
		o.headers[key] = value
	})
}

// WithRequestCookies specifies the cookies to http request.
func WithRequestCookies(cookies ...*http.Cookie) HTTPRequestOption {
	return newFuncHTTPRequestOption(func(o *httpRequestOptions) {
		o.cookies = cookies
	})
}

// WithRequestClose specifies close the connection after
// replying to this request (for servers) or after sending this
// request and reading its response (for clients).
func WithRequestClose(b bool) HTTPRequestOption {
	return newFuncHTTPRequestOption(func(o *httpRequestOptions) {
		o.close = b
	})
}

// WithRequestTimeout specifies the timeout to http request.
func WithRequestTimeout(d time.Duration) HTTPRequestOption {
	return newFuncHTTPRequestOption(func(o *httpRequestOptions) {
		o.timeout = d
	})
}

// WithZipkinSpanTag specifies the zipkin span tag to zipkin http request.
func WithZipkinSpanTag(key, value string) HTTPRequestOption {
	return newFuncHTTPRequestOption(func(o *httpRequestOptions) {
		o.zipkinSpanTags[key] = value
	})
}

// HTTPClient http client
type HTTPClient struct {
	client  *http.Client
	timeout time.Duration
}

// Get http get request
func (h *HTTPClient) Get(url string, options ...HTTPRequestOption) ([]byte, error) {
	o := &httpRequestOptions{
		headers: make(map[string]string),
		timeout: h.timeout,
	}

	if len(options) > 0 {
		for _, option := range options {
			option.apply(o)
		}
	}

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, err
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
	ctx, cancel := context.WithTimeout(req.Context(), o.timeout)

	defer cancel()

	resp, err := h.client.Do(req.WithContext(ctx))

	if err != nil {
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

// Post http post request
func (h *HTTPClient) Post(url string, body []byte, options ...HTTPRequestOption) ([]byte, error) {
	o := &httpRequestOptions{
		headers: make(map[string]string),
		timeout: h.timeout,
	}

	if len(options) > 0 {
		for _, option := range options {
			option.apply(o)
		}
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))

	if err != nil {
		return nil, err
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
	ctx, cancel := context.WithTimeout(req.Context(), o.timeout)

	defer cancel()

	resp, err := h.client.Do(req.WithContext(ctx))

	if err != nil {
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
func NewHTTPClient(options ...HTTPClientOption) *HTTPClient {
	o := &httpClientOptions{
		dialTimeout:           30 * time.Second,
		dialKeepAlive:         60 * time.Second,
		maxIdleConns:          0,
		maxIdleConnsPerHost:   1000,
		maxConnsPerHost:       1000,
		idleConnTimeout:       60 * time.Second,
		tlsHandshakeTimeout:   10 * time.Second,
		expectContinueTimeout: 1 * time.Second,
		defaultTimeout:        defaultHTTPTimeout,
	}

	if len(options) > 0 {
		for _, option := range options {
			option.apply(o)
		}
	}

	t := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   o.dialTimeout,
			KeepAlive: o.dialKeepAlive,
		}).DialContext,
		MaxConnsPerHost:       o.maxConnsPerHost,
		MaxIdleConnsPerHost:   o.maxIdleConnsPerHost,
		MaxIdleConns:          o.maxIdleConns,
		IdleConnTimeout:       o.idleConnTimeout,
		TLSClientConfig:       o.tlsConfig,
		TLSHandshakeTimeout:   o.tlsHandshakeTimeout,
		ExpectContinueTimeout: o.expectContinueTimeout,
	}

	c := &HTTPClient{
		client: &http.Client{
			Transport: t,
		},
		timeout: o.defaultTimeout,
	}

	return c
}

// HTTPGet http get request
func HTTPGet(url string, options ...HTTPRequestOption) ([]byte, error) {
	return defaultHTTPClient.Get(url, options...)
}

// HTTPPost http post request
func HTTPPost(url string, body []byte, options ...HTTPRequestOption) ([]byte, error) {
	return defaultHTTPClient.Post(url, body, options...)
}
