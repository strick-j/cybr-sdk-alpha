package cybr

import (
	"net/http"

	"github.com/strick-j/smithy-go/logging"
)

// HTTPClient provides the interface to provide custom HTTPClients. Generally
// *http.Client is sufficient for most use cases. The HTTPClient should not
// follow 301 or 302 redirects.
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

// Config provides the interface to provide custom configuration.
type Config struct {
	// The sub domain to send requests too. Throws an error if not provided.
	SubDomain string

	// The domain to send requests too. Defaults to cyberark.cloud
	Domain string

	// The credentials object to use when signing requests.
	// Use the LoadDefaultConfig to load configuration from all the SDK's supported
	// sources, and resolve credentials using the SDK's default credential chain.
	Credentials CredentialsProvider

	// ConfigSources are the sources that were used to construct the Config.
	// Allows for additional configuration to be loaded by clients.
	ConfigSources []interface{}

	// The logger writer interface to write logging messages to. Defaults to
	// standard error.
	Logger logging.Logger

	// Configures the events that will be sent to the configured logger. This
	// can be used to configure the logging of signing, retries, request, and
	// responses of the SDK clients.
	//
	// See the ClientLogMode type documentation for the complete set of logging
	// modes and available configuration.
	ClientLogMode ClientLogMode

	// The HTTP Client the SDK's API clients will use to invoke HTTP requests.
	// The SDK defaults to a BuildableClient allowing API clients to create
	// copies of the HTTP Client for service specific customizations.
	//
	// Use a (*http.Client) for custom behavior. Using a custom http.Client
	// will prevent the SDK from modifying the HTTP client.
	HTTPClient HTTPClient
}

// NewConfig returns a new Config pointer that can be chained with builder
// methods to set multiple configuration values inline without using pointers.
func NewConfig() *Config {
	return &Config{}
}

// Copy will return a shallow copy of the Config object. If any additional
// configurations are provided they will be merged into the new config returned.
func (c Config) Copy() Config {
	cp := c
	return cp
}
