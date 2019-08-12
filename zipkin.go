package yiigo

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/idgenerator"
	zipkinHTTP "github.com/openzipkin/zipkin-go/middleware/http"
	"github.com/openzipkin/zipkin-go/model"
	"github.com/openzipkin/zipkin-go/reporter"
	zipkinHTTPReporter "github.com/openzipkin/zipkin-go/reporter/http"
)

type zipkinTracerOptions struct {
	// tracer options
	tracerDefaultTags          map[string]string
	tracerExtractFailurePolicy zipkin.ExtractFailurePolicy
	tracerSampler              zipkin.Sampler
	tracerIDGenerator          idgenerator.IDGenerator
	tracerEndpoint             *model.Endpoint
	tracerNoop                 bool
	tracerSharedSpans          bool
	tracerUnsampledNoop        bool
	// reporter options
	reporterClientOptions   []HTTPClientOption
	reporterBatchInterval   time.Duration
	reporterBatchSize       int
	reporterMaxBacklog      int
	reporterRequestCallback zipkinHTTPReporter.RequestCallbackFn
	reporterLogger          *log.Logger
	reporterSerializer      reporter.SpanSerializer
}

// ZipkinTracerOption configures how we set up the zipkin tracer
type ZipkinTracerOption interface {
	apply(options *zipkinTracerOptions) error
}

// funcZipkinTracerOption implements zipkin tracer option
type funcZipkinTracerOption struct {
	f func(options *zipkinTracerOptions) error
}

func (fo *funcZipkinTracerOption) apply(o *zipkinTracerOptions) error {
	return fo.f(o)
}

func newFuncZipkinTracerOption(f func(o *zipkinTracerOptions) error) *funcZipkinTracerOption {
	return &funcZipkinTracerOption{f: f}
}

// WithZipkinTracerTag specifies the `Tags` to zipkin tracer.
func WithZipkinTracerTag(key, value string) ZipkinTracerOption {
	return newFuncZipkinTracerOption(func(o *zipkinTracerOptions) error {
		o.tracerDefaultTags[key] = value

		return nil
	})
}

// WithZipkinTracerExtractFailurePolicy specifies the `ExtractFailurePolicy` to zipkin tracer.
func WithZipkinTracerExtractFailurePolicy(p zipkin.ExtractFailurePolicy) ZipkinTracerOption {
	return newFuncZipkinTracerOption(func(o *zipkinTracerOptions) error {
		o.tracerExtractFailurePolicy = p

		return nil
	})
}

// WithZipkinTracerSamplerMod specifies the `Sampler` to zipkin tracer.
func WithZipkinTracerSamplerMod(m int) ZipkinTracerOption {
	return newFuncZipkinTracerOption(func(o *zipkinTracerOptions) error {
		o.tracerSampler = zipkin.NewModuloSampler(uint64(m))

		return nil
	})
}

// WithZipkinTracerIDGenerator specifies the `IDGenerator` to zipkin tracer.
func WithZipkinTracerIDGenerator(g idgenerator.IDGenerator) ZipkinTracerOption {
	return newFuncZipkinTracerOption(func(o *zipkinTracerOptions) error {
		o.tracerIDGenerator = g

		return nil
	})
}

// WithZipkinTracerEndpoint specifies the `Endpoint` to zipkin tracer.
func WithZipkinTracerEndpoint(name, host string) ZipkinTracerOption {
	return newFuncZipkinTracerOption(func(o *zipkinTracerOptions) error {
		endpoint, err := zipkin.NewEndpoint(name, host)

		if err != nil {
			return err
		}

		o.tracerEndpoint = endpoint

		return nil
	})
}

// WithZipkinTracerNoop specifies the `Noop` to zipkin tracer.
func WithZipkinTracerNoop(b bool) ZipkinTracerOption {
	return newFuncZipkinTracerOption(func(o *zipkinTracerOptions) error {
		o.tracerNoop = b

		return nil
	})
}

// WithZipkinTracerSharedSpans specifies the `SharedSpans` to zipkin tracer.
func WithZipkinTracerSharedSpans(b bool) ZipkinTracerOption {
	return newFuncZipkinTracerOption(func(o *zipkinTracerOptions) error {
		o.tracerSharedSpans = b

		return nil
	})
}

// WithZipkinTracerUnsampledNoop specifies the `UnsampledNoop` to zipkin tracer.
func WithZipkinTracerUnsampledNoop(b bool) ZipkinTracerOption {
	return newFuncZipkinTracerOption(func(o *zipkinTracerOptions) error {
		o.tracerUnsampledNoop = b

		return nil
	})
}

