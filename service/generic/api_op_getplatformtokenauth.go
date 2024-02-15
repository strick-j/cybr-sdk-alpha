package generic

import (
	"context"
	"fmt"

	cybrmiddleware "github.com/strick-j/cybr-sdk-alpha/cybr/middleware"
	"github.com/strick-j/smithy-go/middleware"
	smithyhttp "github.com/strick-j/smithy-go/transport/http"
)

// Authenticates to the platform token endpoint
func (c *Client) GetPlatformToken(ctx context.Context, params *GetPlatformTokenInput, optFns ...func(*Options)) (*GetPlatformTokenOutput, error) {
	if params == nil {
		params = &GetPlatformTokenInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "PlatformTokenAuth", params, optFns, c.addOperationGetPlatformTokenMiddleware)
	if err != nil {
		return nil, err
	}

	out := result.(*GetPlatformTokenOutput)
	out.ResultMetadata = metadata

	return out, nil
}

type GetPlatformTokenInput struct {

	// The grant type to use for the token request.
	// Typically this should be client_credentials.
	GrantType string

	// The client ID to use for the token request.
	ClientId string

	// The client secret to use for the token request.
	ClientSecret string
}

type GetPlatformTokenOutput struct {
	ResultMetadata middleware.Metadata
}

func (c *Client) addOperationGetPlatformTokenMiddleware(stack *middleware.Stack, options Options) (err error) {
	if err := stack.Serialize.Add(&setOperationInputMiddleware{}, middleware.After); err != nil {
		return err
	}
	err = stack.Serialize.Add(&cybrQuery_serializeOpGetPlatformToken{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&cybrQuery_deserializeOpGetPlatformToken{}, middleware.After)
	if err != nil {
		return err
	}
	if err := addProtocolFinalizerMiddlewares(stack, options, "GetPlatformToken"); err != nil {
		return fmt.Errorf("add protocol finalizers: %v", err)
	}

	if err = addSetLoggerMiddleware(stack, options); err != nil {
		return err
	}
	if err = cybrmiddleware.AddClientRequestIDMiddleware(stack); err != nil {
		return err
	}
	if err = smithyhttp.AddComputeContentLengthMiddleware(stack); err != nil {
		return err
	}
	if err = addResolveEndpointMiddleware(stack, options); err != nil {
		return err
	}
	if err = cybrmiddleware.AddRawResponseToMetadata(stack); err != nil {
		return err
	}
	if err = cybrmiddleware.AddRecordResponseTiming(stack); err != nil {
		return err
	}
	if err = smithyhttp.AddErrorCloseResponseBodyMiddleware(stack); err != nil {
		return err
	}
	if err = smithyhttp.AddCloseResponseBodyMiddleware(stack); err != nil {
		return err
	}
	//if err = addOpAddUserToGroupValidationMiddleware(stack); err != nil {
	//	return err
	//}
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opGetPlatformToken(options.Subdomain, options.Domain), middleware.Before); err != nil {
		return err
	}
	if err = addRequestIDRetrieverMiddleware(stack); err != nil {
		return err
	}
	if err = addResponseErrorMiddleware(stack); err != nil {
		return err
	}
	if err = addRequestResponseLogging(stack, options); err != nil {
		return err
	}

	return nil
}

func newServiceMetadataMiddleware_opGetPlatformToken(subdomain string, domain string) *cybrmiddleware.RegisterServiceMetadata {
	return &cybrmiddleware.RegisterServiceMetadata{
		Subdomain:     subdomain,
		Domain:        domain,
		ServiceID:     ServiceID,
		OperationName: "AddUserToGroup",
	}
}
