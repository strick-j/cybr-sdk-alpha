package config

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/strick-j/cybr-sdk-alpha/cybr"
	"github.com/strick-j/cybr-sdk-alpha/internal/ini"
	"github.com/strick-j/smithy-go/logging"
)

var _ subdomainProvider = (*SharedConfig)(nil)

var (
	testConfigFilename      = filepath.Join("testdata", "shared_config")
	testConfigOtherFilename = filepath.Join("testdata", "shared_config_other")
	testCredentialsFilename = filepath.Join("testdata", "shared_credentials")
)

func TestNewSharedConfig(t *testing.T) {
	cases := map[string]struct {
		ConfigFilenames      []string
		CredentialsFilenames []string
		Profile              string
		Expected             SharedConfig
		Err                  error
	}{
		"file not exist": {
			ConfigFilenames: []string{"file_not_exist"},
			Profile:         "default",
			Err:             fmt.Errorf("failed to get shared config profile"),
		},
		"default profile": {
			ConfigFilenames: []string{testConfigFilename},
			Profile:         "default",
			Expected: SharedConfig{
				Profile:   "default",
				Subdomain: "default_subdomain",
			},
		},
		"multiple config files": {
			ConfigFilenames: []string{testConfigOtherFilename, testConfigFilename},
			Profile:         "config_file_load_order",
			Expected: SharedConfig{
				Profile:   "config_file_load_order",
				Subdomain: "shared_config_subdomain",
				Credentials: cybr.Credentials{
					Username: "shared_config_username",
					Password: "shared_config_password",
					Source:   fmt.Sprintf("SharedConfigCredentials: %s", testConfigFilename),
				},
			},
		},
		"mutliple config files reverse order": {
			ConfigFilenames: []string{testConfigFilename, testConfigOtherFilename},
			Profile:         "config_file_load_order",
			Expected: SharedConfig{
				Profile:   "config_file_load_order",
				Subdomain: "shared_config_other_subdomain",
				Credentials: cybr.Credentials{
					Username: "shared_config_other_username",
					Password: "shared_config_other_password",
					Source:   fmt.Sprintf("SharedConfigCredentials: %s", testConfigOtherFilename),
				},
			},
		},
		"Invalid INI file": {
			ConfigFilenames: []string{filepath.Join("testdata", "shared_config_invalid_ini")},
			Profile:         "profile_name",
			Err: SharedConfigProfileNotExistError{
				Filename: []string{filepath.Join("testdata", "shared_config_invalid_ini")},
				Profile:  "profile_name",
				Err:      nil,
			},
		},
		"profile names are case-sensitive (Mixed)": {
			ConfigFilenames:      []string{testConfigFilename},
			CredentialsFilenames: []string{testCredentialsFilename},
			Profile:              "DoNotNormalize",
			Expected: SharedConfig{
				Profile: "DoNotNormalize",
				Credentials: cybr.Credentials{
					Username:     "DoNotNormalize_credentials_username",
					Password:     "DoNotNormalize_credentials_password",
					SessionToken: "DoNotNormalize_config_session_token",
					Source:       fmt.Sprintf("SharedConfigCredentials: %s", testCredentialsFilename),
				},
				Subdomain: "default_subdomain_1",
			},
		},
		"profile names are case-sensitive (lower)": {
			ConfigFilenames:      []string{testConfigFilename},
			CredentialsFilenames: []string{testCredentialsFilename},
			Profile:              "donotnormalize",
			Expected: SharedConfig{
				Profile: "donotnormalize",
				Credentials: cybr.Credentials{
					Username:     "donotnormalize_credentials_username",
					Password:     "donotnormalize_credentials_password",
					SessionToken: "donotnormalize_config_session_token",
					Source:       fmt.Sprintf("SharedConfigCredentials: %s", testCredentialsFilename),
				},
				Subdomain: "default_subdomain_2",
			},
		},
		"profile names are case-sensitive (upper)": {
			ConfigFilenames:      []string{testConfigFilename},
			CredentialsFilenames: []string{testCredentialsFilename},
			Profile:              "DONOTNORMALIZE",
			Expected: SharedConfig{
				Profile: "DONOTNORMALIZE",
				Credentials: cybr.Credentials{
					Username:     "DONOTNORMALIZE_credentials_username",
					Password:     "DONOTNORMALIZE_credentials_password",
					SessionToken: "DONOTNORMALIZE_config_session_token",
					Source:       fmt.Sprintf("SharedConfigCredentials: %s", testCredentialsFilename),
				},
				Subdomain: "default_subdomain_3",
			},
		},

		"merged profiles across files": {
			ConfigFilenames:      []string{testConfigFilename},
			CredentialsFilenames: []string{testCredentialsFilename},
			Profile:              "merged_profiles",
			Expected: SharedConfig{
				Profile:   "merged_profiles",
				Subdomain: "default_subdomain_credentails",
				Domain:    "default_domain_credentials",
			},
		},
		"merged profiles across config files": {
			ConfigFilenames:      []string{testConfigFilename, testConfigFilename},
			CredentialsFilenames: []string{},
			Profile:              "merged_profiles",
			Expected: SharedConfig{
				Profile:   "merged_profiles",
				Subdomain: "default_subdomain_config",
				Domain:    "default_domain_config",
			},
		},
		"merged profiles across credentials files": {
			ConfigFilenames:      []string{},
			CredentialsFilenames: []string{testCredentialsFilename, testCredentialsFilename},
			Profile:              "merged_profiles",
			Expected: SharedConfig{
				Profile:   "merged_profiles",
				Subdomain: "default_subdomain_credentails",
				Domain:    "default_domain_credentials",
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			cfg, err := LoadSharedConfigProfile(context.TODO(), c.Profile, func(o *LoadSharedConfigOptions) {
				o.ConfigFiles = c.ConfigFilenames
				if c.CredentialsFilenames != nil {
					o.CredentialsFiles = c.CredentialsFilenames
				} else {
					o.CredentialsFiles = []string{filepath.Join("testdata", "empty_creds_config")}
				}
			})
			if c.Err != nil && err != nil {
				if e, a := c.Err.Error(), err.Error(); !strings.Contains(a, e) {
					t.Errorf("expect %q to be in %q", e, a)
				}
				return
			}
			if err != nil {
				t.Fatalf("expect no error, got %v", err)
			}
			if c.Err != nil {
				t.Errorf("expect error: %v, got none", c.Err)
			}
			if diff := cmp.Diff(c.Expected, cfg); len(diff) > 0 {
				t.Error(diff)
			}
		})
	}
}

