package http

import (
	"errors"
	"github.com/bytepowered/flux"
	"net/http"
)

var (
	ErrUnknownHttpBackendResponse = errors.New("BACKEND:UNKNOWN_HTTP_RESPONSE")
)

func NewBackendResultDecodeFunc() flux.BackendResultDecodeFunc {
	return func(ctx flux.Context, value interface{}) (*flux.BackendResult, error) {
		resp, ok := value.(*http.Response)
		if !ok {
			return &flux.BackendResult{
				StatusCode: http.StatusBadGateway,
				Headers:    make(http.Header, 0),
				Body:       nil,
			}, ErrUnknownHttpBackendResponse
		}
		return &flux.BackendResult{
			StatusCode: resp.StatusCode,
			Headers:    resp.Header,
			Body:       resp.Body,
		}, nil
	}
}
