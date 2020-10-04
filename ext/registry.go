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
	_identityRegistryFactories = make(map[string]EndpointRegistryFactory, 2)
)

// EndpointRegistryFactory 用于构建EndpointRegistry的工厂函数。
type EndpointRegistryFactory func() flux.EndpointRegistry

// SetEndpointRegistryFactory 设置指定ID名的EndpointRegistry工厂函数。
func SetEndpointRegistryFactory(id string, factory EndpointRegistryFactory) {
	_identityRegistryFactories[id] = pkg.RequireNotNil(factory, "EndpointRegistryFactory is nil").(EndpointRegistryFactory)
}

// GetEndpointRegistryFactory 根据ID名，获取EndpointRegistry的工厂函数
func GetEndpointRegistryFactory(id string) (EndpointRegistryFactory, bool) {
	e, ok := _identityRegistryFactories[id]
	return e, ok
}