func TestLoadSharedConfigFromSection(t *testing.T) {
	filename := testConfigFilename
	sections, err := ini.OpenFile(filename)

	if err != nil {
		t.Fatalf("failed to load test config file, %s, %v", filename, err)
	}
	cases := map[string]struct {
		Profile  string
		Expected SharedConfig
		Err      error
	}{
		"Default as profile": {
			Profile:  "default",
			Expected: SharedConfig{Subdomain: "default_subdomain"},
		},
		"prefixed profile": {
			Profile:  "profile alt_profile_name",
			Expected: SharedConfig{Subdomain: "alt_profile_name_subdomain"},
		},
		"prefixed profile 2": {
			Profile:  "profile short_profile_name_first",
			Expected: SharedConfig{Subdomain: "short_profile_name_first_alt"},
		},
		"profile with partial creds": {
			Profile:  "profile partial_creds",
			Expected: SharedConfig{},
		},
		"profile with complete creds": {
			Profile: "profile complete_creds",
			Expected: SharedConfig{
				Credentials: cybr.Credentials{
					Username: "complete_creds_username",
					Password: "complete_creds_password",
					Source:   fmt.Sprintf("SharedConfigCredentials: %s", filename),
				},
			},
		},
		"profile with complete creds and token": {
			Profile: "profile complete_creds_with_token",
			Expected: SharedConfig{
				Credentials: cybr.Credentials{
					Username:     "complete_creds_with_token_username",
					Password:     "complete_creds_with_token_password",
					SessionToken: "complete_creds_with_token_token",
					Source:       fmt.Sprintf("SharedConfigCredentials: %s", filename),
				},
			},
		},
		"complete profile": {
			Profile: "profile full_profile",
			Expected: SharedConfig{
				Credentials: cybr.Credentials{
					Username: "full_profile_username",
					Password: "full_profile_password",
					Source:   fmt.Sprintf("SharedConfigCredentials: %s", filename),
				},
				Subdomain: "full_profile_subdomain",
			},
		},
		"does not exist": {
			Profile: "does_not_exist",
			Err: SharedConfigProfileNotExistError{
				Filename: []string{filename},
				Profile:  "does_not_exist",
				Err:      nil,
			},
		},
		"profile with mixed casing": {
			Profile: "profile with_mixed_case_keys",
			Expected: SharedConfig{
				Credentials: cybr.Credentials{
					Username: "username",
					Password: "password",
					Source:   fmt.Sprintf("SharedConfigCredentials: %s", filename),
				},
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			var cfg SharedConfig

			section, ok := sections.GetSection(c.Profile)
			if !ok {
				if c.Err == nil {
					t.Fatalf("expected section to be present, was not")
				} else {
					if e, a := c.Err.Error(), "failed to get shared config profile"; !strings.Contains(e, a) {
						t.Fatalf("expect %q to be in %q", a, e)
					}
					return
				}
			}

			err := cfg.setFromIniSection(c.Profile, section)
			if c.Err != nil {
				if e, a := c.Err.Error(), err.Error(); !strings.Contains(a, e) {
					t.Errorf("expect %q to be in %q", e, a)
				}
				return
			}
			if err != nil {
				t.Fatalf("expect no error, got %v", err)
			}

			if diff := cmp.Diff(c.Expected, cfg); diff != "" {
				t.Errorf("expect shared config match\n%s", diff)
			}
		})
	}
}

