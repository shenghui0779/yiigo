package yiigo

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// defaultHTTPTimeout default http request timeout
const defaultHTTPTimeout = 10 * time.Second

// errCookieFileNotFound cookie file not found error
var errCookieFileNotFound = errors.New("cookie file not found")

// httpClientOptions http client options
type httpClientOptions struct {
	dialTimeout           time.Duration
	dialKeepAlive         time.Duration
	fallbackDelay         time.Duration
	maxConnsPerHost       int
	maxIdleConnsPerHost   int
	maxIdleConns          int
	idleConnTimeout       time.Duration
	sslCertificates       []tls.Certificate
	tlsHandshakeTimeout   time.Duration
	expectContinueTimeout time.Duration
}

// HTTPClientOption configures how we set up the http client
type HTTPClientOption interface {
	apply(options *httpClientOptions) error
}

// funcHTTPClientOption implements http client option
type funcHTTPClientOption struct {
	f func(options *httpClientOptions) error
}

func (fo *funcHTTPClientOption) apply(o *httpClientOptions) error {
	return fo.f(o)
}

func newFuncHTTPOption(f func(options *httpClientOptions) error) *funcHTTPClientOption {
	return &funcHTTPClientOption{f: f}
}

// WithHTTPDialTimeout specifies the `DialTimeout` to net.Dialer.
func WithHTTPDialTimeout(d time.Duration) HTTPClientOption {
	return newFuncHTTPOption(func(o *httpClientOptions) error {
		o.dialTimeout = d

		return nil
	})
}

// WithHTTPDialKeepAlive specifies the `KeepAlive` to net.Dialer.
func WithHTTPDialKeepAlive(d time.Duration) HTTPClientOption {
	return newFuncHTTPOption(func(o *httpClientOptions) error {
		o.dialKeepAlive = d

		return nil
	})
}

// WithHTTPDialFallbackDelay specifies the `FallbackDelay` to net.Dialer.
func WithHTTPDialFallbackDelay(d time.Duration) HTTPClientOption {
	return newFuncHTTPOption(func(o *httpClientOptions) error {
		o.fallbackDelay = d

		return nil
	})
}

// WithHTTPMaxConnsPerHost specifies the `MaxConnsPerHost` to http client.
func WithHTTPMaxConnsPerHost(n int) HTTPClientOption {
	return newFuncHTTPOption(func(o *httpClientOptions) error {
		o.maxConnsPerHost = n

		return nil
	})
}

// WithHTTPMaxIdleConnsPerHost specifies the `MaxIdleConnsPerHost` to http client.
func WithHTTPMaxIdleConnsPerHost(n int) HTTPClientOption {
	return newFuncHTTPOption(func(o *httpClientOptions) error {
		o.maxIdleConnsPerHost = n

		return nil
	})
}

// WithHTTPMaxIdleConns specifies the `MaxIdleConns` to http client.
func WithHTTPMaxIdleConns(n int) HTTPClientOption {
	return newFuncHTTPOption(func(o *httpClientOptions) error {
		o.maxIdleConns = n

		return nil
	})
}

// WithHTTPIdleConnTimeout specifies the `IdleConnTimeout` to http client.
func WithHTTPIdleConnTimeout(d time.Duration) HTTPClientOption {
	return newFuncHTTPOption(func(o *httpClientOptions) error {
		o.idleConnTimeout = d

		return nil
	})
}

// WithHTTPSSLCertFile specifies the TLS with cert file to http client.
func WithHTTPSSLCertFile(certFile, keyFile string) HTTPClientOption {
	return newFuncHTTPOption(func(o *httpClientOptions) error {
		certFilePath, err := filepath.Abs(certFile)

		if err != nil {
			return err
		}

		keyFilePath, err := filepath.Abs(keyFile)

		if err != nil {
			return err
		}

		cert, err := tls.LoadX509KeyPair(certFilePath, keyFilePath)

		if err != nil {
			return err
		}

		o.sslCertificates = []tls.Certificate{cert}

		return nil
	})
}

// WithHTTPSSLCertBlock specifies the TLS with cert pem block to http client.
func WithHTTPSSLCertBlock(certPEMBlock, keyPEMBlock []byte) HTTPClientOption {
	return newFuncHTTPOption(func(o *httpClientOptions) error {
		cert, err := tls.X509KeyPair(certPEMBlock, keyPEMBlock)

		if err != nil {
			return err
		}

		o.sslCertificates = []tls.Certificate{cert}

		return nil
	})
}

