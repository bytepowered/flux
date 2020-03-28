package server

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/pkg"
	"sync"
)

const (
	ConfigDir = "conf.d"
	ConfigApp = ConfigDir + "/application.toml"
)

var (
	_globals    flux.Config
	_globalOnce = new(sync.Once)
)

// GlobalConfig 返回全局配置
func GlobalConfig() flux.Config {
	return _globals
}

func LoadConfig() flux.Config {
	_globalOnce.Do(func() {
		if data, err := pkg.LoadTomlFile(ConfigApp); nil != err {
			logger.Panicf("Config not found: %s", ConfigApp)
		} else {
			logger.Infof("Using config: %s", ConfigApp)
			_globals = ext.ConfigFactory()("globals", data)
		}
	})
	return _globals
}
