package http

import (
	"github.com/bytepowered/flux"
	"github.com/spf13/cast"
	"net/url"
)

func AssembleHttpValues(arguments []flux.Argument, ctx flux.Context) (url.Values, error) {
	values := make(url.Values, len(arguments))
	for _, arg := range arguments {
		if val, err := arg.Resolve(ctx); nil != err {
			return nil, err
		} else {
			values.Add(arg.Name, cast.ToString(val))
		}
	}
	return values, nil
}
