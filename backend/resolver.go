package backend

import (
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/logger"
)

func LookupResolveWith(arg flux.Argument, lookup flux.ArgumentValueLookupFunc, resolver flux.ArgumentValueResolveFunc, ctx flux.Context) (interface{}, error) {
	mtValue, err := lookup(arg.HttpScope, arg.HttpName, ctx)
	if nil != err {
		logger.TraceContext(ctx).Warnw("Failed to lookup argument",
			"http.scope", arg.HttpScope, "http.name", arg.HttpName, "arg.name", arg.Name, "error", err)
		return nil, fmt.Errorf("BACKEND:LOOKUP:%w", err)
	}
	value, err := resolver(mtValue, arg, ctx)
	if nil != err {
		logger.TraceContext(ctx).Warnw("Failed to resolve argument",
			"mime-value", mtValue, "arg.class", arg.Class, "error", err)
		return nil, fmt.Errorf("BACKEND:RESOLVE:%w", err)
	}
	return value, err
}
