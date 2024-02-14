package generic

import (
	"context"

	"github.com/strick-j/cybr-sdk-alpha/cybr"
	cybrmiddleware "github.com/strick-j/cybr-sdk-alpha/cybr/middleware"
	cybrhttp "github.com/strick-j/cybr-sdk-alpha/cybr/transport/http"
	smithy "github.com/strick-j/smithy-go"
	"github.com/strick-j/smithy-go/logging"
	"github.com/strick-j/smithy-go/middleware"
	smithyhttp "github.com/strick-j/smithy-go/transport/http"
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
	options = options.Copy()

	resolveDefaultLogger(&options)

	resolveHTTPClient(&options)

	for _, fn := range optFns {
		fn(&options)
	}

	client := &Client{
		options: options,
	}

	return client
}

func (c *Client) Options() Options {
	return c.options.Copy()
}

func (c *Client) invokeOperation(ctx context.Context, opID string, params interface{}, optFns []func(*Options), stackFns ...func(*middleware.Stack, Options) error) (result interface{}, metadata middleware.Metadata, err error) {
	ctx = middleware.ClearStackValues(ctx)
	stack := middleware.NewStack(opID, smithyhttp.NewStackRequest)
	options := c.options.Copy()

	for _, fn := range optFns {
		fn(&options)
	}

	for _, fn := range stackFns {
		if err := fn(stack, options); err != nil {
			return nil, metadata, err
		}
	}

	for _, fn := range options.APIOptions {
		if err := fn(stack); err != nil {
			return nil, metadata, err
		}
	}

	handler := middleware.DecorateHandler(smithyhttp.NewClientHandler(options.HTTPClient), stack)
	result, metadata, err = handler.Handle(ctx, params)
	if err != nil {
		err = &smithy.OperationError{
			ServiceID:     ServiceID,
			OperationName: opID,
			Err:           err,
		}
	}
	return result, metadata, err
}

type operationInputKey struct{}

func setOperationInput(ctx context.Context, input interface{}) context.Context {
	return middleware.WithStackValue(ctx, operationInputKey{}, input)
}

func getOperationInput(ctx context.Context) interface{} {
	return middleware.GetStackValue(ctx, operationInputKey{})
}

type setOperationInputMiddleware struct {
}

func (*setOperationInputMiddleware) ID() string {
	return "setOperationInput"
}

func (m *setOperationInputMiddleware) HandleSerialize(ctx context.Context, in middleware.SerializeInput, next middleware.SerializeHandler) (
	out middleware.SerializeOutput, metadata middleware.Metadata, err error,
) {
	ctx = setOperationInput(ctx, in.Parameters)
	return next.HandleSerialize(ctx, in)
}

func resolveDefaultLogger(o *Options) {
	if o.Logger != nil {
		return
	}
	o.Logger = logging.Nop{}
}

func addSetLoggerMiddleware(stack *middleware.Stack, o Options) error {
	return middleware.AddSetLoggerMiddleware(stack, o.Logger)
}

func resolveHTTPClient(o *Options) {
	var service *cybrhttp.HTTPTransportBuilder

	if o.HTTPClient != nil {
		var ok bool
		service, ok = o.HTTPClient.(*cybrhttp.HTTPTransportBuilder)
		if !ok {
			return
		}
	} else {
		service = cybrhttp.NewHTTPTransportBuilder()
	}

	o.HTTPClient = service
}

func NewFromConfig(cfg cybr.Config, optFns ...func(*Options)) *Client {
	opts := Options{
		Domain:     cfg.Domain,
		Subdomain:  cfg.SubDomain,
		HTTPClient: cfg.HTTPClient,
		Logger:     cfg.Logger,
	}
	return New(opts, optFns...)
}

func addRequestIDRetrieverMiddleware(stack *middleware.Stack) error {
	return cybrmiddleware.AddRequestIDRetrieverMiddleware(stack)
}

func addResponseErrorMiddleware(stack *middleware.Stack) error {
	return cybrhttp.AddResponseErrorMiddleware(stack)
}

func addRequestResponseLogging(stack *middleware.Stack, o Options) error {
	return stack.Deserialize.Add(&smithyhttp.RequestResponseLogger{
		LogRequest:          o.ClientLogMode.IsRequest(),
		LogRequestWithBody:  o.ClientLogMode.IsRequestWithBody(),
		LogResponse:         o.ClientLogMode.IsResponse(),
		LogResponseWithBody: o.ClientLogMode.IsResponseWithBody(),
	}, middleware.After)
}
