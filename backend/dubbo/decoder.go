package dubbo

import (
	"github.com/bytepowered/flux"
	"net/http"
)

const (
	ResponseKeyStatusCode = "@net.bytepowered.flux.http-status"
	ResponseKeyHeaders    = "@net.bytepowered.flux.http-headers"
	ResponseKeyBody       = "@net.bytepowered.flux.http-body"
)

func NewDubboBackendTransportDecodeFuncWith(codeKey, headerKey, bodyKey string) flux.BackendTransportDecodeFunc {
	return func(ctx flux.Context, input interface{}) (int, http.Header, interface{}, error) {
		bodyValues, ok := WrapBodyValues(input)
		if !ok {
			return flux.StatusOK, make(http.Header, 0), input, nil
		}
		// Header
		header, err := bodyValues.ReadHeaderValue(headerKey)
		if nil != err {
			return flux.StatusServerError, make(http.Header, 0), nil, err
		}
		// StatusCode
		status, err := bodyValues.ReadStatusValue(codeKey)
		if nil != err {
			return flux.StatusServerError, make(http.Header, 0), nil, err
		}
		// Body
		body := bodyValues.ReadBodyValue(bodyKey)
		return status, header, body, nil
	}
}

func NewDubboBackendTransportDecodeFunc() flux.BackendTransportDecodeFunc {
	return NewDubboBackendTransportDecodeFuncWith(ResponseKeyStatusCode, ResponseKeyHeaders, ResponseKeyBody)
}
