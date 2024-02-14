package config

import (
	"context"
	"os"

	"github.com/strick-j/cybr-sdk-alpha/cybr"
	"github.com/strick-j/smithy-go/logging"
)

// resolveDefaultCYBRConfig will write default configuration values into the cfg
// value. It will write the default values, overwriting any previous value.
//
// This should be used as the first resolver in the slice of resolvers when
// resolving external configuration.
func resolveDefaultCYBRConfig(ctx context.Context, cfg *cybr.Config, cfgs configs) error {
	var sources []interface{}
	for _, s := range cfgs {
		sources = append(sources, s)
	}

	*cfg = cybr.Config{
		Logger:        logging.NewStandardLogger(os.Stderr),
		ConfigSources: sources,
	}
	return nil
}

// resolveDefaultSubdomain extracts the first instance of a Subdomain and sets `cybr.Config.Subdomain` to the
// default Subdomain if domain had not been resolved from other sources.
func resolveDefaultSubdomain(ctx context.Context, cfg *cybr.Config, configs configs) error {
	if len(cfg.Domain) > 0 {
		return nil
	}

	v, found, err := getDefaultSubdomain(ctx, configs)
	if err != nil {
		return err
	}
	if !found {
		return nil
	}

	cfg.SubDomain = v

	return nil
}

// resolveSubdomain extracts the first instance of a Subdomain from the configs slice.
//
// Config providers used:
// * subdomainProvider
func resolveSubdomain(ctx context.Context, cfg *cybr.Config, configs configs) error {
	v, found, err := getSubdomain(ctx, configs)
	if err != nil {
		// TODO error handling, What is the best way to handle this?
		// capture previous errors continue. error out if all errors
		return err
	}
	if !found {
		return nil
	}

	cfg.SubDomain = v
	return nil
}

// resolveDefaultDomain extracts the first instance of a Domain and sets `cybr.Config.Domain` to the
// default domain if domain had not been resolved from other sources.
func resolveDefaultDomain(ctx context.Context, cfg *cybr.Config, configs configs) error {
	if len(cfg.Domain) > 0 {
		return nil
	}

	v, found, err := getDefaultDomain(ctx, configs)
	if err != nil {
		return err
	}
	if !found {
		return nil
	}

	cfg.Domain = v

	return nil
}

// resolveDomain extracts the first instance of a Domain from the configs slice.
//
// Config providers used:
// * domainProvider
func resolveDomain(ctx context.Context, cfg *cybr.Config, configs configs) error {
	v, found, err := getDomain(ctx, configs)
	if err != nil {
		// TODO error handling, What is the best way to handle this?
		// capture previous errors continue. error out if all errors
		return err
	}
	if !found {
		return nil
	}

	cfg.Domain = v
	return nil
}

// resolveEndpointResolver extracts the first instance of a EndpointResolverFunc from the config slice
// and sets the functions result on the cybr.Config.EndpointResolver
func resolveEndpointResolverWithOptions(ctx context.Context, cfg *cybr.Config, configs configs) error {
	endpointResolver, found, err := getEndpointResolverWithOptions(ctx, configs)
	if err != nil {
		return err
	}
	if !found {
		return nil
	}

	cfg.EndpointResolverWithOptions = endpointResolver

	return nil
}

func resolveLogger(ctx context.Context, cfg *cybr.Config, configs configs) error {
	logger, found, err := getLogger(ctx, configs)
	if err != nil {
		return err
	}
	if !found {
		return nil
	}

	cfg.Logger = logger

	return nil
}

func resolveClientLogMode(ctx context.Context, cfg *cybr.Config, configs configs) error {
	mode, found, err := getClientLogMode(ctx, configs)
	if err != nil {
		return err
	}
	if !found {
		return nil
	}

	cfg.ClientLogMode = mode

	return nil
}

// resolveHTTPClient extracts the first instance of a HTTPClient and sets `cybr.Config.HTTPClient` to the HTTPClient instance
// if one has not been resolved from other sources.
func resolveHTTPClient(ctx context.Context, cfg *cybr.Config, configs configs) error {
	c, found, err := getHTTPClient(ctx, configs)
	if err != nil {
		return err
	}
	if !found {
		return nil
	}

	cfg.HTTPClient = c
	return nil
}

// resolveAPIOptions extracts the first instance of APIOptions and sets `aws.Config.APIOptions` to the resolved API options
// if one has not been resolved from other sources.
func resolveAPIOptions(ctx context.Context, cfg *cybr.Config, configs configs) error {
	o, found, err := getAPIOptions(ctx, configs)
	if err != nil {
		return err
	}
	if !found {
		return nil
	}

	cfg.APIOptions = o

	return nil
}
