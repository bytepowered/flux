package extension

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/pkg"
	"os"
	"sync"
)

const (
	ConfigDir = "conf.d"
	ConfigApp = ConfigDir + "/application.toml"
)

var (
	_globalConfig = flux.Config{}
	_globalOnce   = new(sync.Once)
)

// GetGlobalConfig 返回全局配置
func GetGlobalConfig() flux.Config {
	return _globalConfig
}

// GetSectionConfig 返回指定配置优的配置
func GetSectionConfig(sectionName string) flux.Config {
	return _globalConfig.Config(sectionName)
}

// GetConfigRoot 返回配置文件的根目录
func GetConfigRoot() string {
	path, _ := os.Getwd()
	return path + "/" + ConfigDir
}

// GetConfigPath 返回指定配置的绝对路径
func GetConfigPath(config string) string {
	return GetConfigRoot() + "/" + config
}

func LoadConfig() {
	_globalOnce.Do(func() {
		if data, err := pkg.LoadTomlConfig(ConfigApp); nil != err {
			GetLogger().Panicf("Config not found: %s", ConfigApp)
		} else {
			GetLogger().Infof("Using config: %s", ConfigApp)
			_globalConfig = data
		}
	})
}
