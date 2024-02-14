package config

import (
	"context"
	"net/http"

	"github.com/strick-j/cybr-sdk-alpha/cybr"
	"github.com/strick-j/smithy-go/logging"
	"github.com/strick-j/smithy-go/middleware"
)

// sharedConfigProfileProvider provides access to the shared config profile
// name external configuration value.
type sharedConfigProfileProvider interface {
	getSharedConfigProfile(ctx context.Context) (string, bool, error)
}

// getSharedConfigProfile searches the configs for a sharedConfigProfileProvider
// and returns the value if found. Returns an error if a provider fails before a
// value is found.
func getSharedConfigProfile(ctx context.Context, configs configs) (value string, found bool, err error) {
	for _, cfg := range configs {
		if p, ok := cfg.(sharedConfigProfileProvider); ok {
			value, found, err = p.getSharedConfigProfile(ctx)
			if err != nil || found {
				break
			}
		}
	}
	return
}

// sharedConfigFilesProvider provides access to the shared config filesnames
// external configuration value.
type sharedConfigFilesProvider interface {
	getSharedConfigFiles(ctx context.Context) ([]string, bool, error)
}

// getSharedConfigFiles searches the configs for a sharedConfigFilesProvider
// and returns the value if found. Returns an error if a provider fails before a
// value is found.
func getSharedConfigFiles(ctx context.Context, configs configs) (value []string, found bool, err error) {
	for _, cfg := range configs {
		if p, ok := cfg.(sharedConfigFilesProvider); ok {
			value, found, err = p.getSharedConfigFiles(ctx)
			if err != nil || found {
				break
			}
		}
	}

	return
}

// sharedCredentialsFilesProvider provides access to the shared credentials filesnames
// external configuration value.
type sharedCredentialsFilesProvider interface {
	getSharedCredentialsFiles(ctx context.Context) ([]string, bool, error)
}

// getSharedCredentialsFiles searches the configs for a sharedCredentialsFilesProvider
// and returns the value if found. Returns an error if a provider fails before a
// value is found.
func getSharedCredentialsFiles(ctx context.Context, configs configs) (value []string, found bool, err error) {
	for _, cfg := range configs {
		if p, ok := cfg.(sharedCredentialsFilesProvider); ok {
			value, found, err = p.getSharedCredentialsFiles(ctx)
			if err != nil || found {
				break
			}
		}
	}

	return
}

// credentialsProviderProvider provides access to the credentials external
// configuration value.
type credentialsProviderProvider interface {
	getCredentialsProvider(ctx context.Context) (cybr.CredentialsProvider, bool, error)
}

// getCredentialsProvider searches the configs for a credentialsProviderProvider
// and returns the value if found. Returns an error if a provider fails before a
// value is found.
func getCredentialsProvider(ctx context.Context, configs configs) (p cybr.CredentialsProvider, found bool, err error) {
	for _, cfg := range configs {
		if provider, ok := cfg.(credentialsProviderProvider); ok {
			p, found, err = provider.getCredentialsProvider(ctx)
			if err != nil || found {
				break
			}
		}
	}
	return
}

// defaultSubdomainProvider is an interface for retrieving a default subdomain if a subdomain was not resolved from other sources
type defaultSubdomainProvider interface {
	getDefaultSubdomain(ctx context.Context) (string, bool, error)
}

// getDefaultSubdomain searches the slice of configs and returns the first fallback subdomain found
func getDefaultSubdomain(ctx context.Context, configs configs) (value string, found bool, err error) {
	for _, config := range configs {
		if p, ok := config.(defaultSubdomainProvider); ok {
			value, found, err = p.getDefaultSubdomain(ctx)
			if err != nil || found {
				break
			}
		}
	}
	return
}

// subdomainProvider provides access to the subdomain external configuration value.
type subdomainProvider interface {
	getSubdomain(ctx context.Context) (string, bool, error)
}

// getSubdomain searches the configs for a subdomainProvider and returns the value
// if found. Returns an error if a provider fails before a value is found.
func getSubdomain(ctx context.Context, configs configs) (value string, found bool, err error) {
	for _, cfg := range configs {
		if p, ok := cfg.(subdomainProvider); ok {
			value, found, err = p.getSubdomain(ctx)
			if err != nil || found {
				break
			}
		}
	}
	return
}

// defaultDomainProvider is an interface for retrieving a default domain if a domain was not resolved from other sources
type defaultDomainProvider interface {
	getDefaultDomain(ctx context.Context) (string, bool, error)
}

// getDefaultDomain searches the slice of configs and returns the first fallback domain found
func getDefaultDomain(ctx context.Context, configs configs) (value string, found bool, err error) {
	for _, config := range configs {
		if p, ok := config.(defaultDomainProvider); ok {
			value, found, err = p.getDefaultDomain(ctx)
			if err != nil || found {
				break
			}
		}
	}
	return
}

