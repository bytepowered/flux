package internal

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/pkg"
	"net/http"
)

// 默认实现：查找Argument的值函数
func DefaultArgumentValueLookupFunc(arg flux.Argument, ctx flux.Context) (interface{}, error) {
	request := ctx.Request()
	switch arg.HttpScope {
	case flux.ScopeQuery:
		return request.QueryValue(arg.HttpKey), nil
	case flux.ScopePath:
		return request.PathValue(arg.HttpKey), nil
	case flux.ScopeParam:
		if v := request.QueryValue(arg.HttpKey); "" == v {
			return request.FormValue(arg.HttpKey), nil
		} else {
			return v, nil
		}
	case flux.ScopeHeader:
		return request.HeaderValue(arg.HttpKey), nil
	case flux.ScopeForm:
		return request.FormValue(arg.HttpKey), nil
	case flux.ScopeAttrs:
		return ctx.Attributes(), nil
	case flux.ScopeAttr:
		value, _ := ctx.GetAttribute(arg.HttpKey)
		return value, nil
	case flux.ScopeAuto:
		fallthrough
	default:
		if v := request.PathValue(arg.HttpKey); "" != v {
			return v, nil
		} else if v := request.QueryValue(arg.HttpKey); "" != v {
			return v, nil
		} else if v := request.FormValue(arg.HttpKey); "" != v {
			return v, nil
		} else if v := request.HeaderValue(arg.HttpKey); "" != v {
			return v, nil
		} else if v, _ := ctx.GetAttribute(arg.HttpKey); "" != v {
			return v, nil
		} else {
			return nil, nil
		}
	}
}

func shouldResolve(ctx flux.Context, args []flux.Argument) bool {
	// HEAD, OPTIONS 不需要解析参数
	if http.MethodHead == ctx.Method() || http.MethodOptions == ctx.Method() {
		return false
	}
	if len(args) == 0 {
		return false
	}
	return true
}

func resolveArguments(lookupFunc ext.ArgumentLookupFunc, arguments []flux.Argument, ctx flux.Context) *flux.StateError {
	for _, arg := range arguments {
		if flux.ArgumentTypePrimitive == arg.Type {
			if err := resolve(lookupFunc, arg, ctx); nil != err {
				return err
			}
		} else if flux.ArgumentTypeComplex == arg.Type {
			if err := resolveArguments(lookupFunc, arg.Fields, ctx); nil != err {
				return err
			}
		} else {
			logger.Trace(ctx.RequestId()).Warnw("Unsupported argument type",
				"class", arg.TypeClass, "generic", arg.TypeGeneric, "type", arg.Type)
		}
	}
	return nil
}

func resolve(lookupFunc ext.ArgumentLookupFunc, arg flux.Argument, ctx flux.Context) *flux.StateError {
	value, err := lookupFunc(arg, ctx)
	if nil != err {
		logger.Trace(ctx.RequestId()).Warnw("Failed to lookup argument",
			"http.key", arg.HttpKey, "arg.name", arg.Name, "error", err)
		return &flux.StateError{
			StatusCode: flux.StatusServerError,
			ErrorCode:  flux.ErrorCodeGatewayInternal,
			Message:    "PARAMETERS:LOOKUP_VALUE",
			Internal:   err,
		}
	}
	valueResolver := pkg.GetValueResolver(arg.TypeClass)
	if nil == valueResolver {
		valueResolver = pkg.GetDefaultResolver()
	}
	if v, err := valueResolver(arg.TypeClass, arg.TypeGeneric, value); nil != err {
		logger.Trace(ctx.RequestId()).Warnw("Failed to resolve argument",
			"http.key", arg.HttpKey, "arg.name", arg.Name, "class", arg.TypeClass, "generic", arg.TypeGeneric,
			"value", value, "error", err)
		return &flux.StateError{
			StatusCode: flux.StatusServerError,
			ErrorCode:  flux.ErrorCodeGatewayInternal,
			Message:    "PARAMETERS:RESOLVE_VALUE",
			Internal:   err,
		}
	} else {
		arg.HttpValue.SetValue(v)
		return nil
	}
}
