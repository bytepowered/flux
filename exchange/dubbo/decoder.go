package dubbo

import (
	hessian "github.com/apache/dubbo-go-hessian2"
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

func NewDubboExchangeDecoderWithKeys(codeKey, headerKey, bodyKey string) flux.ExchangeDecoder {
	return func(ctx flux.Context, input interface{}) (statusCode int, header http.Header, body flux.Object, err error) {
		header = make(http.Header)
		if mapValues, ok := input.(map[interface{}]interface{}); ok {
			// Header
			header, err = ReadHeaderObject(headerKey, mapValues)
			if nil != err {
				return flux.StatusServerError, header, nil, err
			}
			// StatusCode
			statusCode, err = ReadStatusCode(codeKey, mapValues)
			if nil != err {
				return flux.StatusServerError, header, nil, err
			}
			// Body
			body = ReadBodyObject(bodyKey, mapValues)
			return statusCode, header, body, nil
		} else {
			return flux.StatusOK, header, input, nil
		}
	}
}

func NewDubboExchangeDecoder() flux.ExchangeDecoder {
	return NewDubboExchangeDecoderWithKeys(ResponseKeyStatusCode, ResponseKeyHeaders, ResponseKeyBody)
}

func ReadBodyObject(key string, values map[interface{}]interface{}) hessian.Object {
	if body, ok := values[key]; ok {
		return body.(hessian.Object)
	} else {
		return values
	}
}

func ReadStatusCode(key string, values map[interface{}]interface{}) (int, error) {
	if status, ok := values[key]; ok {
		if code, err := cast.ToIntE(status); nil != err {
			logger.Warnf("Invalid rpc response status, type: %s, value: %+v", reflect.TypeOf(status), status)
			return 0, ErrInvalidStatus
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
	logger.Warnf("Invalid rpc response headers, type: %s, value: %+v", reflect.TypeOf(hkv), hkv)
	return nil, ErrInvalidHeaders
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
