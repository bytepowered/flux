package http

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/pkg"
	"net/url"
)

func _toHttpUrlValues(arguments []flux.Argument) url.Values {
	values := make(url.Values)
	for _, kv := range arguments {
		values.Add(kv.Name, pkg.ToString(kv.HttpValue.Value()))
	}
	return values
}