// WithHTTPTLSHandshakeTimeout specifies the `TLSHandshakeTimeout` to http client.
func WithHTTPTLSHandshakeTimeout(d time.Duration) HTTPClientOption {
	return newFuncHTTPOption(func(o *httpClientOptions) error {
		o.tlsHandshakeTimeout = d

		return nil
	})
}

// WithHTTPExpectContinueTimeout specifies the `ExpectContinueTimeout` to http client.
func WithHTTPExpectContinueTimeout(d time.Duration) HTTPClientOption {
	return newFuncHTTPOption(func(o *httpClientOptions) error {
		o.expectContinueTimeout = d

		return nil
	})
}

// httpRequestOptions http request options
type httpRequestOptions struct {
	headers          map[string]string
	cookieFile       string
	withCookies      bool
	cookieSave       bool
	cookieReplace    bool
	disableKeepAlive bool
	timeout          time.Duration
}

// HTTPRequestOption configures how we set up the http request
type HTTPRequestOption interface {
	apply(*httpRequestOptions) error
}

// funcHTTPRequestOption implements request option
type funcHTTPRequestOption struct {
	f func(*httpRequestOptions) error
}

func (fo *funcHTTPRequestOption) apply(r *httpRequestOptions) error {
	return fo.f(r)
}

func newFuncHTTPRequestOption(f func(*httpRequestOptions) error) *funcHTTPRequestOption {
	return &funcHTTPRequestOption{f: f}
}

// WithRequestHeader specifies the headers to http request.
func WithRequestHeader(key, value string) HTTPRequestOption {
	return newFuncHTTPRequestOption(func(o *httpRequestOptions) error {
		o.headers[key] = value

		return nil
	})
}

// WithRequestCookieFile specifies the file which to save http response cookies.
func WithRequestCookieFile(file string) HTTPRequestOption {
	return newFuncHTTPRequestOption(func(o *httpRequestOptions) error {
		path, err := filepath.Abs(file)

		if err != nil {
			return err
		}

		o.cookieFile = path

		return mkCookieFile(path)
	})
}

// WithRequestCookies specifies http requested with cookies.
func WithRequestCookies() HTTPRequestOption {
	return newFuncHTTPRequestOption(func(o *httpRequestOptions) error {
		o.withCookies = true

		return nil
	})
}

// WithRequestCookieSave specifies save the http response cookies.
func WithRequestCookieSave() HTTPRequestOption {
	return newFuncHTTPRequestOption(func(o *httpRequestOptions) error {
		o.cookieSave = true

		return nil
	})
}

// WithRequestCookieReplace specifies replace the old http response cookies.
func WithRequestCookieReplace() HTTPRequestOption {
	return newFuncHTTPRequestOption(func(o *httpRequestOptions) error {
		o.cookieReplace = true

		return nil
	})
}

// WithRequestDisableKeepAlive specifies close the connection after
// replying to this request (for servers) or after sending this
// request and reading its response (for clients).
func WithRequestDisableKeepAlive() HTTPRequestOption {
	return newFuncHTTPRequestOption(func(o *httpRequestOptions) error {
		o.disableKeepAlive = true

		return nil
	})
}

// WithRequestTimeout specifies the timeout to http request.
func WithRequestTimeout(d time.Duration) HTTPRequestOption {
	return newFuncHTTPRequestOption(func(o *httpRequestOptions) error {
		o.timeout = d

		return nil
	})
}

// HTTPClient http client
type HTTPClient struct {
	client *http.Client
}

