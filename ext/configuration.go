package ext

import (
	"github.com/bytepowered/flux"
)

var (
	_configurationFactory flux.ConfigurationFactory
)

func SetConfigurationFactory(factory flux.ConfigurationFactory) {
	_configurationFactory = factory
}

func ConfigurationFactory() flux.ConfigurationFactory {
	return _configurationFactory
}