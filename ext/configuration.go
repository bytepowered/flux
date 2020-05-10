package ext

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/pkg"
)

var (
	_configurationFactory flux.ConfigurationFactory
)

func SetConfigurationFactory(factory flux.ConfigurationFactory) {
	_configurationFactory = pkg.RequireNotNil(factory, "ConfigurationFactory is nil").(flux.ConfigurationFactory)
}

func ConfigurationFactory() flux.ConfigurationFactory {
	return _configurationFactory
}
