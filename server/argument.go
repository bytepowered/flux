package server

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/pkg"
	https "net/http"
)

func shouldResolve(ctx flux.Context, args []flux.Argument) bool {
	// HEAD, OPTIONS 不需要解析参数
	if https.MethodHead == ctx.RequestMethod() || https.MethodOptions == ctx.RequestMethod() {
		return false
	}
	if len(args) == 0 {
		return false
	}
	return true
}

func resolveArguments(argumentLookupFunc ext.ArgumentLookupFunc, arguments []flux.Argument, ctx flux.Context) *flux.InvokeError {
	trace := logger.Trace(ctx.RequestId())
	for _, arg := range arguments {
		if flux.ArgumentTypePrimitive == arg.Type {
			value, err := argumentLookupFunc(arg, ctx)
			if nil != err {
				trace.Warnw("Failed to lookup argument",
					"http.key", arg.HttpKey, "arg.name", arg.Name, "error", err)
				return &flux.InvokeError{
					StatusCode: flux.StatusServerError,
					ErrorCode:  flux.ErrorCodeGatewayInternal,
					Message:    "PARAMETERS:LOOKUP",
					Internal:   err,
				}
			}
			valueResolver := pkg.GetValueResolver(arg.TypeClass)
			if nil == valueResolver {
				valueResolver = pkg.GetDefaultResolver()
			}
			if v, err := valueResolver(arg.TypeClass, arg.TypeGeneric, value); nil != err {
				trace.Warnw("Failed to resolve argument",
					"http.key", arg.HttpKey, "arg.name", arg.Name, "class", arg.TypeClass, "generic", arg.TypeGeneric,
					"value", value, "error", err)
				return &flux.InvokeError{
					StatusCode: flux.StatusServerError,
					ErrorCode:  flux.ErrorCodeGatewayInternal,
					Message:    "PARAMETERS:RESOLVE",
					Internal:   err,
				}
			} else {
				arg.HttpValue.SetValue(v)
			}
		} else if flux.ArgumentTypeComplex == arg.Type {
			if err := resolveArguments(argumentLookupFunc, arg.Fields, ctx); nil != err {
				return err
			}
		} else {
			trace.Warnw("Unsupported argument type",
				"class", arg.TypeClass, "generic", arg.TypeGeneric, "type", arg.Type)
		}
	}
	return nil
}
