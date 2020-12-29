package ext

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/pkg"
)

const (
	EndpointRegistryIdDefault   = "default"
	EndpointRegistryIdZookeeper = "zookeeper"
)

var (
	identityRegistryFactories = make(map[string]EndpointRegistryFactory, 2)
)

// EndpointRegistryFactory 用于构建EndpointRegistry的工厂函数。
type EndpointRegistryFactory func() flux.EndpointRegistry

// StoreEndpointRegistryFactory 设置指定ID名的EndpointRegistry工厂函数。
func StoreEndpointRegistryFactory(id string, factory EndpointRegistryFactory) {
	id = pkg.RequireNotEmpty(id, "factory id is empty")
	identityRegistryFactories[id] = pkg.RequireNotNil(factory, "EndpointRegistryFactory is nil").(EndpointRegistryFactory)
}

// LoadEndpointRegistryFactory 根据ID名，获取EndpointRegistry的工厂函数
func LoadEndpointRegistryFactory(id string) (EndpointRegistryFactory, bool) {
	id = pkg.RequireNotEmpty(id, "factory id is empty")
	e, ok := identityRegistryFactories[id]
	return e, ok
}
