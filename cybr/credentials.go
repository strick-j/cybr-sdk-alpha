package cybr

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/strick-j/cybr-sdk-alpha/internal/sdk"
	//"github.com/strick-j/cybr-sdk-alpha/credentials"
)

type AnonymousCredentials struct{}

// Retrieve implements the CredentialsProvider interface, but will always
// return error, and cannot be used to sign a request. The AnonymousCredentials
// type is used as a sentinel type instructing the CYBR request signing
// middleware to not sign a request.
func (AnonymousCredentials) Retrieve(context.Context) (Credentials, error) {
	return Credentials{Source: "AnonymousCredentials"},
		fmt.Errorf("the AnonymousCredentials is not a valid credential provider, and cannot be used to sign CYBR requests with")
}

// A Credentials is the AWS credentials value for individual credential fields.
type Credentials struct {
	// CyberArk Username
	Username string

	// CyberArk Password
	Password string

	// CyberArk Session Token
	SessionToken string

	// Source of the credentials
	Source string

	// States if the credentials can expire or not.
	CanExpire bool

	// The time the credentials will expire at. Should be ignored if CanExpire
	// is false.
	Expires time.Time
}

// Expired returns if the credentials have expired.
func (v Credentials) Expired() bool {
	if v.CanExpire {
		// Calling Round(0) on the current time will truncate the monotonic
		// reading only. Ensures credential expiry time is always based on
		// reported wall-clock time.
		return !v.Expires.After(sdk.NowTime().Round(0))
	}

	return false
}

// HasKeys returns if the credentials keys are set.
func (v Credentials) HasKeys() bool {
	return len(v.Username) > 0 && len(v.Password) > 0
}

// A CredentialsProvider is the interface for any component which will provide
// credentials Credentials. A CredentialsProvider is required to manage its own
// Expired state, and what to be expired means.
//
// A credentials provider implementation can be wrapped with a CredentialCache
// to cache the credential value retrieved. Without the cache the SDK will
// attempt to retrieve the credentials for every request.
type CredentialsProvider interface {
	// Retrieve returns nil if it successfully retrieved the value.
	// Error is returned if the value were not obtainable, or empty.
	Retrieve(ctx context.Context) (Credentials, error)
}

// CredentialsProviderFunc provides a helper wrapping a function value to
// satisfy the CredentialsProvider interface.
type CredentialsProviderFunc func(context.Context) (Credentials, error)

// Retrieve delegates to the function value the CredentialsProviderFunc wraps.
func (fn CredentialsProviderFunc) Retrieve(ctx context.Context) (Credentials, error) {
	return fn(ctx)
}

type isCredentialsProvider interface {
	IsCredentialsProvider(CredentialsProvider) bool
}

// IsCredentialsProvider returns whether the target CredentialProvider is the same type as provider when comparing the
// implementation type.
//
// If provider has a method IsCredentialsProvider(CredentialsProvider) bool it will be responsible for validating
// whether target matches the credential provider type.
//
// When comparing the CredentialProvider implementations provider and target for equality, the following rules are used:
//
//	If provider is of type T and target is of type V, true if type *T is the same as type *V, otherwise false
//	If provider is of type *T and target is of type V, true if type *T is the same as type *V, otherwise false
//	If provider is of type T and target is of type *V, true if type *T is the same as type *V, otherwise false
//	If provider is of type *T and target is of type *V,true if type *T is the same as type *V, otherwise false
func IsCredentialsProvider(provider, target CredentialsProvider) bool {
	if target == nil || provider == nil {
		return provider == target
	}

	if x, ok := provider.(isCredentialsProvider); ok {
		return x.IsCredentialsProvider(target)
	}

	targetType := reflect.TypeOf(target)
	if targetType.Kind() != reflect.Ptr {
		targetType = reflect.PtrTo(targetType)
	}

	providerType := reflect.TypeOf(provider)
	if providerType.Kind() != reflect.Ptr {
		providerType = reflect.PtrTo(providerType)
	}

	return targetType.AssignableTo(providerType)
}
