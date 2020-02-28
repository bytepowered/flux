package extension

import "github.com/bytepowered/flux"

// name of registry
const (
	TypeNameRegistryActive    = "active"
	TypeNameRegistryZookeeper = "zookeeper"
	TypeNameRegistryEcho      = "echo"
)

var (
	_registryFactories = make(map[string]RegistryFactory, 2)
)

type RegistryFactory func() flux.Registry

// SetRegistryFactory 设置指定协议名的Registry工厂函数。此函数会自动注册生命周期Hook
func SetRegistryFactory(protocol string, factory RegistryFactory) {
	_registryFactories[protocol] = func() flux.Registry {
		registry := factory()
		return registry
	}
}

func GetRegistryFactory(protocol string) (RegistryFactory, bool) {
	e, ok := _registryFactories[protocol]
	return e, ok
}
