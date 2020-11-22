package support

import (
	"errors"
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
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
	req := ctx.Request()
	switch strings.ToUpper(scope) {
	case flux.ScopePath:
		return flux.WrapTextMIMEValue(req.PathValue(key)), nil
	case flux.ScopeQuery:
		return flux.WrapTextMIMEValue(req.QueryValue(key)), nil
	case flux.ScopeForm:
		return flux.WrapTextMIMEValue(req.FormValue(key)), nil
	case flux.ScopeHeader:
		return flux.WrapTextMIMEValue(req.HeaderValue(key)), nil
	case flux.ScopeAttr:
		v, _ := ctx.GetAttribute(key)
		return flux.WrapObjectMIMEValue(v), nil
	case flux.ScopeAttrs:
		return flux.WrapStrMapMIMEValue(ctx.Attributes()), nil
	case flux.ScopeBody:
		reader, err := req.RequestBodyReader()
		return flux.MIMEValue{Value: reader, MIMEType: req.HeaderValue("Content-Type")}, err
	case flux.ScopeParam:
		v, _ := SearchValueProviders(key, req.QueryValues, req.FormValues)
		return flux.WrapTextMIMEValue(v), nil
	case flux.ScopeAuto:
		fallthrough
	default:
		if v, ok := SearchValueProviders(key,
			req.PathValues, req.QueryValues, req.FormValues, makeHeaderProvider(req)); ok {
			return flux.WrapTextMIMEValue(v), nil
		} else if v, _ := ctx.GetAttribute(key); "" != v {
			return flux.WrapObjectMIMEValue(v), nil
		} else {
			return flux.WrapObjectMIMEValue(nil), nil
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