// domainProvider provides access to the domain external configuration value.
type domainProvider interface {
	getDomain(ctx context.Context) (string, bool, error)
}

// getDomain searches the configs for a domainProvider and returns the value
// if found. Returns an error if a provider fails before a value is found.
func getDomain(ctx context.Context, configs configs) (value string, found bool, err error) {
	for _, cfg := range configs {
		if p, ok := cfg.(domainProvider); ok {
			value, found, err = p.getDomain(ctx)
			if err != nil || found {
				break
			}
		}
	}
	return
}

type servicesObjectProvider interface {
	getServicesObject(ctx context.Context) (map[string]map[string]string, bool, error)
}

func getServicesObject(ctx context.Context, configs configs) (value map[string]map[string]string, found bool, err error) {
	for _, cfg := range configs {
		if p, ok := cfg.(servicesObjectProvider); ok {
			value, found, err = p.getServicesObject(ctx)
			if err != nil || found {
				break
			}
		}
	}
	return
}

// endpointResolverWithOptionsProvider is an interface for retrieving an aws.EndpointResolverWithOptions from a configuration source
type endpointResolverWithOptionsProvider interface {
	getEndpointResolverWithOptions(ctx context.Context) (cybr.EndpointResolverWithOptions, bool, error)
}

// getEndpointResolver searches the provided config sources for a EndpointResolverFunc that can be used
// to configure the aws.Config.EndpointResolver value.
func getEndpointResolverWithOptions(ctx context.Context, configs configs) (f cybr.EndpointResolverWithOptions, found bool, err error) {
	for _, c := range configs {
		if p, ok := c.(endpointResolverWithOptionsProvider); ok {
			f, found, err = p.getEndpointResolverWithOptions(ctx)
			if err != nil || found {
				break
			}
		}
	}
	return
}

// loggerProvider is an interface for retrieving a logging.Logger from a configuration source.
type loggerProvider interface {
	getLogger(ctx context.Context) (logging.Logger, bool, error)
}

// getLogger searches the provided config sources for a logging.Logger that can be used
// to configure the cybr.Config.Logger value.
func getLogger(ctx context.Context, configs configs) (l logging.Logger, found bool, err error) {
	for _, c := range configs {
		if p, ok := c.(loggerProvider); ok {
			l, found, err = p.getLogger(ctx)
			if err != nil || found {
				break
			}
		}
	}
	return
}

// clientLogModeProvider is an interface for retrieving the aws.ClientLogMode from a configuration source.
type clientLogModeProvider interface {
	getClientLogMode(ctx context.Context) (cybr.ClientLogMode, bool, error)
}

func getClientLogMode(ctx context.Context, configs configs) (m cybr.ClientLogMode, found bool, err error) {
	for _, c := range configs {
		if p, ok := c.(clientLogModeProvider); ok {
			m, found, err = p.getClientLogMode(ctx)
			if err != nil || found {
				break
			}
		}
	}
	return
}

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

// httpClientProvider is an interface for retrieving HTTPClient
type httpClientProvider interface {
	getHTTPClient(ctx context.Context) (HTTPClient, bool, error)
}

// getHTTPClient searches the slice of configs and returns the HTTPClient set on configs
func getHTTPClient(ctx context.Context, configs configs) (client HTTPClient, found bool, err error) {
	for _, config := range configs {
		if p, ok := config.(httpClientProvider); ok {
			client, found, err = p.getHTTPClient(ctx)
			if err != nil || found {
				break
			}
		}
	}
	return
}

// logConfigurationWarningsProvider is an configuration provider for
// retrieving a boolean indicating whether configuration issues should
// be logged when loading from config sources
type logConfigurationWarningsProvider interface {
	getLogConfigurationWarnings(ctx context.Context) (bool, bool, error)
}

func getLogConfigurationWarnings(ctx context.Context, configs configs) (v bool, found bool, err error) {
	for _, c := range configs {
		if p, ok := c.(logConfigurationWarningsProvider); ok {
			v, found, err = p.getLogConfigurationWarnings(ctx)
			if err != nil || found {
				break
			}
		}
	}
	return
}

// apiOptionsProvider is an interface for retrieving APIOptions
type apiOptionsProvider interface {
	getAPIOptions(ctx context.Context) ([]func(*middleware.Stack) error, bool, error)
}

// getAPIOptions searches the slice of configs and returns the APIOptions set on configs
func getAPIOptions(ctx context.Context, configs configs) (apiOptions []func(*middleware.Stack) error, found bool, err error) {
	for _, config := range configs {
		if p, ok := config.(apiOptionsProvider); ok {
			// retrieve APIOptions from configs and set it on cfg
			apiOptions, found, err = p.getAPIOptions(ctx)
			if err != nil || found {
				break
			}
		}
	}
	return
}
