package internal

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
)

var (
	ErrExchangeDecoderNotFound = &flux.InvokeError{
		StatusCode: flux.StatusServerError,
		ErrorCode:  flux.ErrorCodeGatewayInternal,
		Message:    "EXCHANGE:DECODER_NOT_FOUND",
	}
)

func InvokeExchanger(ctx flux.Context, exchange flux.Exchange) *flux.InvokeError {
	endpoint := ctx.Endpoint()
	resp, err := exchange.Invoke(&endpoint, ctx)
	if err != nil {
		return err
	}
	// decode response
	decoder, ok := ext.GetExchangeDecoder(endpoint.Protocol)
	if !ok {
		return ErrExchangeDecoderNotFound
	}
	if code, headers, body, err := decoder(ctx, resp); nil == err {
		ctx.ResponseWriter().SetStatusCode(code).SetHeaders(headers).SetBody(body)
		return nil
	} else {
		return &flux.InvokeError{
			StatusCode: code,
			Message:    err.Error(),
			Internal:   err,
		}
	}
}