func TestLoadSharedConfig(t *testing.T) {
	origProf := defaultSharedConfigProfile
	origConfigFiles := DefaultSharedConfigFiles
	origCredentialFiles := DefaultSharedCredentialsFiles
	defer func() {
		defaultSharedConfigProfile = origProf
		DefaultSharedConfigFiles = origConfigFiles
		DefaultSharedCredentialsFiles = origCredentialFiles
	}()

	cases := []struct {
		LoadOptionFn func(*LoadOptions) error
		Files        []string
		Profile      string
		LoadFn       func(context.Context, configs) (Config, error)
		Expect       SharedConfig
		Err          string
	}{
		{
			LoadOptionFn: WithSharedConfigProfile("alt_profile_name"),
			Files: []string{
				filepath.Join("testdata", "shared_config"),
			},
			LoadFn: loadSharedConfig,
			Expect: SharedConfig{
				Profile:   "alt_profile_name",
				Subdomain: "alt_profile_name_subdomain",
			},
		},
		{
			LoadOptionFn: WithSharedConfigFiles([]string{
				filepath.Join("testdata", "shared_config"),
			}),
			Profile: "alt_profile_name",
			LoadFn:  loadSharedConfig,
			Expect: SharedConfig{
				Profile:   "alt_profile_name",
				Subdomain: "alt_profile_name_subdomain",
			},
		},
		{
			LoadOptionFn: WithSharedConfigProfile("default"),
			Files: []string{
				filepath.Join("file_not_exist"),
			},
			LoadFn: loadSharedConfig,
			Err:    "failed to get shared config profile",
		},
		{
			LoadOptionFn: WithSharedConfigProfile("profile_not_exist"),
			Files: []string{
				filepath.Join("testdata", "shared_config"),
			},
			LoadFn: loadSharedConfig,
			Err:    "failed to get shared config profile",
		},
		{
			LoadOptionFn: WithSharedConfigProfile("default"),
			Files: []string{
				filepath.Join("file_not_exist"),
			},
			LoadFn: loadSharedConfigIgnoreNotExist,
		},
		{
			LoadOptionFn: WithSharedConfigProfile("assume_role_invalid_source_profile"),
			Files: []string{
				testConfigOtherFilename, testConfigFilename,
			},
			LoadFn: loadSharedConfig,
			Err:    "failed to get shared config profile",
		},
		{
			LoadOptionFn: WithSharedConfigProfile("assume_role_invalid_source_profile"),
			Files: []string{
				testConfigOtherFilename, testConfigFilename,
			},
			LoadFn: loadSharedConfigIgnoreNotExist,
			Err:    "failed to get shared config profile",
		},
	}

	for i, c := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			defaultSharedConfigProfile = origProf
			DefaultSharedConfigFiles = origConfigFiles
			DefaultSharedCredentialsFiles = origCredentialFiles

			if len(c.Profile) > 0 {
				defaultSharedConfigProfile = c.Profile
			}
			if len(c.Files) > 0 {
				DefaultSharedConfigFiles = c.Files
			}

			DefaultSharedCredentialsFiles = []string{}

			var options LoadOptions
			c.LoadOptionFn(&options)

			cfg, err := c.LoadFn(context.Background(), configs{options})
			if len(c.Err) > 0 {
				if err == nil {
					t.Fatalf("expected error %v, got none", c.Err)
				}
				if e, a := c.Err, err.Error(); !strings.Contains(a, e) {
					t.Fatalf("expect %q to be in %q", e, a)
				}
				return
			} else if err != nil {
				t.Fatalf("expect no error, got %v", err)
			}

			if e, a := c.Expect, cfg; !reflect.DeepEqual(e, a) {
				t.Errorf("expect %v got %v", e, a)
			}
		})
	}
}

