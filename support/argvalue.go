package support

import (
	"errors"
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"net/url"
	"strings"
)

// 默认实现：查找Argument的值函数
func DefaultArgumentValueLookupFunc(scope, key string, ctx flux.Context) (value flux.MIMEValue, err error) {
	if "" == scope || "" == key {
		return flux.WrapObjectMIMEValue(nil), errors.New("lookup empty scope or key, scope: " + scope + ", key: " + key)
	}
	if nil == ctx {
		return flux.WrapObjectMIMEValue(nil), errors.New("lookup nil context")
	}
	request := ctx.Request()
	switch strings.ToUpper(scope) {
	case flux.ScopeQuery:
		return flux.WrapTextMIMEValue(request.QueryValue(key)), nil
	case flux.ScopePath:
		return flux.WrapTextMIMEValue(request.PathValue(key)), nil
	case flux.ScopeHeader:
		return flux.WrapTextMIMEValue(request.HeaderValue(key)), nil
	case flux.ScopeForm:
		return flux.WrapTextMIMEValue(request.FormValue(key)), nil
	case flux.ScopeBody:
		reader, err := request.RequestBodyReader()
		return flux.MIMEValue{Value: reader, MIMEType: request.HeaderValue("Content-Type")}, err
	case flux.ScopeAttrs:
		return flux.WrapStrMapMIMEValue(ctx.Attributes()), nil
	case flux.ScopeAttr:
		v, _ := ctx.GetAttribute(key)
		return flux.WrapObjectMIMEValue(v), nil
	case flux.ScopeParam:
		if v := request.QueryValue(key); "" != v {
			return flux.WrapTextMIMEValue(v), nil
		} else {
			return flux.WrapTextMIMEValue(request.FormValue(key)), nil
		}
	default:
		find := func(key string, sources ...url.Values) (string, bool) {
			for _, source := range sources {
				if vs, ok := source[key]; ok {
					return vs[0], true
				}
			}
			return "", false
		}
		if v, ok := find(key, request.PathValues(), request.QueryValues(), request.FormValues()); ok {
			return flux.WrapTextMIMEValue(v), nil
		} else if v := request.HeaderValue(key); "" != v {
			return flux.WrapTextMIMEValue(v), nil
		} else if v, _ := ctx.GetAttribute(key); "" != v {
			return flux.WrapObjectMIMEValue(v), nil
		} else {
			return flux.WrapObjectMIMEValue(value), nil
		}
	}
}

// 默认实现：查找Argument的值解析函数
func DefaultArgumentValueResolveFunc(mtValue flux.MIMEValue, arg flux.Argument, ctx flux.Context) (interface{}, error) {
	valueResolver := ext.GetTypedValueResolver(arg.Class)
	if nil == valueResolver {
		logger.TraceContext(ctx).Warnw("Not supported argument type",
			"http.key", arg.HttpName, "arg.name", arg.Name, "class", arg.Class, "generic", arg.Generic)
		valueResolver = ext.GetDefaultTypedValueResolver()
	}
	if value, err := valueResolver(arg.Class, arg.Generic, mtValue); nil != err {
		logger.TraceContext(ctx).Warnw("Failed to resolve argument",
			"http.key", arg.HttpName, "arg.name", arg.Name, "class", arg.Class, "generic", arg.Generic,
			"http.value", mtValue.Value, "error", err)
		return nil, fmt.Errorf("PARAMETERS:RESOLVE_VALUE:%w", err)
	} else {
		return value, nil
	}
}
