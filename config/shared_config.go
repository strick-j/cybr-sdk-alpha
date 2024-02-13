package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/strick-j/cybr-sdk-alpha/cybr"
	"github.com/strick-j/cybr-sdk-alpha/cybr/logging"
	"github.com/strick-j/cybr-sdk-alpha/internal/ini"
	"github.com/strick-j/cybr-sdk-alpha/internal/shareddefaults"
)

const (
	// Prefix to use for filtering profiles. The profile prefix should only
	// exist in the shared config file, not the credentials file.
	profilePrefix = `profile `

	// Prefix for services section. It is referenced in profile via the services
	// parameter to configure clients for service-specific parameters.
	servicesPrefix = `services`

	sourceProfileKey = `source_profile`

	// Static Credentials group
	usernameKey     = `cybr_username`      // group required
	passwordKey     = `cybr_password`      // group required
	sessionTokenKey = `cybr_session_token` // optional

	// Domain group holds both the subdomain for the CyberArk
	// tenant and the domain. If the domain if not specifed
	// the default is cyberark.cloud.
	subdomainKey = `subdomain` // required
	domainKey    = `domain`    // optional

	// DefaultSharedConfigProfile is the default profile to be used when
	// loading configuration from the config files if another profile name
	// is not provided.
	DefaultSharedConfigProfile = `default`
)

// defaultSharedConfigProfile allows for swapping the default profile for testing
var defaultSharedConfigProfile = DefaultSharedConfigProfile

// DefaultSharedCredentialsFilename returns the SDK's default file path
// for the shared credentials file.
//
// Builds the shared config file path based on the OS's platform.
//
//   - Linux/Unix: $HOME/.cybr/credentials
//   - Windows: %USERPROFILE%\.cybr\credentials
func DefaultSharedCredentialsFilename() string {
	return filepath.Join(shareddefaults.UserHomeDir(), ".cybr", "credentials")
}

// DefaultSharedConfigFilename returns the SDK's default file path for
// the shared config file.
//
// Builds the shared config file path based on the OS's platform.
//
//   - Linux/Unix: $HOME/.cybr/config
//   - Windows: %USERPROFILE%\.cybr\config
func DefaultSharedConfigFilename() string {
	return filepath.Join(shareddefaults.UserHomeDir(), ".cybr", "config")
}

// DefaultSharedConfigFiles is a slice of the default shared config files that
// the will be used in order to load the SharedConfig.
var DefaultSharedConfigFiles = []string{
	DefaultSharedConfigFilename(),
}

// DefaultSharedCredentialsFiles is a slice of the default shared credentials
// files that the will be used in order to load the SharedConfig.
var DefaultSharedCredentialsFiles = []string{
	DefaultSharedCredentialsFilename(),
}

type SharedConfig struct {
	// Profile is the name of the profile the SDK should load the config from.
	Profile string

	// Credentials values from the config file. Both cybr_username
	// and cybr_password must be provided together in the same file
	// to be considered valid. The values will be ignored if not a complete group.
	// cybr_session_token is an optional field that can be provided if both of the
	// other two fields are also provided.
	//
	//	cybr_username
	//	cybr_password
	//	cybr_session_token
	Credentials cybr.Credentials

	SourceProfileName string
	Source            *SharedConfig

	// The subdomain of the CyberArk tenant.
	//
	// subdomain = example
	Subdomain string

	// The default domain to use when signing requests if a domain.
	//
	// domain = cyberark.cloud
	Domain string
}

// GetDomain returns the sub domain for the profile if a domain is set.
func (c SharedConfig) getDomain(ctx context.Context) (string, bool, error) {
	if len(c.Domain) == 0 {
		return "", false, nil
	}
	return c.Domain, true, nil
}

// GetSubdomain returns the sub domain for the profile if a subdomain is set.
func (c SharedConfig) getSubdomain(ctx context.Context) (string, bool, error) {
	if len(c.Subdomain) == 0 {
		return "", false, nil
	}
	return c.Subdomain, true, nil
}

// GetCredentialsProvider returns the credentials for a profile if they were set.
func (c SharedConfig) getCredentialsProvider() (cybr.Credentials, bool, error) {
	return c.Credentials, true, nil
}

