package http

import (
	"context"

	cybrmiddleware "github.com/strick-j/cybr-sdk-alpha/cybr/middleware"
	"github.com/strick-j/smithy-go/middleware"
	smithyhttp "github.com/strick-j/smithy-go/transport/http"
)

// AddResponseErrorMiddleware adds response error wrapper middleware
func AddResponseErrorMiddleware(stack *middleware.Stack) error {
	// add error wrapper middleware before request id retriever middleware so that it can wrap the error response
	// returned by operation deserializers
	return stack.Deserialize.Insert(&responseErrorWrapper{}, "RequestIDRetriever", middleware.Before)
}

type responseErrorWrapper struct {
}

// ID returns the middleware identifier
func (m *responseErrorWrapper) ID() string {
	return "ResponseErrorWrapper"
}

func (m *responseErrorWrapper) HandleDeserialize(ctx context.Context, in middleware.DeserializeInput, next middleware.DeserializeHandler) (
	out middleware.DeserializeOutput, metadata middleware.Metadata, err error,
) {
	out, metadata, err = next.HandleDeserialize(ctx, in)
	if err == nil {
		// Nothing to do when there is no error.
		return out, metadata, err
	}

	resp, ok := out.RawResponse.(*smithyhttp.Response)
	if !ok {
		// No raw response to wrap with.
		return out, metadata, err
	}

	// look for request id in metadata
	reqID, _ := cybrmiddleware.GetRequestIDMetadata(metadata)

	// Wrap the returned smithy error with the request id retrieved from the metadata
	err = &ResponseError{
		ResponseError: &smithyhttp.ResponseError{
			Response: resp,
			Err:      err,
		},
		RequestID: reqID,
	}

	return out, metadata, err
}