func TestSharedConfigLoading(t *testing.T) {
	// initialize a logger
	var loggerBuf bytes.Buffer
	logger := logging.NewStandardLogger(&loggerBuf)

	cases := map[string]struct {
		LoadOptionFns []func(*LoadOptions) error
		LoadFn        func(context.Context, configs) (Config, error)
		Expect        SharedConfig
		ExpectLog     string
		Err           string
	}{
		"duplicate profiles in the configuration files": {
			LoadOptionFns: []func(*LoadOptions) error{
				WithSharedConfigProfile("duplicate-profile"),
				WithSharedConfigFiles([]string{filepath.Join("testdata", "load_config")}),
				WithSharedCredentialsFiles([]string{filepath.Join("testdata", "empty_creds_config")}),
				WithLogConfigurationWarnings(true),
				WithLogger(logger),
			},
			LoadFn: loadSharedConfig,
			Expect: SharedConfig{
				Profile: "duplicate-profile",
				Domain:  "dup-2-cyberark.cloud",
			},
			ExpectLog: "For profile: profile duplicate-profile, overriding domain value, with a domain value found in a " +
				"duplicate profile defined later in the same file",
		},

		"profile prefix not used in the configuration files": {
			LoadOptionFns: []func(*LoadOptions) error{
				WithSharedConfigProfile("no-such-profile"),
				WithSharedConfigFiles([]string{filepath.Join("testdata", "load_config")}),
				WithSharedCredentialsFiles([]string{filepath.Join("testdata", "empty_creds_config")}),
				WithLogConfigurationWarnings(true),
				WithLogger(logger),
			},
			LoadFn: loadSharedConfig,
			Expect: SharedConfig{},
			Err:    "failed to get shared config profile",
		},

		"profile prefix overrides default": {
			LoadOptionFns: []func(*LoadOptions) error{
				WithSharedConfigFiles([]string{filepath.Join("testdata", "load_config")}),
				WithSharedCredentialsFiles([]string{filepath.Join("testdata", "empty_creds_config")}),
				WithLogConfigurationWarnings(true),
				WithLogger(logger),
			},
			LoadFn: loadSharedConfig,
			Expect: SharedConfig{
				Profile: "default",
				Domain:  "def-cyberark.cloud",
			},
			ExpectLog: "non-default profile not prefixed with `profile `",
		},

		"duplicate profiles in credentials file": {
			LoadOptionFns: []func(*LoadOptions) error{
				WithSharedConfigProfile("duplicate-profile"),
				WithSharedConfigFiles([]string{filepath.Join("testdata", "empty_creds_config")}),
				WithSharedCredentialsFiles([]string{filepath.Join("testdata", "load_credentials")}),
				WithLogConfigurationWarnings(true),
				WithLogger(logger),
			},
			LoadFn: loadSharedConfig,
			Expect: SharedConfig{
				Profile: "duplicate-profile",
				Domain:  "dup-2-cyberark.cloud",
			},
			ExpectLog: "overriding domain value, with a domain value found in a duplicate profile defined later in the same file",
			Err:       "",
		},

		"profile prefix used in credentials files": {
			LoadOptionFns: []func(*LoadOptions) error{
				WithSharedConfigProfile("unused-profile"),
				WithSharedConfigFiles([]string{filepath.Join("testdata", "empty_creds_config")}),
				WithSharedCredentialsFiles([]string{filepath.Join("testdata", "load_credentials")}),
				WithLogConfigurationWarnings(true),
				WithLogger(logger),
			},
			LoadFn:    loadSharedConfig,
			ExpectLog: "profile defined with name `profile unused-profile` is ignored.",
			Err:       "failed to get shared config profile, unused-profile",
		},
		"partial credentials in configuration files": {
			LoadOptionFns: []func(*LoadOptions) error{
				WithSharedConfigProfile("partial-creds-1"),
				WithSharedConfigFiles([]string{filepath.Join("testdata", "load_config")}),
				WithSharedCredentialsFiles([]string{filepath.Join("testdata", "empty_creds_config")}),
				WithLogConfigurationWarnings(true),
				WithLogger(logger),
			},
			LoadFn: loadSharedConfig,
			Expect: SharedConfig{
				Profile: "partial-creds-1",
			},
			Err: "partial credentials",
		},
		"parital credentials in the credentials files": {
			LoadOptionFns: []func(*LoadOptions) error{
				WithSharedConfigProfile("partial-creds-1"),
				WithSharedConfigFiles([]string{filepath.Join("testdata", "empty_creds_config")}),
				WithSharedCredentialsFiles([]string{filepath.Join("testdata", "load_credentials")}),
				WithLogConfigurationWarnings(true),
				WithLogger(logger),
			},
			LoadFn: loadSharedConfig,
			Expect: SharedConfig{
				Profile: "partial-creds-1",
			},
			Err: "partial credentials found for profile partial-creds-1",
		},
		"credentials override configuration profile": {
			LoadOptionFns: []func(*LoadOptions) error{
				WithSharedConfigProfile("complete"),
				WithSharedConfigFiles([]string{filepath.Join("testdata", "load_config")}),
				WithSharedCredentialsFiles([]string{filepath.Join("testdata", "load_credentials")}),
				WithLogConfigurationWarnings(true),
				WithLogger(logger),
			},
			LoadFn: loadSharedConfig,
			Expect: SharedConfig{
				Profile: "complete",
				Credentials: cybr.Credentials{
					Username: "credsUsername",
					Password: "credsPassword",
					Source: fmt.Sprintf("SharedConfigCredentials: %v",
						filepath.Join("testdata", "load_credentials")),
				},
				Domain: "comp-cyberark.cloud",
			},
		},
		"credentials profile has complete credentials": {
			LoadOptionFns: []func(*LoadOptions) error{
				WithSharedConfigProfile("complete"),
				WithSharedConfigFiles([]string{filepath.Join("testdata", "empty_creds_config")}),
				WithSharedCredentialsFiles([]string{filepath.Join("testdata", "load_credentials")}),
				WithLogConfigurationWarnings(true),
				WithLogger(logger),
			},
			LoadFn: loadSharedConfig,
			Expect: SharedConfig{
				Profile: "complete",
				Credentials: cybr.Credentials{
					Username: "credsUsername",
					Password: "credsPassword",
					Source:   fmt.Sprintf("SharedConfigCredentials: %v", filepath.Join("testdata", "load_credentials")),
				},
			},
		},
		"credentials split between multiple credentials files": {
			LoadOptionFns: []func(*LoadOptions) error{
				WithSharedConfigProfile("partial-creds-1"),
				WithSharedConfigFiles([]string{filepath.Join("testdata", "empty_creds_config")}),
				WithSharedCredentialsFiles([]string{
					filepath.Join("testdata", "load_credentials"),
					filepath.Join("testdata", "load_credentials_secondary"),
				}),
				WithLogConfigurationWarnings(true),
				WithLogger(logger),
			},
			LoadFn: loadSharedConfig,
			Expect: SharedConfig{
				Profile: "partial-creds-1",
			},
			Err: "partial credentials",
		},
		"credentials split between multiple configuration files": {
			LoadOptionFns: []func(*LoadOptions) error{
				WithSharedConfigProfile("partial-creds-1"),
				WithSharedCredentialsFiles([]string{filepath.Join("testdata", "empty_creds_config")}),
				WithSharedConfigFiles([]string{
					filepath.Join("testdata", "load_config"),
					filepath.Join("testdata", "load_config_secondary"),
				}),
				WithLogConfigurationWarnings(true),
				WithLogger(logger),
			},
			LoadFn: loadSharedConfig,
			Expect: SharedConfig{
				Profile: "partial-creds-1",
				Domain:  "part-cyberark.cloud",
			},
			ExpectLog: "",
			Err:       "partial credentials",
		},
		"credentials split between credentials and config files": {
			LoadOptionFns: []func(*LoadOptions) error{
				WithSharedConfigProfile("partial-creds-1"),
				WithSharedConfigFiles([]string{
					filepath.Join("testdata", "load_config"),
				}),
				WithSharedCredentialsFiles([]string{
					filepath.Join("testdata", "load_credentials"),
				}),
				WithLogConfigurationWarnings(true),
				WithLogger(logger),
			},
			LoadFn: loadSharedConfig,
			Expect: SharedConfig{
				Profile: "partial-creds-1",
			},
			ExpectLog: "",
			Err:       "partial credentials",
		},
		"replaced profile with prefixed profile in config files": {
			LoadOptionFns: []func(*LoadOptions) error{
				WithSharedConfigProfile("replaced-profile"),
				WithSharedConfigFiles([]string{
					filepath.Join("testdata", "load_config"),
				}),
				WithSharedCredentialsFiles([]string{
					filepath.Join("testdata", "empty_creds_config"),
				}),
				WithLogConfigurationWarnings(true),
				WithLogger(logger),
			},
			LoadFn: loadSharedConfig,
			Expect: SharedConfig{
				Profile: "replaced-profile",
				Domain:  "rep-cyberark.cloud",
			},
			ExpectLog: "non-default profile not prefixed with `profile `",
		},
		"replaced profile with prefixed profile in credentials files": {
			LoadOptionFns: []func(*LoadOptions) error{
				WithSharedConfigProfile("replaced-profile"),
				WithSharedCredentialsFiles([]string{
					filepath.Join("testdata", "load_credentials"),
				}),
				WithSharedConfigFiles([]string{
					filepath.Join("testdata", "empty_creds_config"),
				}),
				WithLogConfigurationWarnings(true),
				WithLogger(logger),
			},
			LoadFn: loadSharedConfig,
			Expect: SharedConfig{
				Profile: "replaced-profile",
				Domain:  "rep-cyberark.cloud",
			},
			ExpectLog: "profile defined with name `profile replaced-profile` is ignored.",
		},
		"ignored profiles w/o prefixed profile across credentials and config files": {
			LoadOptionFns: []func(*LoadOptions) error{
				WithSharedConfigProfile("replaced-profile"),
				WithSharedCredentialsFiles([]string{
					filepath.Join("testdata", "load_credentials"),
				}),
				WithSharedConfigFiles([]string{
					filepath.Join("testdata", "load_config"),
				}),
				WithLogConfigurationWarnings(true),
				WithLogger(logger),
			},
			LoadFn: loadSharedConfig,
			Expect: SharedConfig{
				Profile: "replaced-profile",
				Domain:  "rep-cyberark.cloud",
			},
			ExpectLog: "profile defined with name `profile replaced-profile` is ignored.",
		},
		"1. profile with name as `profile` in config file": {
			LoadOptionFns: []func(*LoadOptions) error{
				WithSharedConfigProfile("profile"),
				WithSharedCredentialsFiles([]string{
					filepath.Join("testdata", "empty_creds_config"),
				}),
				WithSharedConfigFiles([]string{
					filepath.Join("testdata", "load_config"),
				}),
				WithLogConfigurationWarnings(true),
				WithLogger(logger),
			},
			LoadFn:    loadSharedConfig,
			Err:       "failed to get shared config profile, profile",
			ExpectLog: "profile defined with name `profile` is ignored",
		},
		"2. profile with name as `profile ` in config file": {
			LoadOptionFns: []func(*LoadOptions) error{
				WithSharedConfigProfile("profile "),
				WithSharedCredentialsFiles([]string{
					filepath.Join("testdata", "empty_creds_config"),
				}),
				WithSharedConfigFiles([]string{
					filepath.Join("testdata", "load_config"),
				}),
				WithLogConfigurationWarnings(true),
				WithLogger(logger),
			},
			LoadFn:    loadSharedConfig,
			Err:       "failed to get shared config profile, profile",
			ExpectLog: "profile defined with name `profile` is ignored",
		},
		"3. profile with name as `profile\t` in config file": {
			LoadOptionFns: []func(*LoadOptions) error{
				WithSharedConfigProfile("profile"),
				WithSharedCredentialsFiles([]string{
					filepath.Join("testdata", "empty_creds_config"),
				}),
				WithSharedConfigFiles([]string{
					filepath.Join("testdata", "load_config"),
				}),
				WithLogConfigurationWarnings(true),
				WithLogger(logger),
			},
			LoadFn:    loadSharedConfig,
			Err:       "failed to get shared config profile, profile",
			ExpectLog: "profile defined with name `profile` is ignored",
		},
		"profile with tabs as delimiter for profile prefix in config file": {
			LoadOptionFns: []func(*LoadOptions) error{
				WithSharedConfigProfile("with-tab"),
				WithSharedCredentialsFiles([]string{
					filepath.Join("testdata", "empty_creds_config"),
				}),
				WithSharedConfigFiles([]string{
					filepath.Join("testdata", "load_config"),
				}),
				WithLogConfigurationWarnings(true),
				WithLogger(logger),
			},
			LoadFn: loadSharedConfig,
			Expect: SharedConfig{
				Profile: "with-tab",
				Domain:  "tab-1-cyberark.cloud",
			},
		},
		"profile with tabs as delimiter for profile prefix in credentials file": {
			LoadOptionFns: []func(*LoadOptions) error{
				WithSharedConfigProfile("with-tab"),
				WithSharedCredentialsFiles([]string{
					filepath.Join("testdata", "load_credentials"),
				}),
				WithSharedConfigFiles([]string{
					filepath.Join("testdata", "empty_creds_config"),
				}),
				WithLogConfigurationWarnings(true),
				WithLogger(logger),
			},
			LoadFn:    loadSharedConfig,
			Err:       "failed to get shared config profile, with-tab",
			ExpectLog: "profile defined with name `profile with-tab` is ignored",
		},
		"profile with name as `profile profile` in credentials file": {
			LoadOptionFns: []func(*LoadOptions) error{
				WithSharedConfigProfile("profile"),
				WithSharedCredentialsFiles([]string{
					filepath.Join("testdata", "load_credentials"),
				}),
				WithSharedConfigFiles([]string{
					filepath.Join("testdata", "empty_creds_config"),
				}),
				WithLogConfigurationWarnings(true),
				WithLogger(logger),
			},
			LoadFn:    loadSharedConfig,
			Err:       "failed to get shared config profile, profile",
			ExpectLog: "profile defined with name `profile profile` is ignored",
		},
		"profile with name profile-bar in credentials file": {
			LoadOptionFns: []func(*LoadOptions) error{
				WithSharedConfigProfile("profile-bar"),
				WithSharedCredentialsFiles([]string{
					filepath.Join("testdata", "load_credentials"),
				}),
				WithSharedConfigFiles([]string{
					filepath.Join("testdata", "empty_creds_config"),
				}),
				WithLogConfigurationWarnings(true),
				WithLogger(logger),
			},
			LoadFn: loadSharedConfig,
			Expect: SharedConfig{
				Profile: "profile-bar",
				Domain:  "bar-cyberark.cloud",
			},
		},
		"profile with name profile-bar in config file": {
			LoadOptionFns: []func(*LoadOptions) error{
				WithSharedConfigProfile("profile-bar"),
				WithSharedCredentialsFiles([]string{
					filepath.Join("testdata", "empty_creds_config"),
				}),
				WithSharedConfigFiles([]string{
					filepath.Join("testdata", "load_config"),
				}),
				WithLogConfigurationWarnings(true),
				WithLogger(logger),
			},
			LoadFn:    loadSharedConfig,
			Err:       "failed to get shared config profile, profile-bar",
			ExpectLog: "profile defined with name `profile-bar` is ignored",
		},
		"profile ignored in credentials and config file": {
			LoadOptionFns: []func(*LoadOptions) error{
				WithSharedConfigProfile("ignored-profile"),
				WithSharedCredentialsFiles([]string{
					filepath.Join("testdata", "load_credentials"),
				}),
				WithSharedConfigFiles([]string{
					filepath.Join("testdata", "load_config"),
				}),
				WithLogConfigurationWarnings(true),
				WithLogger(logger),
			},
			LoadFn:    loadSharedConfig,
			Err:       "failed to get shared config profile, ignored-profile",
			ExpectLog: "profile defined with name `ignored-profile` is ignored.",
		},
		"config and creds files explicitly set to empty slice": {
			LoadOptionFns: []func(*LoadOptions) error{
				WithSharedCredentialsFiles([]string{}),
				WithSharedConfigFiles([]string{}),
				WithLogConfigurationWarnings(true),
				WithLogger(logger),
			},
			LoadFn: loadSharedConfig,
			Err:    "failed to get shared config profile, default",
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			defer loggerBuf.Reset()

			var options LoadOptions

			for _, fn := range c.LoadOptionFns {
				fn(&options)
			}

			cfg, err := c.LoadFn(context.Background(), configs{options})

			if e, a := c.ExpectLog, loggerBuf.String(); !strings.Contains(a, e) {
				t.Errorf("expect %v logged in %v", e, a)
			}
			if loggerBuf.Len() == 0 && len(c.ExpectLog) != 0 {
				t.Errorf("expected log, got none")
			}

			if len(c.Err) > 0 {
				if err == nil {
					t.Fatalf("expected error %v, got none", c.Err)
				}
				if e, a := c.Err, err.Error(); !strings.Contains(a, e) {
					t.Fatalf("expect %q to be in %q", e, a)
				}
				return
			} else if err != nil {
				t.Fatalf("expect no error, got %v", err)
			}

			if diff := cmp.Diff(c.Expect, cfg); diff != "" {
				t.Errorf("expect shared config match\n%s", diff)
			}
		})
	}
}