// loadSharedConfigIgnoreNotExist is an alias for loadSharedConfig with the
// addition of ignoring when none of the files exist or when the profile
// is not found in any of the files.
func loadSharedConfigIgnoreNotExist(ctx context.Context, configs configs) (Config, error) {
	cfg, err := loadSharedConfig(ctx, configs)
	if err != nil {
		if _, ok := err.(SharedConfigProfileNotExistError); ok {
			return SharedConfig{}, nil
		}
		return nil, err
	}

	return cfg, nil
}

// loadSharedConfig uses the configs passed in to load the SharedConfig from file
// The file names and profile name are sourced from the configs.
//
// If profile name is not provided DefaultSharedConfigProfile (default) will
// be used.
//
// If shared config filenames are not provided DefaultSharedConfigFiles will
// be used.
//
// Config providers used:
// * sharedConfigProfileProvider
// * sharedConfigFilesProvider
func loadSharedConfig(ctx context.Context, configs configs) (Config, error) {
	var profile string
	var configFiles []string
	var credentialsFiles []string
	var ok bool
	var err error

	profile, ok, err = getSharedConfigProfile(ctx, configs)
	if err != nil {
		return nil, err
	}
	if !ok {
		profile = defaultSharedConfigProfile
	}

	configFiles, _, err = getSharedConfigFiles(ctx, configs)
	if err != nil {
		return nil, err
	}

	credentialsFiles, _, err = getSharedCredentialsFiles(ctx, configs)
	if err != nil {
		return nil, err
	}

	// setup logger if log configuration warning is set
	var logger logging.Logger
	logWarnings, found, err := getLogConfigurationWarnings(ctx, configs)
	if err != nil {
		return SharedConfig{}, err
	}
	if found && logWarnings {
		logger, found, err = getLogger(ctx, configs)
		if err != nil {
			return SharedConfig{}, err
		}
		if !found {
			logger = logging.NewStandardLogger(os.Stderr)
		}
	}

	return LoadSharedConfigProfile(ctx, profile,
		func(o *LoadSharedConfigOptions) {
			o.Logger = logger
			o.ConfigFiles = configFiles
			o.CredentialsFiles = credentialsFiles
		},
	)
}

// LoadSharedConfigOptions struct contains optional values that can be used to load the config.
type LoadSharedConfigOptions struct {

	// CredentialsFiles are the shared credentials files
	CredentialsFiles []string

	// ConfigFiles are the shared config files
	ConfigFiles []string

	// Logger is the logger used to log shared config behavior
	Logger logging.Logger
}

// LoadSharedConfigProfile retrieves the configuration from the list of files
// using the profile provided. The order the files are listed will determine
// precedence. Values in subsequent files will overwrite values defined in
// earlier files.
//
// For example, given two files A and B. Both define credentials. If the order
// of the files are A then B, B's credential values will be used instead of A's.
//
// If config files are not set, SDK will default to using a file at location `.cybr/config` if present.
// If credentials files are not set, SDK will default to using a file at location `.cybr/credentials` if present.
// No default files are set, if files set to an empty slice.
func LoadSharedConfigProfile(ctx context.Context, profile string, optFns ...func(*LoadSharedConfigOptions)) (SharedConfig, error) {
	var option LoadSharedConfigOptions
	for _, fn := range optFns {
		fn(&option)
	}

	if option.ConfigFiles == nil {
		option.ConfigFiles = DefaultSharedConfigFiles
	}

	if option.CredentialsFiles == nil {
		option.CredentialsFiles = DefaultSharedCredentialsFiles
	}

	// load shared configuration sections from shared configuration INI options
	configSections, err := loadIniFiles(option.ConfigFiles)
	if err != nil {
		return SharedConfig{}, err
	}

	// check for profile prefix and drop duplicates or invalid profiles
	err = processConfigSections(ctx, &configSections, option.Logger)
	if err != nil {
		return SharedConfig{}, err
	}

	// load shared credentials sections from shared credentials INI options
	credentialsSections, err := loadIniFiles(option.CredentialsFiles)
	if err != nil {
		return SharedConfig{}, err
	}

	// check for profile prefix and drop duplicates or invalid profiles
	err = processCredentialsSections(ctx, &credentialsSections, option.Logger)
	if err != nil {
		return SharedConfig{}, err
	}

	err = mergeSections(&configSections, credentialsSections)
	if err != nil {
		return SharedConfig{}, err
	}

	cfg := SharedConfig{}
	profiles := map[string]struct{}{}

	if err = cfg.setFromIniSections(profiles, profile, configSections, option.Logger); err != nil {
		return SharedConfig{}, err
	}

	return cfg, nil
}

