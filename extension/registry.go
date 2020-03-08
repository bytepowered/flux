package extension

import "github.com/bytepowered/flux"

// name of registry
const (
	TypeNameRegistryActive    = "active"
	TypeNameRegistryZookeeper = "zookeeper"
	TypeNameRegistryEcho      = "echo"
)

var (
	_protoNamedRegistryFactories = make(map[string]RegistryFactory, 2)
)

type RegistryFactory func() flux.Registry

// SetRegistryFactory 设置指定协议名的Registry工厂函数。此函数会自动注册生命周期Hook
func SetRegistryFactory(protocol string, factory RegistryFactory) {
	_protoNamedRegistryFactories[protocol] = func() flux.Registry {
		registry := factory()
		return registry
	}
}

// GetRegistryFactory 根据Protocol协议名，获取Registry的工厂函数
func GetRegistryFactory(protocol string) (RegistryFactory, bool) {
	e, ok := _protoNamedRegistryFactories[protocol]
	return e, ok
}
