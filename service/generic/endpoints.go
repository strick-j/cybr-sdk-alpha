package generic

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/strick-j/cybr-sdk-alpha/cybr"
	smithyendpoints "github.com/strick-j/smithy-go/endpoints"
	middleware "github.com/strick-j/smithy-go/middleware"
	"github.com/strick-j/smithy-go/ptr"
	smithyhttp "github.com/strick-j/smithy-go/transport/http"
)

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
