package endpoints

import (
	"regexp"

	"github.com/strick-j/cybr-sdk-alpha/cybr"
	"github.com/strick-j/cybr-sdk-alpha/internal/endpoints/v2"
	"github.com/strick-j/smithy-go/logging"
)

type Options struct {
	// Logger is a logging implementation that log events should be sent to.
	Logger logging.Logger

	// LogDeprecated indicates that deprecated endpoints should be logged to the
	// provided logger.
	LogDeprecated bool

	ResolvedDomain string

	ResolvedSubomain string

	// DisableHTTPS informs the resolver to return an endpoint that does not use the
	// HTTPS scheme.
	DisableHTTPS bool
}

func (o Options) GetResolvedDomain() string {
	return o.ResolvedDomain
}

func (o Options) GetResolvedSubomain() string {
	return o.ResolvedDomain
}

func (o Options) GetDisableHTTPS() bool {
	return o.DisableHTTPS
}

func transformToSharedOptions(options Options) endpoints.Options {
	return endpoints.Options{
		Logger:            options.Logger,
		LogDeprecated:     options.LogDeprecated,
		ResolvedDomain:    options.ResolvedDomain,
		ResolvedSubdomain: options.ResolvedSubomain,
		DisableHTTPS:      options.DisableHTTPS,
	}
}

// Resolver CodeDeploy endpoint resolver
type Resolver struct {
	partitions endpoints.Partitions
}

// ResolveEndpoint resolves the service endpoint for the given region and options
func (r *Resolver) ResolveEndpoint(subdomain, domain string, options Options) (endpoint cybr.Endpoint, err error) {
	if len(subdomain) == 0 {
		return endpoint, &cybr.MissingSubdomainError{}
	}

	if len(domain) == 0 {
		return endpoint, &cybr.MissingDomainError{}
	}

	opt := transformToSharedOptions(options)
	return r.partitions.ResolveEndpoint(domain, opt)
}

// New returns a new Resolver
func New() *Resolver {
	return &Resolver{
		partitions: defaultPartitions,
	}
}

var partitionRegexp = struct {
	Cybr *regexp.Regexp
}{

	Cybr: regexp.MustCompile("^(cyberark.cloud)\\d+$"),
}

var defaultPartitions = endpoints.Partitions{
	{
		ID: "cybr",
		Defaults: map[endpoints.DefaultKey]endpoints.Endpoint{
			{
				Variant: 0,
			}: {
				Hostname:  "{domain}",
				Protocols: []string{"https"},
			},
		},
		DomainRegex: partitionRegexp.Cybr,
		Endpoints: endpoints.Endpoints{
			endpoints.EndpointKey{
				Domain:    "cyberark.cloud",
				Subdomain: "",
			}: endpoints.Endpoint{},
		},
	},
}
