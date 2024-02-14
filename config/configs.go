package config

import (
	"context"
	"os"

	"github.com/strick-j/cybr-sdk-alpha/cybr"
)

var defaultCYBRConfigResolvers = []cybrConfigResolver{
	resolveSubdomain,

	resolveDomain,

	// Sets the logger to be used. Could be user provided logger, and client
	// logging mode.
	resolveLogger,

	resolveClientLogMode,

	// Sets the HTTP client and configuration to use for making requests using
	// the HTTP transport.
	resolveHTTPClient,

	resolveAPIOptions,

	resolveCredentials,
}

// A Config represents a generic configuration value or set of values. This type
// will be used by the CYBRConfigResolvers to extract
//
// General the Config type will use type assertion against the Provider interfaces
// to extract specific data from the Config.
type Config interface{}

// A loader is used to load external configuration data and returns it as
// a generic Config type.
//
// The loader should return an error if it fails to load the external configuration
// or the configuration data is malformed, or required components missing.
type loader func(context.Context, configs) (Config, error)

// An cybrConfigResolver will extract configuration data from the configs slice
// using the provider interfaces to extract specific functionality. The extracted
// configuration values will be written to the CYBR Config value.
//
// The resolver should return an error if it it fails to extract the data, the
// data is malformed, or incomplete.
type cybrConfigResolver func(ctx context.Context, cfg *cybr.Config, configs configs) error

// configs is a slice of Config values. These values will be used by the
// CYBRConfigResolvers to extract external configuration values to populate the
// CYBR Config type.
//
// Use AppendFromLoaders to add additional external Config values that are
// loaded from external sources.
//
// Use ResolveCYBRConfig after external Config values have been added or loaded
// to extract the loaded configuration values into the CYBR Config.
type configs []Config

// AppendFromLoaders iterates over the slice of loaders passed in calling each
// loader function in order. The external config value returned by the loader
// will be added to the returned configs slice.
//
// If a loader returns an error this method will stop iterating and return
// that error.
func (cs configs) AppendFromLoaders(ctx context.Context, loaders []loader) (configs, error) {
	for _, fn := range loaders {
		cfg, err := fn(ctx, cs)
		if err != nil {
			return nil, err
		}

		cs = append(cs, cfg)
	}

	return cs, nil
}

// ResolveCYBRConfig returns a CYBR configuration populated with values by calling
// the resolvers slice passed in. Each resolver is called in order. Any resolver
// may overwrite the CYBR Configuration value of a previous resolver.
//
// If an resolver returns an error this method will return that error, and stop
// iterating over the resolvers.
func (cs configs) ResolveCYBRConfig(ctx context.Context, resolvers []cybrConfigResolver) (cybr.Config, error) {
	var cfg cybr.Config

	for _, fn := range resolvers {
		if err := fn(ctx, &cfg, cs); err != nil {
			return cybr.Config{}, err
		}
	}

	return cfg, nil
}

// ResolveConfig calls the provide function passing slice of configuration sources.
// This implements the aws.ConfigResolver interface.
func (cs configs) ResolveConfig(f func(configs []interface{}) error) error {
	var cfgs []interface{}
	for i := range cs {
		cfgs = append(cfgs, cs[i])
	}
	return f(cfgs)
}

func LoadDefaultConfig(ctx context.Context, optFns ...func(*LoadOptions) error) (cfg cybr.Config, err error) {
	var options LoadOptions
	for _, optFn := range optFns {
		if err := optFn(&options); err != nil {
			return cybr.Config{}, err
		}
	}

	// assign Load Options to configs
	var cfgCpy = configs{options}

	cfgCpy, err = cfgCpy.AppendFromLoaders(ctx, resolveConfigLoaders(&options))
	if err != nil {
		return cybr.Config{}, err
	}

	cfg, err = cfgCpy.ResolveCYBRConfig(ctx, defaultCYBRConfigResolvers)
	if err != nil {
		return cybr.Config{}, err
	}

	return cfg, nil
}

func resolveConfigLoaders(options *LoadOptions) []loader {
	loaders := make([]loader, 2)
	loaders[0] = loadEnvConfig

	// specification of a profile should cause a load failure if it doesn't exist
	if os.Getenv(cybrProfileEnvVar) != "" || options.SharedConfigProfile != "" {
		loaders[1] = loadSharedConfig
	} else {
		loaders[1] = loadSharedConfigIgnoreNotExist
	}

	return loaders
}
