package http

import (
	"errors"
	"github.com/bytepowered/flux"
	"net/http"
)

var (
	ErrUnknownHttpResponse = errors.New("TRANSPORTER:HTTP:UNKNOWN_RESPONSE")
)

func NewTransportCodecFunc() flux.TransportCodec {
	return func(ctx *flux.Context, value interface{}) (*flux.ResponseBody, error) {
		resp, ok := value.(*http.Response)
		if !ok {
			return &flux.ResponseBody{
				StatusCode: http.StatusBadGateway,
				Headers:    make(http.Header, 0),
				Body:       nil,
			}, ErrUnknownHttpResponse
		}
		return &flux.ResponseBody{
			StatusCode: resp.StatusCode,
			Headers:    resp.Header,
			Body:       resp.Body,
		}, nil
	}
}
