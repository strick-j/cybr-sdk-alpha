package endpoints

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/strick-j/smithy-go/logging"

	"github.com/strick-j/cybr-sdk-alpha/cybr"
)

// DefaultKey is a compound map key of a variant and other values.
type DefaultKey struct {
	Variant        EndpointVariant
	ServiceVariant ServiceVariant
}

// EndpointKey is a compound map key of a region and associated variant value.
type EndpointKey struct {
	Domain         string
	Subdomain      string
	Variant        EndpointVariant
	ServiceVariant ServiceVariant
}

// EndpointVariant is a bit field to describe the endpoints attributes.
type EndpointVariant uint64

// ServiceVariant is a bit field to describe the service endpoint attributes.
type ServiceVariant uint64

const (
	defaultProtocol = "https"
)

var (
	protocolPriority = []string{"https", "http"}
)

// Options provide configuration needed to direct how endpoints are resolved.
type Options struct {
	// Logger is a logging implementation that log events should be sent to.
	Logger logging.Logger

	// LogDeprecated indicates that deprecated endpoints should be logged to the provided logger.
	LogDeprecated bool

	// ResolvedDomain is the resolved region string. If provided (non-zero length) it takes priority
	// over the domain name passed to the ResolveEndpoint call.
	ResolvedDomain string

	// ResolvedSubomain is the resolved region string. If provided (non-zero length) it takes priority
	// over the subdomain name passed to the ResolveEndpoint call.
	ResolvedSubdomain string

	// Disable usage of HTTPS (TLS / SSL)
	DisableHTTPS bool

	// ServiceVariant is a bitfield of service specified endpoint variant data.
	ServiceVariant ServiceVariant
}

// Partitions is a slice of partition
type Partitions []Partition

// ResolveEndpoint resolves a service endpoint for the given domain and options.
func (ps Partitions) ResolveEndpoint(domain string, opts Options) (cybr.Endpoint, error) {
	if len(ps) == 0 {
		return cybr.Endpoint{}, fmt.Errorf("no partitions found")
	}

	if opts.Logger == nil {
		opts.Logger = logging.Nop{}
	}

	if len(opts.ResolvedDomain) > 0 {
		domain = opts.ResolvedDomain
	}

	for i := 0; i < len(ps); i++ {
		if !ps[i].canResolveEndpoint(domain, opts) {
			continue
		}

		return ps[i].ResolveEndpoint(domain, opts)
	}

	// fallback to first partition format to use when resolving the endpoint.
	return ps[0].ResolveEndpoint(domain, opts)
}

func (p Partition) endpointForDomain(domain string, serviceVariant ServiceVariant, endpoints Endpoints) Endpoint {
	key := EndpointKey{
		Domain: domain,
	}

	if e, ok := endpoints[key]; ok {
		return e
	}

	// Unable to find any matching endpoint, return
	// blank that will be used for generic endpoint creation.
	return Endpoint{}
}

// Partition is an CYBR partition description for a service and its' domain endpoints.
type Partition struct {
	ID                string
	DomainRegex       *regexp.Regexp
	SubdomainRegex    *regexp.Regexp
	PartitionEndpoint string
	IsRegionalized    bool
	Defaults          map[DefaultKey]Endpoint
	Endpoints         Endpoints
}

func (p Partition) canResolveEndpoint(domain string, opts Options) bool {
	_, ok := p.Endpoints[EndpointKey{
		Domain: domain,
	}]
	return ok || p.DomainRegex.MatchString(domain)
}

// ResolveEndpoint resolves and service endpoint for the given region and options.
func (p Partition) ResolveEndpoint(domain string, options Options) (resolved cybr.Endpoint, err error) {
	if len(domain) == 0 && len(p.PartitionEndpoint) != 0 {
		domain = p.PartitionEndpoint
	}

	endpoints := p.Endpoints

	serviceVariant := options.ServiceVariant

	defaults := p.Defaults[DefaultKey{
		ServiceVariant: serviceVariant,
	}]

	return p.endpointForDomain(domain, serviceVariant, endpoints).resolve(p.ID, domain, defaults, options)
}

// Endpoints is a map of service config regions to endpoints
type Endpoints map[EndpointKey]Endpoint

// CredentialScope is the credential scope of a region and service
type CredentialScope struct {
	Domain    string
	Subdomain string
	Service   string
}

// Endpoint is a service endpoint description
type Endpoint struct {
	// True if the endpoint cannot be resolved for this partition/region/service
	Unresolveable cybr.Ternary

	Hostname  string
	Protocols []string

	CredentialScope CredentialScope

	// Indicates that this endpoint is deprecated.
	Deprecated cybr.Ternary
}

// IsZero returns whether the endpoint structure is an empty (zero) value.
func (e Endpoint) IsZero() bool {
	switch {
	case e.Unresolveable != cybr.UnknownTernary:
		return false
	case len(e.Hostname) != 0:
		return false
	case len(e.Protocols) != 0:
		return false
	case e.CredentialScope != (CredentialScope{}):
		return false
	}
	return true
}

func (e Endpoint) resolve(partition, region string, def Endpoint, options Options) (cybr.Endpoint, error) {
	var merged Endpoint
	merged.mergeIn(def)
	merged.mergeIn(e)
	e = merged

	if e.IsZero() {
		return cybr.Endpoint{}, fmt.Errorf("unable to resolve endpoint for region: %v", region)
	}

	var u string
	if e.Unresolveable != cybr.TrueTernary {
		// Only attempt to resolve the endpoint if it can be resolved.
		hostname := strings.Replace(e.Hostname, "{region}", region, 1)

		scheme := getEndpointScheme(e.Protocols, options.DisableHTTPS)
		u = scheme + "://" + hostname
	}

	if e.Deprecated == cybr.TrueTernary && options.LogDeprecated {
		options.Logger.Logf(logging.Warn, "endpoint identifier %q, url %q marked as deprecated", region, u)
	}

	return cybr.Endpoint{
		URL:         u,
		PartitionID: partition,
	}, nil
}

func (e *Endpoint) mergeIn(other Endpoint) {
	if other.Unresolveable != cybr.UnknownTernary {
		e.Unresolveable = other.Unresolveable
	}
	if len(other.Hostname) > 0 {
		e.Hostname = other.Hostname
	}
	if len(other.Protocols) > 0 {
		e.Protocols = other.Protocols
	}
	if len(other.CredentialScope.Domain) > 0 {
		e.CredentialScope.Domain = other.CredentialScope.Domain
	}
	if len(other.CredentialScope.Service) > 0 {
		e.CredentialScope.Service = other.CredentialScope.Service
	}
	if other.Deprecated != cybr.UnknownTernary {
		e.Deprecated = other.Deprecated
	}
}

func getEndpointScheme(protocols []string, disableHTTPS bool) string {
	if disableHTTPS {
		return "http"
	}

	return getByPriority(protocols, protocolPriority, defaultProtocol)
}

func getByPriority(s []string, p []string, def string) string {
	if len(s) == 0 {
		return def
	}

	for i := 0; i < len(p); i++ {
		for j := 0; j < len(s); j++ {
			if s[j] == p[i] {
				return s[j]
			}
		}
	}

	return s[0]
}
