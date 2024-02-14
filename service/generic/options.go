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
