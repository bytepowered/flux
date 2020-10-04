package http

import (
	"errors"
	"github.com/bytepowered/flux"
	"net/http"
)

var (
	ErrUnknownHttpBackendResponse = errors.New("BACKEND:UNKNOWN_HTTP_RESPONSE")
)

func NewHttpBackendResponseDecoder() flux.BackendResponseDecoder {
	return func(ctx flux.Context, value interface{}) (statusCode int, headers http.Header, body interface{}, err error) {
		resp, ok := value.(*http.Response)
		if !ok {
			return http.StatusInternalServerError, http.Header{}, nil, ErrUnknownHttpBackendResponse
		}
		return resp.StatusCode, resp.Header, resp.Body, nil
	}
}
