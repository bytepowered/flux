package ext

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/pkg"
)

const (
	EndpointDiscoveryProtoDefault   = "default"
	EndpointDiscoveryProtoZookeeper = "zookeeper"
)

var (
	endpointDiscoveryFactories = make(map[string]EndpointDiscoveryFactory, 2)
)

// EndpointDiscoveryFactory 用于构建EndpointRegistry的工厂函数。
type EndpointDiscoveryFactory func() flux.EndpointDiscovery

// StoreEndpointDiscoveryFactory 设置指定ID名的EndpointRegistry工厂函数。
func StoreEndpointDiscoveryFactory(protoName string, factory EndpointDiscoveryFactory) {
	protoName = pkg.RequireNotEmpty(protoName, "factory protoName is empty")
	endpointDiscoveryFactories[protoName] = pkg.RequireNotNil(factory, "EndpointDiscoveryFactory is nil").(EndpointDiscoveryFactory)
}

// LoadEndpointDiscoveryFactory 根据ID名，获取EndpointRegistry的工厂函数
func LoadEndpointDiscoveryFactory(protoName string) (EndpointDiscoveryFactory, bool) {
	protoName = pkg.RequireNotEmpty(protoName, "factory protoName is empty")
	e, ok := endpointDiscoveryFactories[protoName]
	return e, ok
}
