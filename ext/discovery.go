package ext

import (
	"github.com/bytepowered/flux"
)

var (
	endpointDiscoveryMap = make(map[string]flux.EndpointDiscovery, 4)
)

func SetEndpointDiscovery(discovery flux.EndpointDiscovery) {
	endpointDiscoveryMap[discovery.Id()] = discovery
}

func GetEndpointDiscovery(id string) (flux.EndpointDiscovery, bool) {
	v, ok := endpointDiscoveryMap[id]
	return v, ok
}

func GetEndpointDiscoveries() []flux.EndpointDiscovery {
	out := make([]flux.EndpointDiscovery, 0, len(endpointDiscoveryMap))
	for _, d := range endpointDiscoveryMap {
		out = append(out, d)
	}
	return out
}
