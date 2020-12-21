package backend

import (
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/pkg"
)

func LookupResolveWith(arg flux.Argument, lookup flux.ArgumentValueLookupFunc, resolver flux.ArgumentValueResolveFunc, ctx flux.Context) (interface{}, error) {
	// Lookup
	var mtValue flux.MTValue
	if pkg.IsNotNil(arg.ValueLoader) {
		mtValue = arg.ValueLoader()
	} else {
		if mtv, err := lookup(arg.HttpScope, arg.HttpName, ctx); nil != err {
			logger.TraceContext(ctx).Warnw("Failed to lookup argument",
				"http.scope", arg.HttpScope, "http.name", arg.HttpName, "arg.name", arg.Name, "error", err)
			return nil, fmt.Errorf("BACKEND:LOOKUP:%w", err)
		} else {
			mtValue = mtv
		}
	}
	// Resolve
	value, err := resolver(mtValue, arg, ctx)
	if nil != err {
		logger.TraceContext(ctx).Warnw("Failed to resolve argument",
			"mime-value", mtValue, "arg.class", arg.Class, "error", err)
		return nil, fmt.Errorf("BACKEND:RESOLVE:%w", err)
	}
	return value, err
}
