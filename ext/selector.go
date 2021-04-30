package ext

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/toolkit"
)

var (
	endpointSelectors = make([]flux.EndpointSelector, 0, 8)
)

func AddEndpointSelector(s flux.EndpointSelector) {
	toolkit.MustNotNil(s, "FilterSelector is nil")
	endpointSelectors = append(endpointSelectors, s)
}

func EndpointSelectors() []flux.EndpointSelector {
	out := make([]flux.EndpointSelector, len(endpointSelectors))
	copy(out, endpointSelectors)
	return out
}
