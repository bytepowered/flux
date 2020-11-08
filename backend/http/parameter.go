package http

import (
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"github.com/spf13/cast"
	"net/url"
)

func _toHttpUrlValues(arguments []flux.Argument, ctx flux.Context) (url.Values, error) {
	values := make(url.Values)
	lookup := ext.GetArgumentValueLookupFunc()
	resolver := ext.GetArgumentValueResolveFunc()
	for _, arg := range arguments {
		mtValue, err := lookup(arg.HttpScope, arg.HttpName, ctx)
		if nil != err {
			logger.TraceContext(ctx).Warnw("Failed to lookup argument",
				"http.key", arg.HttpName, "arg.name", arg.Name, "error", err)
			return nil, fmt.Errorf("ASSEMBLE:LOOKUP:%w", err)
		}
		if v, err := resolver(mtValue, arg, ctx); nil != err {
			return nil, err
		} else {
			values.Add(arg.Name, cast.ToString(v))
		}
	}
	return values, nil
}
