package generic

import (
	"bytes"
	"context"
	"fmt"
	"path"

	"github.com/strick-j/cybr-sdk-alpha/cybr/protocol/query"
	smithy "github.com/strick-j/smithy-go"
	"github.com/strick-j/smithy-go/encoding/httpbinding"
	"github.com/strick-j/smithy-go/middleware"
	smithyhttp "github.com/strick-j/smithy-go/transport/http"
)

type cybrQuery_serializeOpGetPlatformToken struct {
}

func (*cybrQuery_serializeOpGetPlatformToken) ID() string {
	return "OperationSerializer"
}

func (m *cybrQuery_serializeOpGetPlatformToken) HandleSerialize(ctx context.Context, in middleware.SerializeInput, next middleware.SerializeHandler) (
	out middleware.SerializeOutput, metadata middleware.Metadata, err error,
) {
	request, ok := in.Request.(*smithyhttp.Request)
	if !ok {
		return out, metadata, &smithy.SerializationError{Err: fmt.Errorf("unknown transport type %T", in.Request)}
	}

	input, ok := in.Parameters.(*GetPlatformTokenInput)
	_ = input
	if !ok {
		return out, metadata, &smithy.SerializationError{Err: fmt.Errorf("unknown input parameters type %T", in.Parameters)}
	}

	operationPath := "/oauth2/platformtoken"
	if len(request.Request.URL.Path) == 0 {
		request.Request.URL.Path = operationPath
	} else {
		request.Request.URL.Path = path.Join(request.Request.URL.Path, operationPath)
		if request.Request.URL.Path != "/" && operationPath[len(operationPath)-1] == '/' {
			request.Request.URL.Path += "/"
		}
	}
	request.Request.Method = "POST"
	httpBindingEncoder, err := httpbinding.NewEncoder(request.URL.Path, request.URL.RawQuery, request.Header)
	if err != nil {
		return out, metadata, &smithy.SerializationError{Err: err}
	}
	httpBindingEncoder.SetHeader("Content-Type").String("application/x-www-form-urlencoded")
	httpBindingEncoder.SetHeader("Accept").String("application/json")
	if len(input.ClientId) == 0 || len(input.ClientSecret) == 0 {
		return out, metadata, &smithy.SerializationError{Err: fmt.Errorf("missing required parameter ClientId or ClientSecret for operation PlatformTokenAuth")}
	}
	request.SetBasicAuth(input.ClientId, input.ClientSecret)

	bodyWriter := bytes.NewBuffer(nil)
	bodyEncoder := query.NewEncoder(bodyWriter)
	body := bodyEncoder.Object()
	if len(input.GrantType) == 0 {
		body.Key("grant_type").String("client_credentials")
	} else {
		body.Key("grant_type").String(input.GrantType)
	}

	err = bodyEncoder.Encode()
	if err != nil {
		return out, metadata, &smithy.SerializationError{Err: err}
	}

	if request, err = request.SetStream(bytes.NewReader(bodyWriter.Bytes())); err != nil {
		return out, metadata, &smithy.SerializationError{Err: err}
	}

	if request.Request, err = httpBindingEncoder.Encode(request.Request); err != nil {
		return out, metadata, &smithy.SerializationError{Err: err}
	}
	in.Request = request

	return next.HandleSerialize(ctx, in)
}