func processConfigSections(ctx context.Context, sections *ini.Sections, logger logging.Logger) error {
	skipSections := map[string]struct{}{}

	for _, section := range sections.List() {
		if _, ok := skipSections[section]; ok {
			continue
		}

		// drop sections from config file that do not have expected prefixes.
		switch {
		case strings.HasPrefix(section, profilePrefix):
			// Rename sections to remove "profile " prefixing to match with
			// credentials file. If default is already present, it will be
			// dropped.
			newName, err := renameProfileSection(section, sections, logger)
			if err != nil {
				return fmt.Errorf("failed to rename profile section, %w", err)
			}
			skipSections[newName] = struct{}{}

		case strings.EqualFold(section, "default"):
		default:
			// drop this section, as invalid profile name
			sections.DeleteSection(section)

			if logger != nil {
				logger.Logf(logging.Debug, "A profile defined with name `%v` is ignored. "+
					"For use within a shared configuration file, "+
					"a non-default profile must have `profile ` "+
					"prefixed to the profile name.",
					section,
				)
			}
		}
	}
	return nil
}

func renameProfileSection(section string, sections *ini.Sections, logger logging.Logger) (string, error) {
	v, ok := sections.GetSection(section)
	if !ok {
		return "", fmt.Errorf("error processing profiles within the shared configuration files")
	}

	// delete section with profile as prefix
	sections.DeleteSection(section)

	// set the value to non-prefixed name in sections.
	section = strings.TrimPrefix(section, profilePrefix)
	if sections.HasSection(section) {
		oldSection, _ := sections.GetSection(section)
		v.Logs = append(v.Logs,
			fmt.Sprintf("A non-default profile not prefixed with `profile ` found in %s, "+
				"overriding non-default profile from %s",
				v.SourceFile, oldSection.SourceFile))
		sections.DeleteSection(section)
	}

	// assign non-prefixed name to section
	v.Name = section
	sections.SetSection(section, v)

	return section, nil
}

func processCredentialsSections(ctx context.Context, sections *ini.Sections, logger logging.Logger) error {
	for _, section := range sections.List() {
		// drop profiles with prefix for credential files
		if strings.HasPrefix(section, profilePrefix) {
			// drop this section, as invalid profile name
			sections.DeleteSection(section)

			if logger != nil {
				logger.Logf(logging.Debug,
					"The profile defined with name `%v` is ignored. A profile with the `profile ` prefix is invalid "+
						"for the shared credentials file.\n",
					section,
				)
			}
		}
	}
	return nil
}

func loadIniFiles(filenames []string) (ini.Sections, error) {
	mergedSections := ini.NewSections()

	for _, filename := range filenames {
		sections, err := ini.OpenFile(filename)
		var v *ini.UnableToReadFile
		if ok := errors.As(err, &v); ok {
			// Skip files which can't be opened and read for whatever reason.
			// We treat such files as empty, and do not fall back to other locations.
			continue
		} else if err != nil {
			return ini.Sections{}, SharedConfigLoadError{Filename: filename, Err: err}
		}

		// mergeSections into mergedSections
		err = mergeSections(&mergedSections, sections)
		if err != nil {
			return ini.Sections{}, SharedConfigLoadError{Filename: filename, Err: err}
		}
	}

	return mergedSections, nil
}

