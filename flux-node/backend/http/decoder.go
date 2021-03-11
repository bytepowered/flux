package http

import (
	"errors"
	flux2 "github.com/bytepowered/flux/flux-node"
	"net/http"
)

var (
	ErrUnknownHttpBackendResponse = errors.New("BACKEND:UNKNOWN_HTTP_RESPONSE")
)

func NewBackendResponseCodecFunc() flux2.BackendResponseCodecFunc {
	return func(ctx flux2.Context, value interface{}) (*flux2.BackendResponse, error) {
		resp, ok := value.(*http.Response)
		if !ok {
			return &flux2.BackendResponse{
				StatusCode: http.StatusBadGateway,
				Headers:    make(http.Header, 0),
				Body:       nil,
			}, ErrUnknownHttpBackendResponse
		}
		return &flux2.BackendResponse{
			StatusCode: resp.StatusCode,
			Headers:    resp.Header,
			Body:       resp.Body,
		}, nil
	}
}
