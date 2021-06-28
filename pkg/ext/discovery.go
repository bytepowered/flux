package ext

import (
	"github.com/bytepowered/fluxgo/pkg/flux"
)

var (
	metadataDiscoveryMap = make(map[string]flux.MetadataDiscovery, 4)
)

func RegisterMetadataDiscovery(discovery flux.MetadataDiscovery) {
	metadataDiscoveryMap[discovery.Id()] = discovery
}

func MetadataDiscoveryById(id string) (flux.MetadataDiscovery, bool) {
	v, ok := metadataDiscoveryMap[id]
	return v, ok
}

func MetadataDiscoveries() []flux.MetadataDiscovery {
	out := make([]flux.MetadataDiscovery, 0, len(metadataDiscoveryMap))
	for _, d := range metadataDiscoveryMap {
		out = append(out, d)
	}
	return out
}
