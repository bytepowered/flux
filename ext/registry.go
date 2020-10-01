package ext

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/pkg"
)

// known ids of registry
const (
	EndpointRegistryIdDefault   = "active"
	EndpointRegistryIdZookeeper = "zookeeper"
)

var (
	_identityRegistryFactories = make(map[string]RegistryFactory, 2)
)

type RegistryFactory func() flux.EndpointRegistry

// SetEndpointRegistryFactory 设置指定ID名的Registry工厂函数。
func SetEndpointRegistryFactory(id string, factory RegistryFactory) {
	_identityRegistryFactories[id] = pkg.RequireNotNil(factory, "RegistryFactory is nil").(RegistryFactory)
}

// GetRegistryFactory 根据ID名，获取Registry的工厂函数
func GetRegistryFactory(id string) (RegistryFactory, bool) {
	e, ok := _identityRegistryFactories[id]
	return e, ok
}
