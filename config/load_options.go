package config

import (
	"context"

	"github.com/strick-j/cybr-sdk-alpha/cybr"
	"github.com/strick-j/cybr-sdk-alpha/cybr/logging"
)

// LoadOptionsFunc is a type alias for LoadOptions functional option
type LoadOptionsFunc func(*LoadOptions) error

// LoadOptions are discrete set of options that are valid for loading the
// configuration
type LoadOptions struct {

	// Domain is the domain to send requests to.
	Domain string

	// Subdomain is the subdomain to send requests to.
	Subdomain string

	// Credentials object to use when signing requests.
	Credentials cybr.CredentialsProvider

	// HTTPClient the SDK's API clients will use to invoke HTTP requests.
	HTTPClient HTTPClient

	// Logger writer interface to write logging messages to.
	Logger logging.Logger

	// ClientLogMode is used to configure the events that will be sent to the
	// configured logger. This can be used to configure the logging of signing,
	// retries, request, and responses of the SDK clients.
	//
	// See the ClientLogMode type documentation for the complete set of logging
	// modes and available configuration.
	ClientLogMode *cybr.ClientLogMode

	// SharedConfigProfile is the profile to be used when loading the SharedConfig
	SharedConfigProfile string

	// SharedConfigFiles is the slice of custom shared config files to use when
	// loading the SharedConfig. A non-default profile used within config file
	// must have name defined with prefix 'profile '. eg [profile xyz]
	// indicates a profile with name 'xyz'.
	//
	// If duplicate profiles are provided within the same, or across multiple
	// shared config files, the next parsed profile will override only the
	// properties that conflict with the previously defined profile. Note that
	// if duplicate profiles are provided within the SharedCredentialsFiles and
	// SharedConfigFiles, the properties defined in shared credentials file
	// take precedence.
	SharedConfigFiles []string

	// SharedCredentialsFile is the slice of custom shared credentials files to
	// use when loading the SharedConfig. The profile name used within
	// credentials file must not prefix 'profile '. eg [xyz] indicates a
	// profile with name 'xyz'. Profile declared as [profile xyz] will be
	// ignored.
	//
	// If duplicate profiles are provided with a same, or across multiple
	// shared credentials files, the next parsed profile will override only
	// properties that conflict with the previously defined profile. Note that
	// if duplicate profiles are provided within the SharedCredentialsFiles and
	// SharedConfigFiles, the properties defined in shared credentials file
	// take precedence.
	SharedCredentialsFiles []string

	// LogConfigurationWarnings when set to true, enables logging
	// configuration warnings
	LogConfigurationWarnings *bool
}

// getDomain returns Domain from config's LoadOptions
func (o LoadOptions) getDomain(ctx context.Context) (string, bool, error) {
	if len(o.Domain) == 0 {
		return "", false, nil
	}

	return o.Domain, true, nil
}

// getSubdomain returns subDomain from config's LoadOptions
func (o LoadOptions) getSubdomain(ctx context.Context) (string, bool, error) {
	if len(o.Subdomain) == 0 {
		return "", false, nil
	}

	return o.Subdomain, true, nil
}

// WithDomain is a helper function to construct functional options
// that sets Domain on config's LoadOptions. Setting the Domain to
// an empty string, will result in the Domain value being ignored.
// If multiple WithDomain calls are made, the last call overrides
// the previous call values.
func WithDomain(v string) LoadOptionsFunc {
	return func(o *LoadOptions) error {
		o.Domain = v
		return nil
	}
}

// WithSubdomain is a helper function to construct functional options
// that sets Subomain on config's LoadOptions. Setting the Subdomain to
// an empty string, will result in the Subdomain value being ignored.
// If multiple WithSubdomain calls are made, the last call overrides
// the previous call values.
func WithSubomain(v string) LoadOptionsFunc {
	return func(o *LoadOptions) error {
		o.Subdomain = v
		return nil
	}
}

// getSharedConfigProfile returns SharedConfigProfile from config's LoadOptions
func (o LoadOptions) getSharedConfigProfile(ctx context.Context) (string, bool, error) {
	if len(o.SharedConfigProfile) == 0 {
		return "", false, nil
	}

	return o.SharedConfigProfile, true, nil
}

// WithSharedConfigProfile is a helper function to construct functional options
// that sets SharedConfigProfile on config's LoadOptions. Setting the shared
// config profile to an empty string, will result in the shared config profile
// value being ignored.
// If multiple WithSharedConfigProfile calls are made, the last call overrides
// the previous call values.
func WithSharedConfigProfile(v string) LoadOptionsFunc {
	return func(o *LoadOptions) error {
		o.SharedConfigProfile = v
		return nil
	}
}

