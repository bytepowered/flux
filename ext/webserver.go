package ext

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/pkg"
)

var (
	listenServerFactory flux.ListenServerFactory
	endpointSelectors   = make([]flux.EndpointSelector, 0, 8)
)

func SetListenServerFactory(f flux.ListenServerFactory) {
	listenServerFactory = f
}

func ListenServerFactory() flux.ListenServerFactory {
	return listenServerFactory
}

func AddEndpointSelector(s flux.EndpointSelector) {
	pkg.RequireNotNil(s, "FilterSelector is nil")
	endpointSelectors = append(endpointSelectors, s)
}

func EndpointSelectors() []flux.EndpointSelector {
	out := make([]flux.EndpointSelector, len(endpointSelectors))
	copy(out, endpointSelectors)
	return out
}
