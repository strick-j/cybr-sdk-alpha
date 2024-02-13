package http

import (
	"crypto/tls"
	"net"
	"net/http"
	"reflect"
	"sync"
	"time"
)

// Defaults for the HTTPTransportBuilder.
var (
	// Default connection pool options
	DefaultHTTPTransportMaxIdleConns        = 100
	DefaultHTTPTransportMaxIdleConnsPerHost = 10

	// Default connection timeouts
	DefaultHTTPTransportIdleConnTimeout       = 90 * time.Second
	DefaultHTTPTransportTLSHandleshakeTimeout = 10 * time.Second
	DefaultHTTPTransportExpectContinueTimeout = 1 * time.Second

	// Default to TLS 1.2 for all HTTPS requests.
	DefaultHTTPTransportTLSMinVersion uint16 = tls.VersionTLS12
)

// Timeouts for net.Dialer's network connection.
var (
	DefaultDialConnectTimeout   = 30 * time.Second
	DefaultDialKeepAliveTimeout = 30 * time.Second
)

// HTTPTransportBuilder provides a builder pattern for creating an HTTP Transport.
type HTTPTransportBuilder struct {
	transport *http.Transport
	dialer    *net.Dialer

	initOnce sync.Once

	clientTimeout time.Duration
	client        *http.Client
}

// NewHTTPTransportBuilder returns a new HTTPTransportBuilder with default values.
func NewHTTPTransportBuilder() *HTTPTransportBuilder {
	return &HTTPTransportBuilder{}
}

// Do implements the HTTPClient interface's Do method to invoke a HTTP request,
// and receive the response. Uses the BuildableClient's current
// configuration to invoke the http.Request.
func (b *HTTPTransportBuilder) Do(req *http.Request) (*http.Response, error) {
	b.initOnce.Do(b.build)

	return b.client.Do(req)
}

func (b *HTTPTransportBuilder) build() {
	b.client = wrapWithLimitedRedirect(&http.Client{
		Transport: b.GetTransport(),
		Timeout:   b.clientTimeout,
	})
}

func (b *HTTPTransportBuilder) clone() *HTTPTransportBuilder {
	cpy := NewHTTPTransportBuilder()
	cpy.transport = b.GetTransport()
	cpy.dialer = b.GetDialer()
	cpy.clientTimeout = b.clientTimeout

	return cpy
}

// WithTimeout Sets the timeout for all requests made by the client.
func (b *HTTPTransportBuilder) WithTimeout(timeout time.Duration) *HTTPTransportBuilder {
	cpy := b.clone()
	cpy.clientTimeout = timeout

	return cpy
}

// GetTransport returns the client's transport.
func (b *HTTPTransportBuilder) GetTransport() *http.Transport {
	var tr *http.Transport
	if b.transport != nil {
		tr = b.transport.Clone()
	} else {
		tr = defaultHTTPTransport()
	}

	return tr
}

// GetDialer returns the client's dialer.
func (b *HTTPTransportBuilder) GetDialer() *net.Dialer {
	var dialer *net.Dialer
	if b.dialer != nil {
		dialer = shallowCopyStruct(b.dialer).(*net.Dialer)
	} else {
		dialer = defaultDialer()
	}

	return dialer
}

// GetTimeout returns the client timeout.
func (b *HTTPTransportBuilder) GetTimeout() time.Duration {
	return b.clientTimeout
}

func defaultDialer() *net.Dialer {
	return &net.Dialer{
		Timeout:   DefaultDialConnectTimeout,
		KeepAlive: DefaultDialKeepAliveTimeout,
	}
}

func defaultHTTPTransport() *http.Transport {
	dialer := defaultDialer()

	tr := &http.Transport{
		DialContext:           dialer.DialContext,
		MaxIdleConns:          DefaultHTTPTransportMaxIdleConns,
		MaxIdleConnsPerHost:   DefaultHTTPTransportMaxIdleConnsPerHost,
		IdleConnTimeout:       DefaultHTTPTransportIdleConnTimeout,
		TLSHandshakeTimeout:   DefaultHTTPTransportTLSHandleshakeTimeout,
		ExpectContinueTimeout: DefaultHTTPTransportExpectContinueTimeout,
		ForceAttemptHTTP2:     true,
		TLSClientConfig: &tls.Config{
			MinVersion: DefaultHTTPTransportTLSMinVersion,
		},
	}

	return tr
}

