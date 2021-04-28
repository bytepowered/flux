package ext

import (
	"github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-pkg"
)

var (
	endpointSelectors = make([]flux.EndpointSelector, 0, 8)
)

func AddEndpointSelector(s flux.EndpointSelector) {
	fluxpkg.MustNotNil(s, "FilterSelector is nil")
	endpointSelectors = append(endpointSelectors, s)
}

func EndpointSelectors() []flux.EndpointSelector {
	out := make([]flux.EndpointSelector, len(endpointSelectors))
	copy(out, endpointSelectors)
	return out
}
