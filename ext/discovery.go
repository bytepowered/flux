package ext

import (
	"github.com/bytepowered/flux"
)

var (
	endpointDiscoveryMap = make(map[string]flux.EndpointDiscovery, 4)
)

func RegisterEndpointDiscovery(discovery flux.EndpointDiscovery) {
	endpointDiscoveryMap[discovery.Id()] = discovery
}

func EndpointDiscoveryById(id string) (flux.EndpointDiscovery, bool) {
	v, ok := endpointDiscoveryMap[id]
	return v, ok
}

func EndpointDiscoveries() []flux.EndpointDiscovery {
	out := make([]flux.EndpointDiscovery, 0, len(endpointDiscoveryMap))
	for _, d := range endpointDiscoveryMap {
		out = append(out, d)
	}
	return out
}
