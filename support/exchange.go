package support

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
)

var (
	ErrExchangeDecoderNotFound = &flux.StateError{
		StatusCode: flux.StatusServerError,
		ErrorCode:  flux.ErrorCodeGatewayInternal,
		Message:    "EXCHANGE:DECODER_NOT_FOUND",
	}
)

func InvokeExchanger(ctx flux.Context, exchange flux.Exchange) *flux.StateError {
	endpoint := ctx.Endpoint()
	resp, err := exchange.Invoke(&endpoint, ctx)
	if err != nil {
		return err
	}
	// decode responseWriter
	decoder, ok := ext.GetExchangeDecoder(endpoint.Protocol)
	if !ok {
		return ErrExchangeDecoderNotFound
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
			Message:    "EXCHANGE:DECODE_RESPONSE",
			Internal:   err,
		}
	}
}
