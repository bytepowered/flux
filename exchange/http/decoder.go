package http

import (
	"github.com/bytepowered/flux"
	"net/http"
)

var (
	ErrUnknownHttpExchangeResponse = &flux.InvokeError{
		StatusCode: http.StatusInternalServerError,
		Message:    "EXCHANGE:UNKNOWN_HTTP_RESPONSE",
	}
)

func NewHttpExchangeDecoder() flux.ExchangeDecoder {
	return func(ctx flux.Context, value interface{}) (statusCode int, headers http.Header, body flux.Object, err error) {
		resp, ok := value.(*http.Response)
		if !ok {
			return http.StatusBadGateway, http.Header{}, nil, ErrUnknownHttpExchangeResponse
		}
		return resp.StatusCode, resp.Header, resp.Body, nil
	}
}
