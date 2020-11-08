package http

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/backend"
	"github.com/bytepowered/flux/ext"
	"github.com/spf13/cast"
	"net/url"
)

func _toHttpUrlValues(arguments []flux.Argument, ctx flux.Context) (url.Values, error) {
	values := make(url.Values, len(arguments))
	lookup := ext.GetArgumentValueLookupFunc()
	resolver := ext.GetArgumentValueResolveFunc()
	for _, arg := range arguments {
		if value, err := backend.LookupResolveWith(arg, lookup, resolver, ctx); nil != err {
			return nil, err
		} else {
			values.Add(arg.Name, cast.ToString(value))
		}
	}
	return values, nil
}
