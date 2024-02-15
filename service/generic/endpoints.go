package generic

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/strick-j/cybr-sdk-alpha/cybr"
	cybrmiddleware "github.com/strick-j/cybr-sdk-alpha/cybr/middleware"
	internalendpoints "github.com/strick-j/cybr-sdk-alpha/service/generic/internal"
	smithyendpoints "github.com/strick-j/smithy-go/endpoints"
	middleware "github.com/strick-j/smithy-go/middleware"
	"github.com/strick-j/smithy-go/ptr"
	smithyhttp "github.com/strick-j/smithy-go/transport/http"
)

// EndpointResolverOptions is the service endpoint resolver options
type EndpointResolverOptions = internalendpoints.Options

// EndpointResolver interface for resolving service endpoints.
type EndpointResolver interface {
	ResolveEndpoint(subdomain, domain string, options EndpointResolverOptions) (cybr.Endpoint, error)
}

var _ EndpointResolver = &internalendpoints.Resolver{}

// NewDefaultEndpointResolver constructs a new service endpoint resolver
func NewDefaultEndpointResolver() *internalendpoints.Resolver {
	return internalendpoints.New()
}

// EndpointResolverFunc is a helper utility that wraps a function so it satisfies
// the EndpointResolver interface. This is useful when you want to add additional
// endpoint resolving logic, or stub out specific endpoints with custom values.
type EndpointResolverFunc func(subdomain, domain string, options EndpointResolverOptions) (cybr.Endpoint, error)

func (fn EndpointResolverFunc) ResolveEndpoint(subdomain, domain string, options EndpointResolverOptions) (endpoint cybr.Endpoint, err error) {
	return fn(subdomain, domain, options)
}

// EndpointResolverFromURL returns an EndpointResolver configured using the
// provided endpoint url. By default, the resolved endpoint resolver uses the
// client region as signing region, and the endpoint source is set to
// EndpointSourceCustom.You can provide functional options to configure endpoint
// values for the resolved endpoint.
func EndpointResolverFromURL(url string, optFns ...func(*cybr.Endpoint)) EndpointResolver {
	e := cybr.Endpoint{URL: url, Source: cybr.EndpointSourceCustom}
	for _, fn := range optFns {
		fn(&e)
	}

	return EndpointResolverFunc(
		func(subdomain, domain string, options EndpointResolverOptions) (cybr.Endpoint, error) {
			return e, nil
		},
	)
}

type ResolveEndpoint struct {
	Resolver EndpointResolver
	Options  EndpointResolverOptions
}

func (*ResolveEndpoint) ID() string {
	return "ResolveEndpoint"
}

func (m *ResolveEndpoint) HandleSerialize(ctx context.Context, in middleware.SerializeInput, next middleware.SerializeHandler) (
	out middleware.SerializeOutput, metadata middleware.Metadata, err error,
) {
	req, ok := in.Request.(*smithyhttp.Request)
	if !ok {
		return out, metadata, fmt.Errorf("unknown transport type %T", in.Request)
	}

	if m.Resolver == nil {
		return out, metadata, fmt.Errorf("expected endpoint resolver to not be nil")
	}

	eo := m.Options
	eo.Logger = middleware.GetLogger(ctx)

	var endpoint cybr.Endpoint
	endpoint, err = m.Resolver.ResolveEndpoint(cybrmiddleware.GetSubdomain(ctx), cybrmiddleware.GetDomain(ctx), eo)
	if err != nil {
		nf := (&cybr.EndpointNotFoundError{})
		if errors.As(err, &nf) {
			return next.HandleSerialize(ctx, in)
		}
		return out, metadata, fmt.Errorf("failed to resolve service endpoint, %w", err)
	}

	req.URL, err = url.Parse(endpoint.URL)
	if err != nil {
		return out, metadata, fmt.Errorf("failed to parse endpoint URL: %w", err)
	}

	ctx = cybrmiddleware.SetEndpointSource(ctx, endpoint.Source)
	ctx = smithyhttp.SetHostnameImmutable(ctx, endpoint.HostnameImmutable)
	ctx = cybrmiddleware.SetPartitionID(ctx, endpoint.PartitionID)
	return next.HandleSerialize(ctx, in)
}
func addResolveEndpointMiddleware(stack *middleware.Stack, o Options) error {
	return stack.Serialize.Insert(&ResolveEndpoint{
		Resolver: o.EndpointResolver,
		Options:  o.EndpointOptions,
	}, "OperationSerializer", middleware.Before)
}

func removeResolveEndpointMiddleware(stack *middleware.Stack) error {
	_, err := stack.Serialize.Remove((&ResolveEndpoint{}).ID())
	return err
}

type wrappedEndpointResolver struct {
	cybrResolver cybr.EndpointResolverWithOptions
}

func (w *wrappedEndpointResolver) ResolveEndpoint(subdomain, domain string, options EndpointResolverOptions) (endpoint cybr.Endpoint, err error) {
	return w.cybrResolver.ResolveEndpoint(subdomain, ServiceID, domain, options)
}

type cybrEndpointResolverAdaptor func(subdomain, service, domain string) (cybr.Endpoint, error)

func (a cybrEndpointResolverAdaptor) ResolveEndpoint(subdomain, service, domain string, options ...interface{}) (cybr.Endpoint, error) {
	return a(subdomain, service, domain)
}

var _ cybr.EndpointResolverWithOptions = cybrEndpointResolverAdaptor(nil)

