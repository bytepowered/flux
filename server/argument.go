package server

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"net/http"
)

func argumentNeedResolve(ctx flux.Context, args []flux.Argument) bool {
	// HEAD, OPTIONS 不需要解析参数
	if http.MethodHead == ctx.Method() || http.MethodOptions == ctx.Method() {
		return false
	}
	if len(args) == 0 {
		return false
	}
	return true
}

func argumentResolveWith(lookupFunc flux.ArgumentValueLookupFunc, arguments []flux.Argument, ctx flux.Context) *flux.StateError {
	for _, arg := range arguments {
		if flux.ArgumentTypePrimitive == arg.Type {
			if err := doResolveWith(lookupFunc, arg, ctx); nil != err {
				return err
			}
		} else if flux.ArgumentTypeComplex == arg.Type {
			if err := argumentResolveWith(lookupFunc, arg.Fields, ctx); nil != err {
				return err
			}
		} else {
			logger.TraceContext(ctx).Warnw("Unsupported argument type",
				"class", arg.TypeClass, "generic", arg.TypeGeneric, "type", arg.Type)
		}
	}
	return nil
}

func doResolveWith(lookupFunc flux.ArgumentValueLookupFunc, arg flux.Argument, ctx flux.Context) *flux.StateError {
	mtValue, err := lookupFunc(arg.HttpScope, arg.HttpName, ctx)
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
