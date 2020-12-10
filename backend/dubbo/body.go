package dubbo

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/logger"
	"github.com/spf13/cast"
	"net/http"
	"reflect"
)

type BodyValues map[interface{}]interface{}

func WrapBodyValues(v interface{}) (BodyValues, bool) {
	if m, ok := v.(map[interface{}]interface{}); ok {
		return BodyValues(m), true
	} else {
		return nil, false
	}
}

func (b BodyValues) ReadBodyValue(bodyKey string) interface{} {
	if body, ok := b[bodyKey]; ok {
		return body
	} else {
		return b
	}
}

func (b BodyValues) ReadStatusValue(statusKey string) (int, error) {
	if status, ok := b[statusKey]; ok {
		if code, err := cast.ToIntE(status); nil != err {
			logger.Warnw("Invalid rpc response status",
				"type", reflect.TypeOf(status), "status", status)
			return 0, ErrDubboDecodeInvalidStatus
		} else {
			return code, nil
		}
	} else {
		return flux.StatusOK, nil
	}
}

func (b BodyValues) ReadHeaderValue(headerKey string) (http.Header, error) {
	hkv, ok := b[headerKey]
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
	return nil, ErrDubboDecodeInvalidHeaders
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
