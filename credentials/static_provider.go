package credentials

import (
	"context"

	"github.com/strick-j/cybr-sdk-alpha/cybr"
)

const (
	// StaticCredentialsName provides a name of Static provider
	StaticCredentialsName = "StaticCredentials"
)

// StaticCredentialsEmptyError is emitted when static credentials are empty.
type StaticCredentialsEmptyError struct{}

func (*StaticCredentialsEmptyError) Error() string {
	return "static credentials are empty"
}

// A StaticCredentialsProvider is a set of credentials which are set, and will
// never expire.
type StaticCredentialsProvider struct {
	Value cybr.Credentials
}

// NewStaticCredentialsProvider return a StaticCredentialsProvider initialized with the CyberArk
// credentials passed in.
func NewStaticCredentialsProvider(key, secret, session string) StaticCredentialsProvider {
	return StaticCredentialsProvider{
		Value: cybr.Credentials{
			Username:     key,
			Password:     secret,
			SessionToken: session,
		},
	}
}

// Retrieve returns the credentials or error if the credentials are invalid.
func (s StaticCredentialsProvider) Retrieve(_ context.Context) (cybr.Credentials, error) {
	v := s.Value
	if v.Username == "" || v.Password == "" {
		return cybr.Credentials{
			Source: StaticCredentialsName,
		}, &StaticCredentialsEmptyError{}
	}

	if len(v.Source) == 0 {
		v.Source = StaticCredentialsName
	}

	return v, nil
}
