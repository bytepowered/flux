package extension

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/pkg"
	"sync"
)

const (
	ConfigDir = "conf.d"
	ConfigApp = ConfigDir + "/application.toml"
)

var (
	_globals    = flux.Config{}
	_globalOnce = new(sync.Once)
)

// GetGlobalConfig 返回全局配置
func GetGlobalConfig() flux.Config {
	return _globals
}

func LoadConfig() flux.Config {
	_globalOnce.Do(func() {
		if data, err := pkg.LoadTomlConfig(ConfigApp); nil != err {
			GetLogger().Panicf("Config not found: %s", ConfigApp)
		} else {
			GetLogger().Infof("Using config: %s", ConfigApp)
			_globals = data
		}
	})
	return _globals
}
