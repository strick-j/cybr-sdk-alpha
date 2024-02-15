package middleware

import (
	"context"

	"github.com/strick-j/cybr-sdk-alpha/cybr"

	"github.com/strick-j/smithy-go/middleware"
)

// RegisterServiceMetadata registers metadata about the service and operation into the middleware context
// so that it is available at runtime for other middleware to introspect.
type RegisterServiceMetadata struct {
	ServiceID     string
	Domain        string
	Subdomain     string
	OperationName string
}

// ID returns the middleware identifier.
func (s *RegisterServiceMetadata) ID() string {
	return "RegisterServiceMetadata"
}

// HandleInitialize registers service metadata information into the middleware context, allowing for introspection.
func (s RegisterServiceMetadata) HandleInitialize(
	ctx context.Context, in middleware.InitializeInput, next middleware.InitializeHandler,
) (out middleware.InitializeOutput, metadata middleware.Metadata, err error) {
	if len(s.ServiceID) > 0 {
		ctx = SetServiceID(ctx, s.ServiceID)
	}
	if len(s.Domain) > 0 {
		ctx = setDomain(ctx, s.Domain)
	}
	if len(s.Subdomain) > 0 {
		ctx = setSubdomain(ctx, s.Subdomain)
	}
	if len(s.OperationName) > 0 {
		ctx = setOperationName(ctx, s.OperationName)
	}
	return next.HandleInitialize(ctx, in)
}

// service metadata keys for storing and lookup of runtime stack information.
type (
	serviceIDKey     struct{}
	domainKey        struct{}
	subdomainKey     struct{}
	operationNameKey struct{}
	partitionIDKey   struct{}
)

// GetServiceID retrieves the service id from the context.
//
// Scoped to stack values. Use github.com/aws/smithy-go/middleware#ClearStackValues
// to clear all stack values.
func GetServiceID(ctx context.Context) (v string) {
	v, _ = middleware.GetStackValue(ctx, serviceIDKey{}).(string)
	return v
}

// GetDomain retrieves the endpoint domain from the context.
//
// Scoped to stack values. Use github.com/aws/smithy-go/middleware#ClearStackValues
// to clear all stack values.
func GetDomain(ctx context.Context) (v string) {
	v, _ = middleware.GetStackValue(ctx, domainKey{}).(string)
	return v
}

// GetSubdomain retrieves the endpoint subdomain from the context.
//
// Scoped to stack values. Use github.com/aws/smithy-go/middleware#ClearStackValues
// to clear all stack values.
func GetSubdomain(ctx context.Context) (v string) {
	v, _ = middleware.GetStackValue(ctx, subdomainKey{}).(string)
	return v
}

// GetOperationName retrieves the service operation metadata from the context.
//
// Scoped to stack values. Use github.com/aws/smithy-go/middleware#ClearStackValues
// to clear all stack values.
func GetOperationName(ctx context.Context) (v string) {
	v, _ = middleware.GetStackValue(ctx, operationNameKey{}).(string)
	return v
}

// GetPartitionID retrieves the endpoint partition id from the context.
//
// Scoped to stack values. Use github.com/aws/smithy-go/middleware#ClearStackValues
// to clear all stack values.
func GetPartitionID(ctx context.Context) string {
	v, _ := middleware.GetStackValue(ctx, partitionIDKey{}).(string)
	return v
}

// SetServiceID sets the service id on the context.
//
// Scoped to stack values. Use github.com/aws/smithy-go/middleware#ClearStackValues
// to clear all stack values.
func SetServiceID(ctx context.Context, value string) context.Context {
	return middleware.WithStackValue(ctx, serviceIDKey{}, value)
}

// setDomain sets the endpoint domain on the context.
//
// Scoped to stack values. Use github.com/aws/smithy-go/middleware#ClearStackValues
// to clear all stack values.
func setDomain(ctx context.Context, value string) context.Context {
	return middleware.WithStackValue(ctx, domainKey{}, value)
}

// setSubdomain sets the endpoint subdomain on the context.
//
// Scoped to stack values. Use github.com/aws/smithy-go/middleware#ClearStackValues
// to clear all stack values.
func setSubdomain(ctx context.Context, value string) context.Context {
	return middleware.WithStackValue(ctx, subdomainKey{}, value)
}

// setOperationName sets the service operation on the context.
//
// Scoped to stack values. Use github.com/aws/smithy-go/middleware#ClearStackValues
// to clear all stack values.
func setOperationName(ctx context.Context, value string) context.Context {
	return middleware.WithStackValue(ctx, operationNameKey{}, value)
}

// SetPartitionID sets the partition id of a resolved region on the context
//
// Scoped to stack values. Use github.com/aws/smithy-go/middleware#ClearStackValues
// to clear all stack values.
func SetPartitionID(ctx context.Context, value string) context.Context {
	return middleware.WithStackValue(ctx, partitionIDKey{}, value)
}

type signingCredentialsKey struct{}

// GetSigningCredentials returns the credentials that were used for signing if set on context.
func GetSigningCredentials(ctx context.Context) (v cybr.Credentials) {
	v, _ = middleware.GetStackValue(ctx, signingCredentialsKey{}).(cybr.Credentials)
	return v
}

// SetSigningCredentials sets the credentails used for signing on the context.
func SetSigningCredentials(ctx context.Context, value cybr.Credentials) context.Context {
	return middleware.WithStackValue(ctx, signingCredentialsKey{}, value)
}

// EndpointSource key
type endpointSourceKey struct{}

// GetEndpointSource returns an endpoint source if set on context
func GetEndpointSource(ctx context.Context) (v cybr.EndpointSource) {
	v, _ = middleware.GetStackValue(ctx, endpointSourceKey{}).(cybr.EndpointSource)
	return v
}

// SetEndpointSource sets endpoint source on context
func SetEndpointSource(ctx context.Context, value cybr.EndpointSource) context.Context {
	return middleware.WithStackValue(ctx, endpointSourceKey{}, value)
}
