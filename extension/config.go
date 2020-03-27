package extension

import "github.com/bytepowered/flux"

var (
	_configFactory flux.ConfigFactory
)

func SetConfigFactory(config flux.ConfigFactory) {
	_configFactory = config
}

func ConfigFactory() flux.ConfigFactory {
	return _configFactory
}
