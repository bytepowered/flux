package ext

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/pkg"
)

var (
	webServerFactory  WebServerFactory
	endpointSelectors = make([]flux.EndpointSelector, 0, 8)
)

type WebServerFactory func(*flux.Configuration) flux.ListenServer

func SetWebServerFactory(f WebServerFactory) {
	webServerFactory = f
}

func GetWebServerFactory() WebServerFactory {
	return webServerFactory
}

func AddEndpointSelector(s flux.EndpointSelector) {
	pkg.RequireNotNil(s, "FilterSelector is nil")
	endpointSelectors = append(endpointSelectors, s)
}

func ActiveEndpointSelectors() []flux.EndpointSelector {
	out := make([]flux.EndpointSelector, len(endpointSelectors))
	copy(out, endpointSelectors)
	return out
}
