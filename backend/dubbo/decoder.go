package dubbo

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/logger"
	"github.com/spf13/cast"
	"net/http"
	"reflect"
)

const (
	ResponseKeyStatusCode = "@net.bytepowered.flux.http-status"
	ResponseKeyHeaders    = "@net.bytepowered.flux.http-headers"
	ResponseKeyBody       = "@net.bytepowered.flux.http-body"
)

func NewDubboBackendTransportDecodeFuncWith(codeKey, headerKey, bodyKey string) flux.BackendTransportDecodeFunc {
	return func(ctx flux.Context, input interface{}) (int, http.Header, interface{}, error) {
		if mapValues, ok := input.(map[interface{}]interface{}); ok {
			// Header
			header, err := ReadHeaderObject(headerKey, mapValues)
			if nil != err {
				return flux.StatusServerError, make(http.Header, 0), nil, err
			}
			// StatusCode
			status, err := ReadStatusCode(codeKey, mapValues)
			if nil != err {
				return flux.StatusServerError, make(http.Header, 0), nil, err
			}
			// Body
			body := ReadBodyObject(bodyKey, mapValues)
			return status, header, body, nil
		} else {
			return flux.StatusOK, make(http.Header, 0), input, nil
		}
	}
}

func NewDubboBackendTransportDecodeFunc() flux.BackendTransportDecodeFunc {
	return NewDubboBackendTransportDecodeFuncWith(ResponseKeyStatusCode, ResponseKeyHeaders, ResponseKeyBody)
}

func ReadBodyObject(key string, values map[interface{}]interface{}) interface{} {
	if body, ok := values[key]; ok {
		return body
	} else {
		return values
	}
}

func ReadStatusCode(key string, values map[interface{}]interface{}) (int, error) {
	if status, ok := values[key]; ok {
		if code, err := cast.ToIntE(status); nil != err {
			logger.Warnw("Invalid rpc response status", "type", reflect.TypeOf(status), "status", status)
			return 0, ErrDecodeInvalidStatus
		} else {
			return code, nil
		}
	} else {
		return flux.StatusOK, nil
	}
}

func ReadHeaderObject(key string, values map[interface{}]interface{}) (http.Header, error) {
	hkv, ok := values[key]
	if !ok {
		return make(http.Header), nil
	}
	if mss, ok := hkv.(map[string][]string); ok {
		return mss, nil
	}
	if mii, ok := hkv.(map[interface{}]interface{}); ok {
		omap := make(http.Header)
		for k, v := range mii {
			_addToHeader(omap, cast.ToString(k), v)
		}
		return omap, nil
	}
	if msi, ok := hkv.(map[string]interface{}); ok {
		omap := make(http.Header)
		for k, v := range msi {
			_addToHeader(omap, cast.ToString(k), v)
		}
		return omap, nil
	}
	logger.Warnw("Invalid rpc response headers", "type", reflect.TypeOf(hkv), "value", hkv)
	return nil, ErrDecodeInvalidHeaders
}

func _addToHeader(headers http.Header, key string, v interface{}) {
	if sa, ok := v.([]string); ok {
		for _, iv := range sa {
			headers.Add(key, iv)
		}
	} else {
		headers.Add(key, cast.ToString(v))
	}
}