// mergeSections merges source section properties into destination section properties
func mergeSections(dst *ini.Sections, src ini.Sections) error {
	for _, sectionName := range src.List() {
		srcSection, _ := src.GetSection(sectionName)

		if (!srcSection.Has(usernameKey) && srcSection.Has(passwordKey)) ||
			(srcSection.Has(usernameKey) && !srcSection.Has(passwordKey)) {
			srcSection.Errors = append(srcSection.Errors,
				fmt.Errorf("partial credentials found for profile %v", sectionName))
		}

		if !dst.HasSection(sectionName) {
			dst.SetSection(sectionName, srcSection)
			continue
		}

		// merge with destination srcSection
		dstSection, _ := dst.GetSection(sectionName)

		// errors should be overriden if any
		dstSection.Errors = srcSection.Errors

		// Access key id update
		if srcSection.Has(usernameKey) && srcSection.Has(passwordKey) {
			userKey := srcSection.String(usernameKey)
			passKey := srcSection.String(passwordKey)

			if dstSection.Has(usernameKey) {
				dstSection.Logs = append(dstSection.Logs, newMergeKeyLogMessage(sectionName, usernameKey,
					dstSection.SourceFile[usernameKey], srcSection.SourceFile[usernameKey]))
			}

			// update username key
			v, err := ini.NewStringValue(userKey)
			if err != nil {
				return fmt.Errorf("error merging username key, %w", err)
			}
			dstSection.UpdateValue(usernameKey, v)

			// update password key
			v, err = ini.NewStringValue(passKey)
			if err != nil {
				return fmt.Errorf("error merging password key, %w", err)
			}
			dstSection.UpdateValue(passwordKey, v)

			// update session token
			if err = mergeStringKey(&srcSection, &dstSection, sectionName, sessionTokenKey); err != nil {
				return err
			}

			// update source file to reflect where the static creds came from
			dstSection.UpdateSourceFile(usernameKey, srcSection.SourceFile[usernameKey])
			dstSection.UpdateSourceFile(passwordKey, srcSection.SourceFile[passwordKey])
		}

		stringKeys := []string{
			sourceProfileKey,
			domainKey,
			subdomainKey,
		}
		for i := range stringKeys {
			if err := mergeStringKey(&srcSection, &dstSection, sectionName, stringKeys[i]); err != nil {
				return err
			}
		}

		// set srcSection on dst srcSection
		*dst = dst.SetSection(sectionName, dstSection)
	}

	return nil
}

func mergeStringKey(srcSection *ini.Section, dstSection *ini.Section, sectionName, key string) error {
	if srcSection.Has(key) {
		srcValue := srcSection.String(key)
		val, err := ini.NewStringValue(srcValue)
		if err != nil {
			return fmt.Errorf("error merging %s, %w", key, err)
		}

		if dstSection.Has(key) {
			dstSection.Logs = append(dstSection.Logs, newMergeKeyLogMessage(sectionName, key,
				dstSection.SourceFile[key], srcSection.SourceFile[key]))
		}

		dstSection.UpdateValue(key, val)
		dstSection.UpdateSourceFile(key, srcSection.SourceFile[key])
	}
	return nil
}

func newMergeKeyLogMessage(sectionName, key, dstSourceFile, srcSourceFile string) string {
	return fmt.Sprintf("For profile: %v, overriding %v value, defined in %v "+
		"with a %v value found in a duplicate profile defined at file %v. \n",
		sectionName, key, dstSourceFile, key, srcSourceFile)
}

// Returns an error if all of the files fail to load. If at least one file is
// successfully loaded and contains the profile, no error will be returned.
func (c *SharedConfig) setFromIniSections(profiles map[string]struct{}, profile string,
	sections ini.Sections, logger logging.Logger) error {
	c.Profile = profile

	section, ok := sections.GetSection(profile)
	if !ok {
		return SharedConfigProfileNotExistError{
			Profile: profile,
		}
	}

	// if logs are appended to the section, log them
	if section.Logs != nil && logger != nil {
		for _, log := range section.Logs {
			logger.Logf(logging.Debug, log)
		}
	}

	// set config from the provided INI section
	err := c.setFromIniSection(profile, section)
	if err != nil {
		return fmt.Errorf("error fetching config from profile, %v, %w", profile, err)
	}

	// if not top level profile and has credentials, return with credentials.
	if len(profiles) != 0 && c.Credentials.HasKeys() {
		return nil
	}

	profiles[profile] = struct{}{}

	// validate no colliding credentials type are present
	if err := c.validateCredentialType(); err != nil {
		return err
	}

	// Link source profiles
	if len(c.SourceProfileName) != 0 {
		// Linked profile via source_profile ignore credential provider
		// options, the source profile must provide the credentials.
		c.clearCredentialOptions()

		srcCfg := &SharedConfig{}
		err := srcCfg.setFromIniSections(profiles, c.SourceProfileName, sections, logger)
		if err != nil {
			if _, ok := err.(SharedConfigProfileNotExistError); ok {
				err = SharedConfigLinkError{
					Profile: c.SourceProfileName,
					Err:     err,
				}
			}
			return err
		}

		if !srcCfg.hasCredentials() {
			return SharedConfigLinkError{
				Profile: c.SourceProfileName,
			}
		}

		c.Source = srcCfg
	}

	return nil
}

