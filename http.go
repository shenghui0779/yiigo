package yiigo

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"

	"go.uber.org/zap"
)

// defaultHTTPTimeout default http request timeout
const defaultHTTPTimeout = 10 * time.Second

// tlsOptions https tls options
type tlsOptions struct {
	rootCAs            [][]byte
	certificates       []tls.Certificate
	insecureSkipVerify bool
}

// TLSOption configures how we set up the https transport
type TLSOption interface {
	apply(*tlsOptions)
}

// funcTLSOption implements tls option
type funcTLSOption struct {
	f func(*tlsOptions)
}

func (fo *funcTLSOption) apply(o *tlsOptions) {
	fo.f(o)
}

func newFuncTLSOption(f func(*tlsOptions)) *funcTLSOption {
	return &funcTLSOption{f: f}
}

// WithRootCA specifies the `RootCAs` to https transport.
func WithRootCA(crt []byte) TLSOption {
	return newFuncTLSOption(func(o *tlsOptions) {
		o.rootCAs = append(o.rootCAs, crt)
	})
}

// WithInsecureSkipVerify specifies the `Certificates` to https transport.
func WithCertificates(certs ...tls.Certificate) TLSOption {
	return newFuncTLSOption(func(o *tlsOptions) {
		o.certificates = certs
	})
}

// WithInsecureSkipVerify specifies the `InsecureSkipVerify` to https transport.
func WithInsecureSkipVerify(b bool) TLSOption {
	return newFuncTLSOption(func(o *tlsOptions) {
		o.insecureSkipVerify = b
	})
}

// httpClientOptions http client options
type httpClientOptions struct {
	dialTimeout           time.Duration
	dialKeepAlive         time.Duration
	fallbackDelay         time.Duration
	maxIdleConns          int
	maxIdleConnsPerHost   int
	maxConnsPerHost       int
	idleConnTimeout       time.Duration
	proxyURL              *url.URL
	tlsConfig             []TLSOption
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

// WithHTTPProxy specifies the `Proxy` to http client.
func WithHTTPProxy(proxyURL string) HTTPClientOption {
	return newFuncHTTPOption(func(o *httpClientOptions) {
		fixedURL, err := url.Parse(proxyURL)

		if err != nil {
			logger.Error("yiigo: parse proxy url error", zap.Error(err))

			return
		}

		o.proxyURL = fixedURL
	})
}

// WithHTTPTLSConfig specifies the `TLSClientConfig` to http client.
func WithHTTPTLSConfig(options ...TLSOption) HTTPClientOption {
	return newFuncHTTPOption(func(o *httpClientOptions) {
		o.tlsConfig = options
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
	headers map[string]string
	cookies []*http.Cookie
	close   bool
	timeout time.Duration
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

// HTTPClient http client
type HTTPClient struct {
	client  *http.Client
	timeout time.Duration
}

// Get http get request
func (h *HTTPClient) Get(reqURL string, options ...HTTPRequestOption) ([]byte, error) {
	o := &httpRequestOptions{
		headers: make(map[string]string),
		timeout: h.timeout,
	}

	if len(options) > 0 {
		for _, option := range options {
			option.apply(o)
		}
	}

	req, err := http.NewRequest("GET", reqURL, nil)

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
func (h *HTTPClient) Post(reqURL string, body []byte, options ...HTTPRequestOption) ([]byte, error) {
	o := &httpRequestOptions{
		headers: make(map[string]string),
		timeout: h.timeout,
	}

	if len(options) > 0 {
		for _, option := range options {
			option.apply(o)
		}
	}

	req, err := http.NewRequest("POST", reqURL, bytes.NewReader(body))

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
		TLSHandshakeTimeout:   o.tlsHandshakeTimeout,
		ExpectContinueTimeout: o.expectContinueTimeout,
	}

	// set proxy
	if o.proxyURL != nil {
		t.Proxy = http.ProxyURL(o.proxyURL)
	}

	// set tls client config
	if len(o.tlsConfig) > 0 {
		tlso := new(tlsOptions)

		for _, cfg := range o.tlsConfig {
			cfg.apply(tlso)
		}

		tlsCfg := new(tls.Config)

		if len(tlso.rootCAs) > 0 {
			pool := x509.NewCertPool()

			for _, b := range tlso.rootCAs {
				pool.AppendCertsFromPEM(b)
			}

			tlsCfg.RootCAs = pool
		} else {
			if tlso.insecureSkipVerify {
				tlsCfg.InsecureSkipVerify = true
			}
		}

		if len(tlso.certificates) > 0 {
			tlsCfg.Certificates = tlso.certificates
		}

		t.TLSClientConfig = tlsCfg
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
func HTTPGet(reqURL string, options ...HTTPRequestOption) ([]byte, error) {
	return defaultHTTPClient.Get(reqURL, options...)
}

// HTTPPost http post request
func HTTPPost(reqURL string, body []byte, options ...HTTPRequestOption) ([]byte, error) {
	return defaultHTTPClient.Post(reqURL, body, options...)
}
