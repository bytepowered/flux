package dubbo

import (
	hessian "github.com/apache/dubbo-go-hessian2"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/pkg"
	"net/http"
	"reflect"
)

func NewDubboExchangeDecoder() flux.ExchangeDecoder {
	return func(ctx flux.Context, value interface{}) (statusCode int, headers http.Header, body flux.Object, err error) {
		emptyHeaders := make(http.Header)
		if valueM, ok := value.(map[interface{}]interface{}); ok {
			headers, err = _headers(valueM)
			if nil != err {
				return 0, emptyHeaders, nil, err
			}
			// StatusCode
			statusCode, err = _status(valueM)
			if nil != err {
				return 0, emptyHeaders, nil, err
			}
			// body
			body = _body(valueM)
			return statusCode, headers, body, nil
		} else {
			return flux.StatusOK, emptyHeaders, value, nil
		}
	}
}

func _body(values map[interface{}]interface{}) hessian.Object {
	if body, ok := values[KeyHttpBody]; ok {
		return body.(hessian.Object)
	} else {
		return values
	}
}

func _status(values map[interface{}]interface{}) (int, error) {
	if status, ok := values[KeyHttpStatus]; ok {
		if code, err := pkg.ToInt(status); nil != err {
			logger.Warnf("Invalid rpc response status, type: %s, value: %+v", reflect.TypeOf(status), status)
			return 0, ErrInvalidStatus
		} else {
			return code, nil
		}
	} else {
		return flux.StatusOK, nil
	}
}

func _headers(values map[interface{}]interface{}) (http.Header, error) {
	hkv, ok := values[KeyHttpHeaders]
	if !ok {
		return make(http.Header), nil
	}
	if mss, ok := hkv.(map[string][]string); ok {
		return mss, nil
	}
	if mii, ok := hkv.(map[interface{}]interface{}); ok {
		omap := make(http.Header)
		for k, v := range mii {
			_addToHeader(omap, pkg.ToString(k), v)
		}
		return omap, nil
	}
	if msi, ok := hkv.(map[string]interface{}); ok {
		omap := make(http.Header)
		for k, v := range msi {
			_addToHeader(omap, pkg.ToString(k), v)
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
		headers.Add(key, pkg.ToString(v))
	}
}
