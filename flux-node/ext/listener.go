package ext

import (
	"github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-pkg"
)

var (
	webListenerFactory flux.WebListenerFactory
	endpointSelectors  = make([]flux.EndpointSelector, 0, 8)
)

func SetWebListenerFactory(f flux.WebListenerFactory) {
	webListenerFactory = f
}

func WebListenerFactory() flux.WebListenerFactory {
	return webListenerFactory
}

func AddEndpointSelector(s flux.EndpointSelector) {
	fluxpkg.MustNotNil(s, "FilterSelector is nil")
	endpointSelectors = append(endpointSelectors, s)
}

func EndpointSelectors() []flux.EndpointSelector {
	out := make([]flux.EndpointSelector, len(endpointSelectors))
	copy(out, endpointSelectors)
	return out
}
