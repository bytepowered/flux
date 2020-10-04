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
			logger.Trace(ctx.RequestId()).Warnw("Unsupported argument type",
				"class", arg.TypeClass, "generic", arg.TypeGeneric, "type", arg.Type)
		}
	}
	return nil
}

func doResolveWith(lookupFunc flux.ArgumentValueLookupFunc, arg flux.Argument, ctx flux.Context) *flux.StateError {
	value, err := lookupFunc(arg.HttpScope, arg.HttpKey, ctx)
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
	resolver := ext.GetTypedValueResolver(arg.TypeClass)
	if nil == resolver {
		logger.Trace(ctx.RequestId()).Warnw("Not supported argument type",
			"http.key", arg.HttpKey, "arg.name", arg.Name, "class", arg.TypeClass, "generic", arg.TypeGeneric)
		resolver = ext.GetDefaultTypedValueResolver()
	}
	if typedValue, err := resolver(arg.TypeClass, arg.TypeGeneric, value); nil != err {
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
		arg.HttpValue.SetValue(typedValue)
		return nil
	}
}
