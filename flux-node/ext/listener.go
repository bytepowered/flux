package ext

import (
	flux2 "github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-pkg"
)

var (
	webListenerFactory flux2.WebListenerFactory
	endpointSelectors  = make([]flux2.EndpointSelector, 0, 8)
)

func SetWebListenerFactory(f flux2.WebListenerFactory) {
	webListenerFactory = f
}

func WebListenerFactory() flux2.WebListenerFactory {
	return webListenerFactory
}

func AddEndpointSelector(s flux2.EndpointSelector) {
	fluxpkg.MustNotNil(s, "FilterSelector is nil")
	endpointSelectors = append(endpointSelectors, s)
}

func EndpointSelectors() []flux2.EndpointSelector {
	out := make([]flux2.EndpointSelector, len(endpointSelectors))
	copy(out, endpointSelectors)
	return out
}
