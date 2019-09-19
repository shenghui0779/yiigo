package yiigo

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/openzipkin/zipkin-go"
	zipkinHTTP "github.com/openzipkin/zipkin-go/middleware/http"
	"github.com/openzipkin/zipkin-go/model"
	"github.com/openzipkin/zipkin-go/propagation/b3"
	"github.com/openzipkin/zipkin-go/reporter"
	zipkinHTTPReporter "github.com/openzipkin/zipkin-go/reporter/http"
)

type zipkinReporterOptions struct {
	clientOptions   []HTTPClientOption
	reporterOptions []zipkinHTTPReporter.ReporterOption
}

// ZipkinReporterOption configures how we set up the zipkin reporter
type ZipkinReporterOption interface {
	apply(*zipkinReporterOptions) error
}

// funcZipkinReporterOption implements zipkin reporter option
type funcZipkinReporterOption struct {
	f func(*zipkinReporterOptions) error
}

func (fo *funcZipkinReporterOption) apply(o *zipkinReporterOptions) error {
	return fo.f(o)
}

func newFuncZipkinReporterOption(f func(*zipkinReporterOptions) error) *funcZipkinReporterOption {
	return &funcZipkinReporterOption{f: f}
}

// WithZipkinReporterClient specifies the `Client` to zipkin reporter.
func WithZipkinReporterClient(options ...HTTPClientOption) ZipkinReporterOption {
	return newFuncZipkinReporterOption(func(o *zipkinReporterOptions) error {
		o.clientOptions = options

		return nil
	})
}

// WithZipkinReporterOptions specifies the `Options` to zipkin reporter.
func WithZipkinReporterOptions(options ...zipkinHTTPReporter.ReporterOption) ZipkinReporterOption {
	return newFuncZipkinReporterOption(func(o *zipkinReporterOptions) error {
		o.reporterOptions = options

		return nil
	})
}

// NewZipkinHTTPReporter returns a new zipin http reporter
func NewZipkinHTTPReporter(url string, options ...ZipkinReporterOption) reporter.Reporter {
	o := &zipkinReporterOptions{
		clientOptions:   make([]HTTPClientOption, 0),
		reporterOptions: make([]zipkinHTTPReporter.ReporterOption, 0),
	}

	if len(options) > 0 {
		for _, option := range options {
			if err := option.apply(o); err != nil {
				return nil
			}
		}
	}

	clientOptions := &httpClientOptions{
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

	if len(o.clientOptions) > 0 {
		for _, option := range o.clientOptions {
			option.apply(clientOptions)
		}
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   clientOptions.dialTimeout,
				KeepAlive: clientOptions.dialKeepAlive,
			}).DialContext,
			MaxConnsPerHost:       clientOptions.maxConnsPerHost,
			MaxIdleConnsPerHost:   clientOptions.maxIdleConnsPerHost,
			MaxIdleConns:          clientOptions.maxIdleConns,
			IdleConnTimeout:       clientOptions.idleConnTimeout,
			TLSClientConfig:       clientOptions.tlsConfig,
			TLSHandshakeTimeout:   clientOptions.tlsHandshakeTimeout,
			ExpectContinueTimeout: clientOptions.expectContinueTimeout,
		},
		Timeout: clientOptions.defaultTimeout,
	}

	o.reporterOptions = append(o.reporterOptions, zipkinHTTPReporter.Client(client))

	return zipkinHTTPReporter.NewReporter(url, o.reporterOptions...)
}

type zipkinHTTPClientOptions struct {
	httpClientOptions   []HTTPClientOption
	zipkinClientOptions []zipkinHTTP.ClientOption
	transportOptions    []zipkinHTTP.TransportOption
}

// ZipkinHTTPClientOption configures how we set up the zipkin http client
type ZipkinHTTPClientOption interface {
	apply(*zipkinHTTPClientOptions)
}

