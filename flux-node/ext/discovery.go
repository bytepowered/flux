package ext

import (
	flux2 "github.com/bytepowered/flux/flux-node"
)

var (
	endpointDiscoveryMap = make(map[string]flux2.EndpointDiscovery, 4)
)

func RegisterEndpointDiscovery(discovery flux2.EndpointDiscovery) {
	endpointDiscoveryMap[discovery.Id()] = discovery
}

func EndpointDiscoveryById(id string) (flux2.EndpointDiscovery, bool) {
	v, ok := endpointDiscoveryMap[id]
	return v, ok
}

func EndpointDiscoveries() []flux2.EndpointDiscovery {
	out := make([]flux2.EndpointDiscovery, 0, len(endpointDiscoveryMap))
	for _, d := range endpointDiscoveryMap {
		out = append(out, d)
	}
	return out
}