// withEndpointResolver returns an aws.EndpointResolverWithOptions that first delegates endpoint resolution to the awsResolver.
// If awsResolver returns aws.EndpointNotFoundError error, the v1 resolver middleware will swallow the error,
// and set an appropriate context flag such that fallback will occur when EndpointResolverV2 is invoked
// via its middleware.
//
// If another error (besides aws.EndpointNotFoundError) is returned, then that error will be propagated.
func withEndpointResolver(cybrResolver cybr.EndpointResolver, cybrResolverWithOptions cybr.EndpointResolverWithOptions) EndpointResolver {
	var resolver cybr.EndpointResolverWithOptions

	if cybrResolverWithOptions != nil {
		resolver = cybrResolverWithOptions
	} else if cybrResolver != nil {
		resolver = cybrEndpointResolverAdaptor(cybrResolver.ResolveEndpoint)
	}

	return &wrappedEndpointResolver{
		cybrResolver: resolver,
	}
}

func resolveEndpointResolverV2(options *Options) {
	if options.EndpointResolverV2 == nil {
		options.EndpointResolverV2 = NewDefaultEndpointResolverV2()
	}
}

// EndpointResolverV2 provides the interface for resolving service endpoints.
type EndpointResolverV2 interface {
	// ResolveEndpoint attempts to resolve the endpoint with the provided options,
	// returning the endpoint if found. Otherwise an error is returned.
	ResolveEndpoint(ctx context.Context, params EndpointParameters) (
		smithyendpoints.Endpoint, error,
	)
}

// EndpointParameters provides the parameters that influence how endpoints are
// resolved.
type EndpointParameters struct {
	// The CYBR domain used to dispatch the request.
	//
	// Parameter is
	// required.
	//
	// CYBR::Domain
	Domain *string

	// The CYBR subdomain used to dispatch the request.
	//
	// Parameter is
	// required.
	//
	// CYBR::Subdomain
	Subdomain *string

	// Override the endpoint used to send this request
	//
	// Parameter is
	// required.
	//
	// SDK::Endpoint
	Endpoint *string
}

// ValidateRequired validates required parameters are set.
func (p EndpointParameters) ValidateRequired() error {
	if p.Domain == nil {
		return fmt.Errorf("parameter Domain is required")
	}

	return nil
}

// WithDefaults returns a shallow copy of EndpointParameterswith default values
// applied to members where applicable.
func (p EndpointParameters) WithDefaults() EndpointParameters {
	if p.Domain == nil {
		p.Domain = ptr.String("cyberark.cloud")
	}
	return p
}

// resolver provides the implementation for resolving endpoints.
type resolver struct{}

func NewDefaultEndpointResolverV2() EndpointResolverV2 {
	return &resolver{}
}

// ResolveEndpoint attempts to resolve the endpoint with the provided options,
// returning the endpoint if found. Otherwise an error is returned.
func (r *resolver) ResolveEndpoint(
	ctx context.Context, params EndpointParameters,
) (
	endpoint smithyendpoints.Endpoint, err error,
) {
	params = params.WithDefaults()
	if err = params.ValidateRequired(); err != nil {
		return endpoint, fmt.Errorf("endpoint parameters are not valid, %w", err)
	}
	_Domain := *params.Domain
	_Subdomain := *params.Subdomain

	uriString := func() string {
		var out strings.Builder
		out.WriteString("https://")
		out.WriteString(_Subdomain)
		out.WriteString(".")
		out.WriteString(_Domain)
		return out.String()
	}()

	uri, err := url.Parse(uriString)
	if err != nil {
		return endpoint, fmt.Errorf("Failed to parse uri: %s", uriString)
	}

	return smithyendpoints.Endpoint{
		URI:     *uri,
		Headers: http.Header{},
	}, nil
}

type endpointParamsBinder interface {
	bindEndpointParams(*EndpointParameters)
}

func bindEndpointParams(input interface{}, options Options) *EndpointParameters {
	params := &EndpointParameters{}

	params.Domain = cybr.String(options.Domain)
	params.Subdomain = cybr.String(options.Subdomain)

	if b, ok := input.(endpointParamsBinder); ok {
		b.bindEndpointParams(params)
	}

	return params
}

type resolveEndpointV2Middleware struct {
	options Options
}

func (*resolveEndpointV2Middleware) ID() string {
	return "ResolveEndpointV2"
}

func (m *resolveEndpointV2Middleware) HandleFinalize(ctx context.Context, in middleware.FinalizeInput, next middleware.FinalizeHandler) (
	out middleware.FinalizeOutput, metadata middleware.Metadata, err error,
) {
	req, ok := in.Request.(*smithyhttp.Request)
	if !ok {
		return out, metadata, fmt.Errorf("unknown transport type %T", in.Request)
	}

	if m.options.EndpointResolverV2 == nil {
		return out, metadata, fmt.Errorf("expected endpoint resolver to not be nil")
	}

	params := bindEndpointParams(getOperationInput(ctx), m.options)
	endpt, err := m.options.EndpointResolverV2.ResolveEndpoint(ctx, *params)
	if err != nil {
		return out, metadata, fmt.Errorf("failed to resolve service endpoint, %w", err)
	}

	if endpt.URI.RawPath == "" && req.URL.RawPath != "" {
		endpt.URI.RawPath = endpt.URI.Path
	}
	req.URL.Scheme = endpt.URI.Scheme
	req.URL.Host = endpt.URI.Host
	req.URL.Path = smithyhttp.JoinPath(endpt.URI.Path, req.URL.Path)
	req.URL.RawPath = smithyhttp.JoinPath(endpt.URI.RawPath, req.URL.RawPath)
	for k := range endpt.Headers {
		req.Header.Set(k, endpt.Headers.Get(k))
	}

	return next.HandleFinalize(ctx, in)
}