// WithZipkinReporterHTTPClient specifies the `Client` to zipkin reporter.
func WithZipkinReporterHTTPClient(options ...HTTPClientOption) ZipkinTracerOption {
	return newFuncZipkinTracerOption(func(o *zipkinTracerOptions) error {
		o.reporterClientOptions = options

		return nil
	})
}

// WithZipkinReporterHTTPClient specifies the `BatchInterval` to zipkin reporter.
func WithZipkinReporterBatchInterval(t time.Duration) ZipkinTracerOption {
	return newFuncZipkinTracerOption(func(o *zipkinTracerOptions) error {
		o.reporterBatchInterval = t

		return nil
	})
}

// WithZipkinReporterHTTPClient specifies the `BatchSize` to zipkin reporter.
func WithZipkinReporterBatchSize(i int) ZipkinTracerOption {
	return newFuncZipkinTracerOption(func(o *zipkinTracerOptions) error {
		o.reporterBatchSize = i

		return nil
	})
}

// WithZipkinReporterHTTPClient specifies the `MaxBacklog` to zipkin reporter.
func WithZipkinReporterMaxBacklog(i int) ZipkinTracerOption {
	return newFuncZipkinTracerOption(func(o *zipkinTracerOptions) error {
		o.reporterMaxBacklog = i

		return nil
	})
}

// WithZipkinReporterHTTPClient specifies the `RequestCallback` to zipkin reporter.
func WithZipkinReporterRequestCallback(fn zipkinHTTPReporter.RequestCallbackFn) ZipkinTracerOption {
	return newFuncZipkinTracerOption(func(o *zipkinTracerOptions) error {
		o.reporterRequestCallback = fn

		return nil
	})
}

// WithZipkinReporterHTTPClient specifies the `Logger` to zipkin reporter.
func WithZipkinReporterLogger(l *log.Logger) ZipkinTracerOption {
	return newFuncZipkinTracerOption(func(o *zipkinTracerOptions) error {
		o.reporterLogger = l

		return nil
	})
}

// WithZipkinReporterHTTPClient specifies the `Serializer` to zipkin reporter.
func WithZipkinReporterSerializer(s reporter.SpanSerializer) ZipkinTracerOption {
	return newFuncZipkinTracerOption(func(o *zipkinTracerOptions) error {
		o.reporterSerializer = s

		return nil
	})
}

// NewZipkinTracer returns a new zipin tracer
func NewZipkinTracer(reportURL string, options ...ZipkinTracerOption) (*zipkin.Tracer, error) {
	o := &zipkinTracerOptions{
		// zipkin tracer default options
		tracerDefaultTags: make(map[string]string),
		tracerSharedSpans: true,
	}

	if len(options) > 0 {
		for _, option := range options {
			if err := option.apply(o); err != nil {
				return nil, err
			}
		}
	}

	r := zipkinHTTPReporter.NewReporter(reportURL, buildZipkinReporterOptions(o)...)

	tracer, err := zipkin.NewTracer(r, buildZipkinTracerOptions(o)...)

	if err != nil {
		return nil, err
	}

	return tracer, nil
}

func buildZipkinTracerOptions(o *zipkinTracerOptions) []zipkin.TracerOption {
	options := make([]zipkin.TracerOption, 0, 8)

	if len(o.tracerDefaultTags) > 0 {
		options = append(options, zipkin.WithTags(o.tracerDefaultTags))
	}

	options = append(options, zipkin.WithExtractFailurePolicy(o.tracerExtractFailurePolicy))

	if o.tracerSampler != nil {
		options = append(options, zipkin.WithSampler(o.tracerSampler))
	}

	if o.tracerIDGenerator != nil {
		options = append(options, zipkin.WithIDGenerator(o.tracerIDGenerator))
	}

	if o.tracerEndpoint != nil {
		options = append(options, zipkin.WithLocalEndpoint(o.tracerEndpoint))
	}

	options = append(options, zipkin.WithNoopTracer(o.tracerNoop))
	options = append(options, zipkin.WithSharedSpans(o.tracerSharedSpans))
	options = append(options, zipkin.WithNoopSpan(o.tracerUnsampledNoop))

	return options
}

