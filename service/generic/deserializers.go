package generic

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"

	cybrmiddleware "github.com/strick-j/cybr-sdk-alpha/cybr/middleware"
	cybrxml "github.com/strick-j/cybr-sdk-alpha/cybr/protocol/xml"
	smithy "github.com/strick-j/smithy-go"
	"github.com/strick-j/smithy-go/middleware"
	smithyhttp "github.com/strick-j/smithy-go/transport/http"
)

type cybrQuery_deserializeOpGetPlatformToken struct {
}

func (*cybrQuery_deserializeOpGetPlatformToken) ID() string {
	return "OperationDeserializer"
}

func (*cybrQuery_deserializeOpGetPlatformToken) HandleDeserialize(ctx context.Context, in middleware.DeserializeInput, next middleware.DeserializeHandler) (
	out middleware.DeserializeOutput, metadata middleware.Metadata, err error,
) {
	out, metadata, err = next.HandleDeserialize(ctx, in)
	if err != nil {
		return out, metadata, err
	}

	response, ok := out.RawResponse.(*smithyhttp.Response)
	if !ok {
		return out, metadata, &smithy.DeserializationError{Err: fmt.Errorf("unknown transport type %T", out.RawResponse)}
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return out, metadata, cybrQuery_deserializeOpErrorGetPlatformToken(response, &metadata)
	}
	output := &GetPlatformTokenOutput{}
	out.Result = output

	if _, err = io.Copy(ioutil.Discard, response.Body); err != nil {
		return out, metadata, &smithy.DeserializationError{
			Err: fmt.Errorf("failed to discard response body, %w", err),
		}
	}

	return out, metadata, err
}

func cybrQuery_deserializeOpErrorGetPlatformToken(response *smithyhttp.Response, metadata *middleware.Metadata) error {
	var errorBuffer bytes.Buffer
	if _, err := io.Copy(&errorBuffer, response.Body); err != nil {
		return &smithy.DeserializationError{Err: fmt.Errorf("failed to copy error response body, %w", err)}
	}
	errorBody := bytes.NewReader(errorBuffer.Bytes())

	errorCode := "UnknownError"
	errorMessage := errorCode

	errorComponents, err := cybrxml.GetErrorResponseComponents(errorBody, false)
	if err != nil {
		return err
	}
	if reqID := errorComponents.RequestID; len(reqID) != 0 {
		cybrmiddleware.SetRequestIDMetadata(metadata, reqID)
	}
	if len(errorComponents.Code) != 0 {
		errorCode = errorComponents.Code
	}
	if len(errorComponents.Message) != 0 {
		errorMessage = errorComponents.Message
	}
	errorBody.Seek(0, io.SeekStart)
	switch {

	default:
		genericError := &smithy.GenericAPIError{
			Code:    errorCode,
			Message: errorMessage,
		}
		return genericError

	}
}