// Get http get request
func (h *HTTPClient) Get(url string, options ...HTTPRequestOption) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, err
	}

	o := &httpRequestOptions{
		headers: make(map[string]string),
		timeout: defaultHTTPTimeout,
	}

	if len(options) > 0 {
		for _, option := range options {
			if err := option.apply(o); err != nil {
				return nil, err
			}
		}
	}

	if len(o.headers) > 0 {
		for k, v := range o.headers {
			req.Header.Set(k, v)
		}
	}

	if o.withCookies {
		cookies, err := getCookies(o.cookieFile)

		if err != nil {
			return nil, err
		}

		for _, c := range cookies {
			req.AddCookie(c)
		}
	}

	if o.disableKeepAlive {
		req.Close = true
	}

	ctx, cancel := context.WithTimeout(context.TODO(), o.timeout)

	defer cancel()

	resp, err := h.client.Do(req.WithContext(ctx))

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if o.cookieSave {
		if err := saveCookie(resp.Cookies(), o.cookieFile, o.cookieReplace); err != nil {
			return nil, err
		}
	}

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
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))

	if err != nil {
		return nil, err
	}

	o := &httpRequestOptions{
		headers: make(map[string]string),
		timeout: defaultHTTPTimeout,
	}

	if len(options) > 0 {
		for _, option := range options {
			if err := option.apply(o); err != nil {
				return nil, err
			}
		}
	}

	if len(o.headers) > 0 {
		for k, v := range o.headers {
			req.Header.Set(k, v)
		}
	}

	if o.withCookies {
		cookies, err := getCookies(o.cookieFile)

		if err != nil {
			return nil, err
		}

		for _, c := range cookies {
			req.AddCookie(c)
		}
	}

	if o.disableKeepAlive {
		req.Close = true
	}

	ctx, cancel := context.WithTimeout(context.TODO(), o.timeout)

	defer cancel()

	resp, err := h.client.Do(req.WithContext(ctx))

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if o.cookieSave {
		if err := saveCookie(resp.Cookies(), o.cookieFile, o.cookieReplace); err != nil {
			return nil, err
		}
	}

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
			MaxIdleConnsPerHost:   10,
			MaxIdleConns:          100,
			IdleConnTimeout:       60 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	},
}

// NewHTTPClient returns a new http client
func NewHTTPClient(options ...HTTPClientOption) (*HTTPClient, error) {
	o := &httpClientOptions{
		dialTimeout:           30 * time.Second,
		dialKeepAlive:         60 * time.Second,
		maxIdleConnsPerHost:   10,
		maxIdleConns:          100,
		idleConnTimeout:       60 * time.Second,
		tlsHandshakeTimeout:   10 * time.Second,
		expectContinueTimeout: 1 * time.Second,
	}

	if len(options) > 0 {
		for _, option := range options {
			if err := option.apply(o); err != nil {
				return nil, err
			}
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

	if len(o.sslCertificates) > 0 {
		t.TLSClientConfig = &tls.Config{
			Certificates: o.sslCertificates,
		}
	}

	c := &HTTPClient{
		client: &http.Client{
			Transport: t,
		},
	}

	return c, nil
}

// HTTPGet http get request
func HTTPGet(url string, options ...HTTPRequestOption) ([]byte, error) {
	return defaultHTTPClient.Get(url, options...)
}

// HTTPPost http post request
func HTTPPost(url string, body []byte, options ...HTTPRequestOption) ([]byte, error) {
	return defaultHTTPClient.Post(url, body, options...)
}

// mkCookieFile create cookie file
func mkCookieFile(path string) error {
	dir := filepath.Dir(path)

	// make dir if not exsit
	if _, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}
	}

	// create file if not exsit
	if _, err := os.Stat(path); err != nil && os.IsNotExist(err) {
		if _, err := os.Create(path); err != nil {
			return err
		}
	}

	return nil
}

// getCookie get http saved cookies
func getCookies(file string) ([]*http.Cookie, error) {
	if file == "" {
		return nil, errCookieFileNotFound
	}

	cookieM := make(map[string]*http.Cookie)
	content, err := ioutil.ReadFile(file)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(content, &cookieM); err != nil {
		return nil, err
	}

	cookies := make([]*http.Cookie, 0, len(cookieM))

	for _, v := range cookieM {
		cookies = append(cookies, v)
	}

	return cookies, nil
}

// saveCookie save http cookies
func saveCookie(cookies []*http.Cookie, file string, replace bool) error {
	if len(cookies) == 0 {
		return nil
	}

	if file == "" {
		return errCookieFileNotFound
	}

	cookieM := make(map[string]*http.Cookie)

	if !replace {
		content, err := ioutil.ReadFile(file)

		if err != nil {
			return err
		}

		if len(content) > 0 {
			if err := json.Unmarshal(content, &cookieM); err != nil {
				return err
			}
		}
	}

	for _, c := range cookies {
		cookieM[c.Name] = c
	}

	b, err := json.Marshal(cookieM)

	if err != nil {
		return err
	}

	return ioutil.WriteFile(file, b, 0777)
}
