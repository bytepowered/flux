package support

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
)

var (
	ErrBackendResponseDecoderNotFound = &flux.StateError{
		StatusCode: flux.StatusServerError,
		ErrorCode:  flux.ErrorCodeGatewayInternal,
		Message:    "BACKEND:RESPONSE_DECODER:NOT_FOUND",
	}
)

func InvokeBackendExchange(ctx flux.Context, exchange flux.Backend) *flux.StateError {
	endpoint := ctx.Endpoint()
	resp, err := exchange.Invoke(&endpoint, ctx)
	if err != nil {
		return err
	}
	// decode responseWriter
	decoder, ok := ext.GetBackendResponseDecoder(endpoint.Service.Protocol)
	if !ok {
		return ErrBackendResponseDecoderNotFound
	}
	if code, headers, body, err := decoder(ctx, resp); nil == err {
		ctx.Response().SetStatusCode(code)
		ctx.Response().SetHeaders(headers)
		ctx.Response().SetBody(body)
		return nil
	} else {
		return &flux.StateError{
			StatusCode: flux.StatusServerError,
			ErrorCode:  flux.ErrorCodeGatewayInternal,
			Message:    "BACKEND:DECODE_RESPONSE",
			Internal:   err,
		}
	}
}