// funcZipkinClientOption implements zipkin client option
type funcZipkinClientOption struct {
	f func(*zipkinHTTPClientOptions)
}

func (fo *funcZipkinClientOption) apply(o *zipkinHTTPClientOptions) {
	fo.f(o)
}

func newFuncZipkinClientOption(f func(*zipkinHTTPClientOptions)) *funcZipkinClientOption {
	return &funcZipkinClientOption{f: f}
}

// WithZipkinHTTPClient specifies the `Client` to zipkin http client.
func WithZipkinHTTPClient(options ...HTTPClientOption) ZipkinHTTPClientOption {
	return newFuncZipkinClientOption(func(o *zipkinHTTPClientOptions) {
		o.httpClientOptions = options
	})
}

// WithZipkinClientOptions specifies the `Options` to zipkin http client.
func WithZipkinClientOptions(options ...zipkinHTTP.ClientOption) ZipkinHTTPClientOption {
	return newFuncZipkinClientOption(func(o *zipkinHTTPClientOptions) {
		o.zipkinClientOptions = options
	})
}

// WithZipkinHTTPTransport specifies the `Transport` to zipkin http client transport.
func WithZipkinHTTPTransport(options ...zipkinHTTP.TransportOption) ZipkinHTTPClientOption {
	return newFuncZipkinClientOption(func(o *zipkinHTTPClientOptions) {
		o.transportOptions = options
	})
}

// ZipkinHTTPClient zipkin http client
type ZipkinHTTPClient struct {
	client  *zipkinHTTP.Client
	timeout time.Duration
}

