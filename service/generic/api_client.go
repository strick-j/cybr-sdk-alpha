package generic

import (
	"github.com/strick-j/cybr-sdk-alpha/cybr/logging"
)

const ServiceID = "Generic"

// Client provides the API client to make operations call for "generic" service.
type Client struct {
	options Options
}

// New returns an initialized Client based on the functional options. Provide
// additional functional options to further configure the behavior of the client,
// such as changing the client's endpoint or adding custom middleware behavior.
func New(options Options, optFns ...func(*Options)) *Client {
	//options = options.Copy()

	resolveDefaultLogger(&options)

	for _, fn := range optFns {
		fn(&options)
	}

	client := &Client{
		options: options,
	}

	return client
}

func resolveDefaultLogger(o *Options) {
	if o.Logger != nil {
		return
	}
	o.Logger = logging.Nop{}
}
