package ext

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/pkg"
)

const (
	EndpointRegistryProtoDefault   = "default"
	EndpointRegistryProtoZookeeper = "zookeeper"
)

var (
	protoRegistryFactories = make(map[string]EndpointRegistryFactory, 2)
)

// EndpointRegistryFactory 用于构建EndpointRegistry的工厂函数。
type EndpointRegistryFactory func() flux.EndpointRegistry

// StoreEndpointRegistryFactory 设置指定ID名的EndpointRegistry工厂函数。
func StoreEndpointRegistryFactory(protoName string, factory EndpointRegistryFactory) {
	protoName = pkg.RequireNotEmpty(protoName, "factory protoName is empty")
	protoRegistryFactories[protoName] = pkg.RequireNotNil(factory, "EndpointRegistryFactory is nil").(EndpointRegistryFactory)
}

// LoadEndpointRegistryFactory 根据ID名，获取EndpointRegistry的工厂函数
func LoadEndpointRegistryFactory(protoName string) (EndpointRegistryFactory, bool) {
	protoName = pkg.RequireNotEmpty(protoName, "factory protoName is empty")
	e, ok := protoRegistryFactories[protoName]
	return e, ok
}