// getSharedConfigFiles returns SharedConfigFiles set on config's LoadOptions
func (o LoadOptions) getSharedConfigFiles(ctx context.Context) ([]string, bool, error) {
	if o.SharedConfigFiles == nil {
		return nil, false, nil
	}

	return o.SharedConfigFiles, true, nil
}

// WithSharedConfigFiles is a helper function to construct functional options
// that sets slice of SharedConfigFiles on config's LoadOptions.
// Setting the shared config files to an nil string slice, will result in the
// shared config files value being ignored.
// If multiple WithSharedConfigFiles calls are made, the last call overrides
// the previous call values.
func WithSharedConfigFiles(v []string) LoadOptionsFunc {
	return func(o *LoadOptions) error {
		o.SharedConfigFiles = v
		return nil
	}
}

// getSharedCredentialsFiles returns SharedCredentialsFiles set on config's LoadOptions
func (o LoadOptions) getSharedCredentialsFiles(ctx context.Context) ([]string, bool, error) {
	if o.SharedCredentialsFiles == nil {
		return nil, false, nil
	}

	return o.SharedCredentialsFiles, true, nil
}

// WithSharedCredentialsFiles is a helper function to construct functional options
// that sets slice of SharedCredentialsFiles on config's LoadOptions.
// Setting the shared credentials files to an nil string slice, will result in the
// shared credentials files value being ignored.
// If multiple WithSharedCredentialsFiles calls are made, the last call overrides
// the previous call values.
func WithSharedCredentialsFiles(v []string) LoadOptionsFunc {
	return func(o *LoadOptions) error {
		o.SharedCredentialsFiles = v
		return nil
	}
}

// getCredentialsProvider returns the credentials value
func (o LoadOptions) getCredentialsProvider(ctx context.Context) (cybr.CredentialsProvider, bool, error) {
	if o.Credentials == nil {
		return nil, false, nil
	}

	return o.Credentials, true, nil
}

// WithCredentialsProvider is a helper function to construct functional options
// that sets Credential provider value on config's LoadOptions. If credentials
// provider is set to nil, the credentials provider value will be ignored.
// If multiple WithCredentialsProvider calls are made, the last call overrides
// the previous call values.
func WithCredentialsProvider(v cybr.CredentialsProvider) LoadOptionsFunc {
	return func(o *LoadOptions) error {
		o.Credentials = v
		return nil
	}
}

func (o LoadOptions) getHTTPClient(ctx context.Context) (HTTPClient, bool, error) {
	if o.HTTPClient == nil {
		return nil, false, nil
	}

	return o.HTTPClient, true, nil
}

// WithHTTPClient is a helper function to construct functional options
// that sets HTTPClient on LoadOptions. If HTTPClient is set to nil,
// the HTTPClient value will be ignored.
// If multiple WithHTTPClient calls are made, the last call overrides
// the previous call values.
func WithHTTPClient(v HTTPClient) LoadOptionsFunc {
	return func(o *LoadOptions) error {
		o.HTTPClient = v
		return nil
	}
}

func (o LoadOptions) getLogger(ctx context.Context) (logging.Logger, bool, error) {
	if o.Logger == nil {
		return nil, false, nil
	}

	return o.Logger, true, nil
}

// WithLogger is a helper function to construct functional options
// that sets Logger on LoadOptions. If Logger is set to nil, the
// Logger value will be ignored. If multiple WithLogger calls are made,
// the last call overrides the previous call values.
func WithLogger(v logging.Logger) LoadOptionsFunc {
	return func(o *LoadOptions) error {
		o.Logger = v
		return nil
	}
}

func (o LoadOptions) getClientLogMode(ctx context.Context) (cybr.ClientLogMode, bool, error) {
	if o.ClientLogMode == nil {
		return 0, false, nil
	}

	return *o.ClientLogMode, true, nil
}

// WithClientLogMode is a helper function to construct functional options
// that sets client log mode on LoadOptions. If client log mode is set to nil,
// the client log mode value will be ignored. If multiple WithClientLogMode calls are made,
// the last call overrides the previous call values.
func WithClientLogMode(v cybr.ClientLogMode) LoadOptionsFunc {
	return func(o *LoadOptions) error {
		o.ClientLogMode = &v
		return nil
	}
}

func (o LoadOptions) getLogConfigurationWarnings(ctx context.Context) (v bool, found bool, err error) {
	if o.LogConfigurationWarnings == nil {
		return false, false, nil
	}
	return *o.LogConfigurationWarnings, true, nil
}

// WithLogConfigurationWarnings is a helper function to construct
// functional options that can be used to set LogConfigurationWarnings
// on LoadOptions.
//
// If multiple WithLogConfigurationWarnings calls are made, the last call
// overrides the previous call values.
func WithLogConfigurationWarnings(v bool) LoadOptionsFunc {
	return func(o *LoadOptions) error {
		o.LogConfigurationWarnings = &v
		return nil
	}
}
