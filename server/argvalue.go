package server

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"net/url"
)

// 默认实现：查找Argument的值函数
func DefaultArgumentValueResolver(scope, key string, ctx flux.Context) (value flux.TypedValue, err error) {
	request := ctx.Request()
	switch scope {
	case flux.ScopeQuery:
		return _wrapTextTypeValue(request.QueryValue(key)), nil
	case flux.ScopePath:
		return _wrapTextTypeValue(request.PathValue(key)), nil
	case flux.ScopeHeader:
		return _wrapTextTypeValue(request.HeaderValue(key)), nil
	case flux.ScopeForm:
		return _wrapTextTypeValue(request.FormValue(key)), nil
	case flux.ScopeBody:
		reader, err := request.RequestBodyReader()
		return flux.TypedValue{Value: reader, MIMEType: request.HeaderValue("Content-Type")}, err
	case flux.ScopeAttrs:
		return flux.TypedValue{Value: ctx.Attributes(), MIMEType: flux.ValueMIMETypeLangStringMap}, nil
	case flux.ScopeAttr:
		v, _ := ctx.GetAttribute(key)
		return _wrapObjectTypeValue(v), nil
	case flux.ScopeParam:
		if v := request.QueryValue(key); "" != v {
			return _wrapTextTypeValue(v), nil
		} else {
			return _wrapTextTypeValue(request.FormValue(key)), nil
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
			return _wrapTextTypeValue(v), nil
		} else if v := request.HeaderValue(key); "" != v {
			return _wrapTextTypeValue(v), nil
		} else if v, _ := ctx.GetAttribute(key); "" != v {
			return _wrapObjectTypeValue(v), nil
		} else {
			return _wrapObjectTypeValue(value), nil
		}
	}
}

func resolveArgumentWith(lookupFunc flux.ArgumentValueResolver, arguments []flux.Argument, ctx flux.Context) *flux.StateError {
	for _, arg := range arguments {
		if flux.ArgumentTypePrimitive == arg.Type {
			if err := _doResolve(lookupFunc, arg, ctx); nil != err {
				return err
			}
		} else if flux.ArgumentTypeComplex == arg.Type {
			if err := resolveArgumentWith(lookupFunc, arg.Fields, ctx); nil != err {
				return err
			}
		} else {
			logger.TraceContext(ctx).Warnw("Unsupported argument type",
				"class", arg.TypeClass, "generic", arg.TypeGeneric, "type", arg.Type)
		}
	}
	return nil
}

func _wrapTextTypeValue(value interface{}) flux.TypedValue {
	return flux.TypedValue{Value: value, MIMEType: flux.ValueMIMETypeLangText}
}

func _wrapObjectTypeValue(value interface{}) flux.TypedValue {
	return flux.TypedValue{Value: value, MIMEType: flux.ValueMIMETypeLangObject}
}

func _doResolve(resolver flux.ArgumentValueResolver, arg flux.Argument, ctx flux.Context) *flux.StateError {
	mtValue, err := resolver(arg.HttpScope, arg.HttpName, ctx)
	if nil != err {
		logger.TraceContext(ctx).Warnw("Failed to lookup argument",
			"http.key", arg.HttpName, "arg.name", arg.Name, "error", err)
		return &flux.StateError{
			StatusCode: flux.StatusServerError,
			ErrorCode:  flux.ErrorCodeGatewayInternal,
			Message:    "PARAMETERS:LOOKUP_VALUE",
			Internal:   err,
		}
	}
	valueResolver := ext.GetTypedValueResolver(arg.TypeClass)
	if nil == valueResolver {
		logger.TraceContext(ctx).Warnw("Not supported argument type",
			"http.key", arg.HttpName, "arg.name", arg.Name, "class", arg.TypeClass, "generic", arg.TypeGeneric)
		valueResolver = ext.GetDefaultTypedValueResolver()
	}
	if value, err := valueResolver(arg.TypeClass, arg.TypeGeneric, mtValue); nil != err {
		logger.TraceContext(ctx).Warnw("Failed to resolve argument",
			"http.key", arg.HttpName, "arg.name", arg.Name, "class", arg.TypeClass, "generic", arg.TypeGeneric,
			"http.value", mtValue.Value, "error", err)
		return &flux.StateError{
			StatusCode: flux.StatusServerError,
			ErrorCode:  flux.ErrorCodeGatewayInternal,
			Message:    "PARAMETERS:RESOLVE_VALUE",
			Internal:   err,
		}
	} else {
		arg.HttpValue.SetValue(value)
		return nil
	}
}