// setFromIniSection loads the configuration from the profile section defined in
// the provided INI file. A SharedConfig pointer type value is used so that
// multiple config file loadings can be chained.
//
// Only loads complete logically grouped values, and will not set fields in cfg
// for incomplete grouped values in the config. Such as credentials. For example
// if a config file only includes aws_access_key_id but no aws_secret_access_key
// the aws_access_key_id will be ignored.
func (c *SharedConfig) setFromIniSection(profile string, section ini.Section) error {
	if len(section.Name) == 0 {
		sources := make([]string, 0)
		for _, v := range section.SourceFile {
			sources = append(sources, v)
		}

		return fmt.Errorf("parsing error : could not find profile section name after processing files: %v", sources)
	}

	if len(section.Errors) != 0 {
		var errStatement string
		for i, e := range section.Errors {
			errStatement = fmt.Sprintf("%d, %v\n", i+1, e.Error())
		}
		return fmt.Errorf("Error using profile: \n %v", errStatement)
	}

	updateString(&c.Domain, section, domainKey)
	updateString(&c.Subdomain, section, subdomainKey)
	updateString(&c.SourceProfileName, section, sourceProfileKey)

	// Shared Credentials
	creds := cybr.Credentials{
		Username: section.String(usernameKey),
		Password: section.String(passwordKey),
		Source:   fmt.Sprintf("SharedConfigCredentials: %s", section.SourceFile[usernameKey]),
	}

	if creds.HasKeys() {
		c.Credentials = creds
	}

	return nil
}

func (c *SharedConfig) validateCredentialType() error {
	// Only one or no credential type can be defined.
	if !oneOrNone(
		len(c.SourceProfileName) != 0,
	) {
		return fmt.Errorf("only one credential type may be specified per profile: source profile, credential source, credential process, web identity token")
	}

	return nil
}

func (c *SharedConfig) hasCredentials() bool {
	switch {
	case len(c.SourceProfileName) != 0:
	case c.Credentials.HasKeys():
	default:
		return false
	}

	return true
}

func (c *SharedConfig) clearCredentialOptions() {
	c.Credentials = cybr.Credentials{}
}

// SharedConfigLoadError is an error for the shared config file failed to load.
type SharedConfigLoadError struct {
	Filename string
	Err      error
}

// Unwrap returns the underlying error that caused the failure.
func (e SharedConfigLoadError) Unwrap() error {
	return e.Err
}

func (e SharedConfigLoadError) Error() string {
	return fmt.Sprintf("failed to load shared config file, %s, %v", e.Filename, e.Err)
}

// SharedConfigProfileNotExistError is an error for the shared config when
// the profile was not find in the config file.
type SharedConfigProfileNotExistError struct {
	Filename []string
	Profile  string
	Err      error
}

// Unwrap returns the underlying error that caused the failure.
func (e SharedConfigProfileNotExistError) Unwrap() error {
	return e.Err
}

func (e SharedConfigProfileNotExistError) Error() string {
	return fmt.Sprintf("failed to get shared config profile, %s", e.Profile)
}

// SharedConfigLinkError is an error for the shared config when the
// profile contains Link information, but that information is invalid
// or not complete.
type SharedConfigLinkError struct {
	Profile string
	Err     error
}

// Unwrap returns the underlying error that caused the failure.
func (e SharedConfigLinkError) Unwrap() error {
	return e.Err
}

func (e SharedConfigLinkError) Error() string {
	return fmt.Sprintf("failed to load token of profile %s, %v",
		e.Profile, e.Err)
}

func oneOrNone(bs ...bool) bool {
	var count int

	for _, b := range bs {
		if b {
			count++
			if count > 1 {
				return false
			}
		}
	}

	return true
}

// updateString will only update the dst with the value in the section key, key
// is present in the section.
func updateString(dst *string, section ini.Section, key string) {
	if !section.Has(key) {
		return
	}
	*dst = section.String(key)
}

// updateInt will only update the dst with the value in the section key, key
// is present in the section.
//
// Down casts the INI integer value from a int64 to an int, which could be
// different bit size depending on platform.
func updateInt(dst *int, section ini.Section, key string) error {
	if !section.Has(key) {
		return nil
	}

	v, ok := section.Int(key)
	if !ok {
		return fmt.Errorf("invalid value %s=%s, expect integer", key, section.String(key))
	}

	*dst = int(v)
	return nil
}
