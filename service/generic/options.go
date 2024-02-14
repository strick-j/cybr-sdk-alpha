package generic

import (
	"net/http"

	"github.com/strick-j/cybr-sdk-alpha/cybr"
	"github.com/strick-j/smithy-go/logging"
	"github.com/strick-j/smithy-go/middleware"
)

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

type Options struct {
	// Set of options to modify how an operation is invoked. These apply to all
	// operations invoked for this client. Use functional options on operation call to
	// modify this list for per operation behavior.
	APIOptions []func(*middleware.Stack) error

	// The Domain to use for the API client.
	Domain string

	// The Subdomain to use for the API client.
	Subdomain string

	// The logger writer interface to write logging messages to.
	Logger logging.Logger

	// Configures the events that will be sent to the configured logger.
	ClientLogMode cybr.ClientLogMode

	// Resolves the endpoint used for a particular service operation. This should be
	// used over the deprecated EndpointResolver.
	EndpointResolverV2 EndpointResolverV2

	// The HTTP client to invoke API calls with. Defaults to client's default HTTP
	// implementation if nil.
	HTTPClient HTTPClient
}

// Copy creates a clone where the APIOptions list is deep copied.
func (o Options) Copy() Options {
	to := o
	to.APIOptions = make([]func(*middleware.Stack) error, len(o.APIOptions))
	copy(to.APIOptions, o.APIOptions)

	return to
}

// WithAPIOptions returns a functional option for setting the Client's APIOptions
// option.
func WithAPIOptions(optFns ...func(*middleware.Stack) error) func(*Options) {
	return func(o *Options) {
		o.APIOptions = append(o.APIOptions, optFns...)
	}
}

// WithEndpointResolverV2 returns a functional option for setting the Client's
// EndpointResolverV2 option.
func WithEndpointResolverV2(v EndpointResolverV2) func(*Options) {
	return func(o *Options) {
		o.EndpointResolverV2 = v
	}
}
