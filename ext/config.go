package ext

import "github.com/bytepowered/flux"

var (
	_configFactory flux.ConfigFactory
)

func SetConfigFactory(factory flux.ConfigFactory) {
	_configFactory = factory
}

func ConfigFactory() flux.ConfigFactory {
	return _configFactory
}