// Get zipkin http get request
func (z *ZipkinHTTPClient) Get(ctx context.Context, url string, options ...HTTPRequestOption) ([]byte, error) {
	o := &httpRequestOptions{
		headers:        make(map[string]string),
		timeout:        z.timeout,
		zipkinSpanTags: make(map[string]string),
	}

	if len(options) > 0 {
		for _, option := range options {
			option.apply(o)
		}
	}

	// zipkin span
	span := zipkin.SpanOrNoopFromContext(ctx)

	defer span.Finish()

	span.Tag("request_id", strconv.FormatInt(time.Now().UnixNano(), 36))

	if len(o.zipkinSpanTags) > 0 {
		for k, v := range o.zipkinSpanTags {
			span.Tag(k, v)
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

	// zipkin ctx & timeout
	c, cancel := context.WithTimeout(zipkin.NewContext(req.Context(), span), o.timeout)

	defer cancel()

	resp, err := z.client.DoWithAppSpan(req.WithContext(c), fmt.Sprintf("%s:%s", req.Method, req.URL.Path))

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

// Post zipkin http post request
func (z *ZipkinHTTPClient) Post(ctx context.Context, url string, body []byte, options ...HTTPRequestOption) ([]byte, error) {
	o := &httpRequestOptions{
		headers:        make(map[string]string),
		timeout:        z.timeout,
		zipkinSpanTags: make(map[string]string),
	}

	if len(options) > 0 {
		for _, option := range options {
			option.apply(o)
		}
	}

	// zipkin span
	span := zipkin.SpanOrNoopFromContext(ctx)

	defer span.Finish()

	span.Tag("request_id", strconv.FormatInt(time.Now().UnixNano(), 36))

	if len(o.zipkinSpanTags) > 0 {
		for k, v := range o.zipkinSpanTags {
			span.Tag(k, v)
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

	// zipkin ctx & timeout
	c, cancel := context.WithTimeout(zipkin.NewContext(req.Context(), span), o.timeout)

	defer cancel()

	resp, err := z.client.DoWithAppSpan(req.WithContext(c), fmt.Sprintf("%s:%s", req.Method, req.URL.Path))

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

// ZipkinTracer zipkin tracer
type ZipkinTracer struct {
	tracer *zipkin.Tracer
}

// HTTPClient returns a new zipkin http client
func (z *ZipkinTracer) HTTPClient(options ...ZipkinHTTPClientOption) (*ZipkinHTTPClient, error) {
	o := &zipkinHTTPClientOptions{
		httpClientOptions:   make([]HTTPClientOption, 0),
		zipkinClientOptions: make([]zipkinHTTP.ClientOption, 0),
		transportOptions:    make([]zipkinHTTP.TransportOption, 0),
	}

	if len(options) > 0 {
		for _, option := range options {
			option.apply(o)
		}
	}

	clientOptions := &httpClientOptions{
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

	if len(o.httpClientOptions) > 0 {
		for _, option := range o.httpClientOptions {
			option.apply(clientOptions)
		}
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   clientOptions.dialTimeout,
				KeepAlive: clientOptions.dialKeepAlive,
			}).DialContext,
			MaxConnsPerHost:       clientOptions.maxConnsPerHost,
			MaxIdleConnsPerHost:   clientOptions.maxIdleConnsPerHost,
			MaxIdleConns:          clientOptions.maxIdleConns,
			IdleConnTimeout:       clientOptions.idleConnTimeout,
			TLSClientConfig:       clientOptions.tlsConfig,
			TLSHandshakeTimeout:   clientOptions.tlsHandshakeTimeout,
			ExpectContinueTimeout: clientOptions.expectContinueTimeout,
		},
	}

	o.zipkinClientOptions = append(o.zipkinClientOptions, zipkinHTTP.WithClient(client), zipkinHTTP.TransportOptions(o.transportOptions...))

	zipkinClient, err := zipkinHTTP.NewClient(z.tracer, o.zipkinClientOptions...)

	if err != nil {
		return nil, err
	}

	return &ZipkinHTTPClient{
		client:  zipkinClient,
		timeout: clientOptions.defaultTimeout,
	}, nil
}

// Start returns a new zipkin span
//
// use as below:
//
// span := yiigo.ZTracer.Start(r)
// defer span.Finish()
// ctx := zipkin.NewContext(r.Context(), span)
func (z *ZipkinTracer) Start(req *http.Request) zipkin.Span {
	// try to extract B3 Headers from upstream
	sc := z.tracer.Extract(b3.ExtractHTTP(req))

	endpoint, _ := zipkin.NewEndpoint("", req.RemoteAddr)

	// create Span using SpanContext if found
	sp := z.tracer.StartSpan(fmt.Sprintf("%s:%s", req.Method, req.URL.Path),
		zipkin.Kind(model.Server),
		zipkin.Parent(sc),
		zipkin.StartTime(time.Now()),
		zipkin.RemoteEndpoint(endpoint),
	)

	// tag typical HTTP request items
	zipkin.TagHTTPMethod.Set(sp, req.Method)
	zipkin.TagHTTPPath.Set(sp, req.URL.Path)

	if req.ContentLength > 0 {
		zipkin.TagHTTPRequestSize.Set(sp, strconv.FormatInt(req.ContentLength, 10))
	}

	return sp
}

var (
	ZTracer   *ZipkinTracer
	zipkinMap sync.Map
)

// RegisterZipkinTracer register a zipkin tracer
func RegisterZipkinTracer(name string, r reporter.Reporter, options ...zipkin.TracerOption) error {
	t, err := zipkin.NewTracer(r, options...)

	if err != nil {
		return err
	}

	ztracer := &ZipkinTracer{tracer: t}

	zipkinMap.Store(name, ztracer)

	if name == AsDefault {
		ZTracer = ztracer
	}

	return nil
}

// UseZipkinTracer returns a zipkin tracer
func UseZipkinTracer(name ...string) *ZipkinTracer {
	k := AsDefault

	if len(name) != 0 {
		k = name[0]
	}

	v, ok := zipkinMap.Load(k)

	if !ok {
		panic(fmt.Errorf("yiigo: zipkin.%s is not registered", name))
	}

	return v.(*ZipkinTracer)
}