// shallowCopyStruct creates a shallow copy of the passed in source struct, and
// returns that copy of the same struct type.
func shallowCopyStruct(src interface{}) interface{} {
	srcVal := reflect.ValueOf(src)
	srcValType := srcVal.Type()

	var returnAsPtr bool
	if srcValType.Kind() == reflect.Ptr {
		srcVal = srcVal.Elem()
		srcValType = srcValType.Elem()
		returnAsPtr = true
	}
	dstVal := reflect.New(srcValType).Elem()

	for i := 0; i < srcValType.NumField(); i++ {
		ft := srcValType.Field(i)
		if len(ft.PkgPath) != 0 {
			// unexported fields have a PkgPath
			continue
		}

		dstVal.Field(i).Set(srcVal.Field(i))
	}

	if returnAsPtr {
		dstVal = dstVal.Addr()
	}

	return dstVal.Interface()
}

// wrapWithLimitedRedirect updates the Client's Transport and CheckRedirect to
// not follow any redirect other than 307 and 308. No other redirect will be
// followed.
//
// If the client does not have a Transport defined will use a new SDK default
// http.Transport configuration.
func wrapWithLimitedRedirect(c *http.Client) *http.Client {
	tr := c.Transport
	if tr == nil {
		tr = defaultHTTPTransport()
	}

	cc := *c
	cc.CheckRedirect = limitedRedirect
	cc.Transport = suppressBadHTTPRedirectTransport{
		tr: tr,
	}

	return &cc
}

// limitedRedirect is a CheckRedirect that prevents the client from following
// any non 307/308 HTTP status code redirects.
//
// The 307 and 308 redirects are allowed because the client must use the
// original HTTP method for the redirected to location. Whereas 301 and 302
// allow the client to switch to GET for the redirect.
//
// Suppresses all redirect requests with a URL of badHTTPRedirectLocation.
func limitedRedirect(r *http.Request, via []*http.Request) error {
	// Request.Response, in CheckRedirect is the response that is triggering
	// the redirect.
	resp := r.Response
	if r.URL.String() == badHTTPRedirectLocation {
		resp.Header.Del(badHTTPRedirectLocation)
		return http.ErrUseLastResponse
	}

	switch resp.StatusCode {
	case 307, 308:
		// Only allow 307 and 308 redirects as they preserve the method.
		return nil
	}

	return http.ErrUseLastResponse
}

// suppressBadHTTPRedirectTransport provides an http.RoundTripper
// implementation that wraps another http.RoundTripper to prevent HTTP client
// receiving 301 and 302 HTTP responses redirects without the required location
// header.
//
// Clients using this utility must have a CheckRedirect, e.g. limitedRedirect,
// that check for responses with having a URL of baseHTTPRedirectLocation, and
// suppress the redirect.
type suppressBadHTTPRedirectTransport struct {
	tr http.RoundTripper
}

const badHTTPRedirectLocation = `https://cyberark.com/badhttpredirectlocation`

// RoundTrip backfills a stub location when a 301/302 response is received
// without a location. This stub location is used by limitedRedirect to prevent
// the HTTP client from failing attempting to use follow a redirect without a
// location value.
func (t suppressBadHTTPRedirectTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	resp, err := t.tr.RoundTrip(r)
	if err != nil {
		return resp, err
	}

	// S3 is the only known service to return 301 without location header.
	// The Go standard library HTTP client will return an opaque error if it
	// tries to follow a 301/302 response missing the location header.
	switch resp.StatusCode {
	case 301, 302:
		if v := resp.Header.Get("Location"); len(v) == 0 {
			resp.Header.Set("Location", badHTTPRedirectLocation)
		}
	}

	return resp, err
}
