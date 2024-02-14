package config

import (
	"context"
	"os"

	"github.com/strick-j/cybr-sdk-alpha/cybr"
)

// CredentialsSourceName provides a name of the provider when config is
// loaded from environment.
const CredentialsSourceName = "EnvConfigCredentials"

const (
	cybrUsernameEnvVar   = "CYBR_USERNAME"
	cybrUsernameIDEnvVar = "CYBR_USERNAME_ID"

	cybrPasswordEnvVar = "CYBR_PASSWORD"
	cybrSecretEnvVar   = "CYBR_SECRET"

	cybrSessionTokenEnvVar = "CYBR_SESSION_TOKEN"

	cybrSubdomainEnvVar        = "CYBR_SUBDOMAIN"
	cybrDefaultSubdomainEnvVar = "CYBR_DEFAULT_SUBDOMAIN"
	cybrDomainEnvVar           = "CYBR_DOMAIN"
	cybrDefaultDomainEnvVar    = "CYBR_DEFAULT_DOMAIN"

	cybrProfileEnvVar        = "CYBR_PROFILE"
	cybrDefaultProfileEnvVar = "CYBR_DEFAULT_PROFILE"

	cybrSharedCredentialsFileEnvVar = "CYBR_SHARED_CREDENTIALS_FILE"

	cybrConfigFileEnvVar = "CYBR_CONFIG_FILE"
)

var (
	credUsernameEnvKeys = []string{
		cybrUsernameEnvVar,
		cybrUsernameIDEnvVar,
	}
	credPasswordEnvKeys = []string{
		cybrPasswordEnvVar,
		cybrSecretEnvVar,
	}
	domainEnvKeys = []string{
		cybrDomainEnvVar,
		cybrDefaultDomainEnvVar,
	}
	subdomainEnvKeys = []string{
		cybrSubdomainEnvVar,
		cybrDefaultSubdomainEnvVar,
	}
	profileEnvKeys = []string{
		cybrProfileEnvVar,
		cybrDefaultProfileEnvVar,
	}
)

// EnvConfig is a collection of environment values the SDK will read
// setup config from. All environment values are optional. But some values
// such as credentials require multiple values to be complete or the values
// will be ignored.
type EnvConfig struct {
	Credentials cybr.Credentials

	Domain string

	Subdomain string

	SharedConfigProfile string

	SharedCredentialsFile string

	SharedConfigFile string
}

// loadEnvConfig reads configuration values from the OS's environment variables.
// Returning the a Config typed EnvConfig to satisfy the ConfigLoader func type.
func loadEnvConfig(ctx context.Context, cfgs configs) (Config, error) {
	return NewEnvConfig()
}

// NewEnvConfig retrieves the SDK's environment configuration.
// See `EnvConfig` for the values that will be retrieved.
func NewEnvConfig() (EnvConfig, error) {
	var cfg EnvConfig

	creds := cybr.Credentials{
		Source: CredentialsSourceName,
	}

	setStringFromEnvVal(&creds.Username, credUsernameEnvKeys)
	setStringFromEnvVal(&creds.Password, credPasswordEnvKeys)

	setStringFromEnvVal(&cfg.Domain, domainEnvKeys)
	setStringFromEnvVal(&cfg.Subdomain, subdomainEnvKeys)
	setStringFromEnvVal(&cfg.SharedConfigProfile, profileEnvKeys)

	cfg.SharedCredentialsFile = os.Getenv(cybrSharedCredentialsFileEnvVar)
	cfg.SharedConfigFile = os.Getenv(cybrConfigFileEnvVar)

	return cfg, nil
}

// GetDomain returns the CYBR Domain if set in the environment. Returns an empty
// string if not set.
func (c EnvConfig) getDomain(ctx context.Context) (string, bool, error) {
	if len(c.Domain) == 0 {
		return "", false, nil
	}
	return c.Domain, true, nil
}

// GetSubdomain returns the CYBR Subdomain if set in the environment. Returns an empty
// string if not set.
func (c EnvConfig) getSubdomain(ctx context.Context) (string, bool, error) {
	if len(c.Subdomain) == 0 {
		return "", false, nil
	}
	return c.Subdomain, true, nil
}

// GetSharedConfigProfile returns the shared config profile if set in the
// environment. Returns an empty string if not set.
func (c EnvConfig) getSharedConfigProfile(ctx context.Context) (string, bool, error) {
	if len(c.SharedConfigProfile) == 0 {
		return "", false, nil
	}

	return c.SharedConfigProfile, true, nil
}

// getSharedConfigFiles returns a slice of filenames set in the environment.
//
// Will return the filenames in the order of:
// * Shared Config
func (c EnvConfig) getSharedConfigFiles(context.Context) ([]string, bool, error) {
	var files []string
	if v := c.SharedConfigFile; len(v) > 0 {
		files = append(files, v)
	}

	if len(files) == 0 {
		return nil, false, nil
	}
	return files, true, nil
}

// getSharedCredentialsFiles returns a slice of filenames set in the environment.
//
// Will return the filenames in the order of:
// * Shared Credentials
func (c EnvConfig) getSharedCredentialsFiles(context.Context) ([]string, bool, error) {
	var files []string
	if v := c.SharedCredentialsFile; len(v) > 0 {
		files = append(files, v)
	}
	if len(files) == 0 {
		return nil, false, nil
	}
	return files, true, nil
}

func setStringFromEnvVal(dst *string, keys []string) {
	for _, k := range keys {
		if v := os.Getenv(k); len(v) > 0 {
			*dst = v
			break
		}
	}
}
