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
func DefaultArgumentValueLookupFunc(scope, key string, ctx flux.Context) (value flux.MTValue, err error) {
	if "" == scope || "" == key {
		return flux.WrapObjectMTValue(nil), errors.New("lookup empty scope or key, scope: " + scope + ", key: " + key)
	}
	if nil == ctx {
		return flux.WrapObjectMTValue(nil), errors.New("lookup nil context")
	}
	req := ctx.Request()
	switch strings.ToUpper(scope) {
	case flux.ScopePath:
		return flux.WrapTextMTValue(req.PathValue(key)), nil
	case flux.ScopePathMap:
		return flux.WrapStrValuesMapMTValue(req.PathValues()), nil
	case flux.ScopeQuery:
		return flux.WrapTextMTValue(req.QueryValue(key)), nil
	case flux.ScopeQueryMap:
		return flux.WrapStrValuesMapMTValue(req.QueryValues()), nil
	case flux.ScopeForm:
		return flux.WrapTextMTValue(req.FormValue(key)), nil
	case flux.ScopeFormMap:
		return flux.WrapStrValuesMapMTValue(req.FormValues()), nil
	case flux.ScopeHeader:
		return flux.WrapTextMTValue(req.HeaderValue(key)), nil
	case flux.ScopeHeaderMap:
		header, _ := req.HeaderValues()
		return flux.WrapStrValuesMapMTValue(header), nil
	case flux.ScopeAttr:
		v, _ := ctx.GetAttribute(key)
		return flux.WrapObjectMTValue(v), nil
	case flux.ScopeAttrs:
		return flux.WrapStrMapMTValue(ctx.Attributes()), nil
	case flux.ScopeBody:
		reader, err := req.RequestBodyReader()
		return flux.MTValue{Value: reader, MediaType: req.HeaderValue(flux.HeaderContentType)}, err
	case flux.ScopeParam:
		v, _ := SearchValueProviders(key, req.QueryValues, req.FormValues)
		return flux.WrapTextMTValue(v), nil
	case flux.ScopeValue:
		v, _ := ctx.GetValue(key)
		return flux.WrapObjectMTValue(v), nil
	case flux.ScopeAuto:
		fallthrough
	default:
		if v, ok := SearchValueProviders(key, req.PathValues, req.QueryValues, req.FormValues, HeaderProviderFunc(req)); ok {
			return flux.WrapTextMTValue(v), nil
		}
		if v, ok := ctx.GetAttribute(key); ok {
			return flux.WrapObjectMTValue(v), nil
		}
		if v, ok := ctx.GetValue(key); ok {
			return flux.WrapObjectMTValue(v), nil
		}
		return flux.WrapObjectMTValue(nil), nil
	}
}

// 默认实现：查找Argument的值解析函数
func DefaultArgumentValueResolveFunc(mtValue flux.MTValue, arg flux.Argument, ctx flux.Context) (interface{}, error) {
	valueResolver := ext.LoadMTValueResolver(arg.Class)
	if nil == valueResolver {
		logger.TraceContext(ctx).Warnw("Not supported argument type",
			"http.key", arg.HttpName, "arg.name", arg.Name, "resolver-class", arg.Class, "generic", arg.Generic)
		valueResolver = ext.LoadMTValueDefaultResolver()
	}
	if value, err := valueResolver(mtValue, arg.Class, arg.Generic); nil != err {
		logger.TraceContext(ctx).Warnw("Failed to resolve argument",
			"http.key", arg.HttpName, "arg.name", arg.Name, "value-class", arg.Class, "generic", arg.Generic,
			"mime.value", mtValue.Value, "mime.type", mtValue.MediaType, "error", err)
		return nil, fmt.Errorf("PARAMETERS:RESOLVE_VALUE:%w", err)
	} else {
		return value, nil
	}
}
