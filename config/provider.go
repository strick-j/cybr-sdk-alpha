package config

import (
	"context"
	"net/http"

	"github.com/strick-j/cybr-sdk-alpha/cybr/logging"
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