func buildZipkinReporterOptions(o *zipkinTracerOptions) []zipkinHTTPReporter.ReporterOption {
	reporterOptions := make([]zipkinHTTPReporter.ReporterOption, 0, 7)

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

	if len(o.reporterClientOptions) > 0 {
		for _, option := range o.reporterClientOptions {
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

	reporterOptions = append(reporterOptions, zipkinHTTPReporter.Client(client))

	if o.reporterBatchInterval != 0 {
		reporterOptions = append(reporterOptions, zipkinHTTPReporter.BatchInterval(o.reporterBatchInterval))
	}

	if o.reporterBatchSize != 0 {
		reporterOptions = append(reporterOptions, zipkinHTTPReporter.BatchSize(o.reporterBatchSize))
	}

	if o.reporterMaxBacklog != 0 {
		reporterOptions = append(reporterOptions, zipkinHTTPReporter.MaxBacklog(o.reporterMaxBacklog))
	}

	if o.reporterRequestCallback != nil {
		reporterOptions = append(reporterOptions, zipkinHTTPReporter.RequestCallback(o.reporterRequestCallback))
	}

	if o.reporterLogger != nil {
		reporterOptions = append(reporterOptions, zipkinHTTPReporter.Logger(o.reporterLogger))
	}

	if o.reporterSerializer != nil {
		reporterOptions = append(reporterOptions, zipkinHTTPReporter.Serializer(o.reporterSerializer))
	}

	return reporterOptions
}

type zipkinClientOptions struct {
	// zipkin client options
	clientOptions []HTTPClientOption
	clientTrace   bool
	clientTags    map[string]string
	// zipkin transport options
	roundTripper               http.RoundTripper
	transportTrace             bool
	transportTags              map[string]string
	transportErrHandler        zipkinHTTP.ErrHandler
	transportErrResponseReader zipkinHTTP.ErrResponseReader
	transportLogger            *log.Logger
	transportRequestSampler    zipkinHTTP.RequestSamplerFunc
}

// ZipkinClientOption configures how we set up the zipkin client
type ZipkinClientOption interface {
	apply(options *zipkinClientOptions)
}

// funcZipkinClientOption implements zipkin client option
type funcZipkinClientOption struct {
	f func(options *zipkinClientOptions)
}

func (fo *funcZipkinClientOption) apply(o *zipkinClientOptions) {
	fo.f(o)
}

func newFuncZipkinClientOption(f func(o *zipkinClientOptions)) *funcZipkinClientOption {
	return &funcZipkinClientOption{f: f}
}

// WithZipkinHTTPClient specifies the `Client` to zipkin client.
func WithZipkinHTTPClient(options ...HTTPClientOption) ZipkinClientOption {
	return newFuncZipkinClientOption(func(o *zipkinClientOptions) {
		o.clientOptions = options
	})
}

// WithZipkinClientTrace specifies the `HttpTrace` to zipkin client.
func WithZipkinClientTrace(b bool) ZipkinClientOption {
	return newFuncZipkinClientOption(func(o *zipkinClientOptions) {
		o.clientTrace = b
	})
}

// WithZipkinClientTag specifies the `Tags` to zipkin client.
func WithZipkinClientTag(key, value string) ZipkinClientOption {
	return newFuncZipkinClientOption(func(o *zipkinClientOptions) {
		o.clientTags[key] = value
	})
}

// WithZipkinRoundTripper specifies the `RoundTripper` to zipkin transport.
func WithZipkinRoundTripper(rt http.RoundTripper) ZipkinClientOption {
	return newFuncZipkinClientOption(func(o *zipkinClientOptions) {
		o.roundTripper = rt
	})
}

// WithZipkinTransportTrace specifies the `HttpTrace` to zipkin transport.
func WithZipkinTransportTrace(b bool) ZipkinClientOption {
	return newFuncZipkinClientOption(func(o *zipkinClientOptions) {
		o.transportTrace = b
	})
}

// WithZipkinTransportTag specifies the `Tags` to zipkin transport.
func WithZipkinTransportTag(key, value string) ZipkinClientOption {
	return newFuncZipkinClientOption(func(o *zipkinClientOptions) {
		o.transportTags[key] = value
	})
}

// WithZipkinTransportErrHandler specifies the `ErrHandler` to zipkin transport.
func WithZipkinTransportErrHandler(fn zipkinHTTP.ErrHandler) ZipkinClientOption {
	return newFuncZipkinClientOption(func(o *zipkinClientOptions) {
		o.transportErrHandler = fn
	})
}

// WithZipkinTransportErrResponseReader specifies the `ErrResponseReader` to zipkin transport.
func WithZipkinTransportErrResponseReader(fn zipkinHTTP.ErrResponseReader) ZipkinClientOption {
	return newFuncZipkinClientOption(func(o *zipkinClientOptions) {
		o.transportErrResponseReader = fn
	})
}

// WithZipkinTransportLogger specifies the `Logger` to zipkin transport.
func WithZipkinTransportLogger(l *log.Logger) ZipkinClientOption {
	return newFuncZipkinClientOption(func(o *zipkinClientOptions) {
		o.transportLogger = l
	})
}

// WithZipkinTransportRequestSampler specifies the `RequestSampler` to zipkin transport.
func WithZipkinTransportRequestSampler(fn zipkinHTTP.RequestSamplerFunc) ZipkinClientOption {
	return newFuncZipkinClientOption(func(o *zipkinClientOptions) {
		o.transportRequestSampler = fn
	})
}

// ZipkinClient zipkin client
type ZipkinClient struct {
	client  *zipkinHTTP.Client
	timeout time.Duration
}

// Get zipkin get request
func (z *ZipkinClient) Get(ctx context.Context, url string, options ...HTTPRequestOption) ([]byte, error) {
	o := &httpRequestOptions{
		headers: make(map[string]string),
		timeout: z.timeout,
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
	ctx, cancel := context.WithTimeout(zipkin.NewContext(req.Context(), span), o.timeout)

	defer cancel()

	resp, err := z.client.DoWithAppSpan(req.WithContext(ctx), fmt.Sprintf("%s:%s", req.Method, req.URL.Path))

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

// Post zipkin post request
func (z *ZipkinClient) Post(ctx context.Context, url string, body []byte, options ...HTTPRequestOption) ([]byte, error) {
	o := &httpRequestOptions{
		headers: make(map[string]string),
		timeout: z.timeout,
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
	ctx, cancel := context.WithTimeout(zipkin.NewContext(req.Context(), span), o.timeout)

	defer cancel()

	resp, err := z.client.DoWithAppSpan(req.WithContext(ctx), fmt.Sprintf("%s:%s", req.Method, req.URL.Path))

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

// NewZipkinClient returns a zipin client
func NewZipkinClient(t *zipkin.Tracer, options ...ZipkinClientOption) (*ZipkinClient, error) {
	o := &zipkinClientOptions{
		clientTags:    make(map[string]string),
		transportTags: make(map[string]string),
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
	}

	zipkinClient, err := zipkinHTTP.NewClient(t, buildZipkinClientOptions(client, o)...)

	if err != nil {
		return nil, err
	}

	return &ZipkinClient{
		client:  zipkinClient,
		timeout: clientOptions.defaultTimeout,
	}, nil
}

func buildZipkinClientOptions(c *http.Client, o *zipkinClientOptions) []zipkinHTTP.ClientOption {
	options := make([]zipkinHTTP.ClientOption, 0, 4)

	options = append(options, zipkinHTTP.WithClient(c))

	if len(o.clientTags) > 0 {
		options = append(options, zipkinHTTP.ClientTags(o.clientTags))
	}

	options = append(options, zipkinHTTP.ClientTrace(o.clientTrace))

	if v := buildZipkinTransportOptions(o); len(v) > 0 {
		options = append(options, zipkinHTTP.TransportOptions(v...))
	}

	return options
}

func buildZipkinTransportOptions(o *zipkinClientOptions) []zipkinHTTP.TransportOption {
	options := make([]zipkinHTTP.TransportOption, 0, 7)

	if o.roundTripper != nil {
		options = append(options, zipkinHTTP.RoundTripper(o.roundTripper))
	}

	if len(o.transportTags) > 0 {
		options = append(options, zipkinHTTP.TransportTags(o.transportTags))
	}

	options = append(options, zipkinHTTP.TransportTrace(o.transportTrace))

	if o.transportErrHandler != nil {
		options = append(options, zipkinHTTP.TransportErrHandler(o.transportErrHandler))
	}

	if o.transportErrResponseReader != nil {
		options = append(options, zipkinHTTP.TransportErrResponseReader(o.transportErrResponseReader))
	}

	if o.transportLogger != nil {
		options = append(options, zipkinHTTP.TransportLogger(o.transportLogger))
	}

	if o.transportRequestSampler != nil {
		options = append(options, zipkinHTTP.TransportRequestSampler(o.transportRequestSampler))
	}

	return options
}
