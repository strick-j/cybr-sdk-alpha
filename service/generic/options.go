package generic

import (
	"net/http"

	logging "github.com/strick-j/cybr-sdk-alpha/cybr/logging"
)

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

type Options struct {
	// The logger writer interface to write logging messages to.
	Logger logging.Logger

	// The HTTP client to invoke API calls with. Defaults to client's default HTTP
	// implementation if nil.
	HTTPClient HTTPClient
}
