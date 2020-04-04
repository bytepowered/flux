package internal

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
)

const (
	configKeyDisable    = "disable"
	configKeyTypeId     = "type-id"
	configKeyInitConfig = "InitConfig"
)

type awareConfig struct {
	Name     string
	Type     string
	ConfigNs string
	Config   flux.Config
	Factory  flux.Factory
}

func dynloadConfig(globals flux.Config) []awareConfig {
	out := make([]awareConfig, 0)
	globals.Foreach(func(name string, v interface{}) bool {
		m, is := v.(map[string]interface{})
		if !is {
			return true
		}
		config := NewMapConfig(m)
		if config.IsEmpty() || !(config.Contains(configKeyTypeId) && config.Contains(configKeyInitConfig)) {
			return true
		}
		typeName := config.String(configKeyTypeId)
		if config.BooleanOrDefault(configKeyDisable, false) {
			logger.Infof("Component is DISABLED, type: %s", typeName)
			return true
		}
		f, ok := ext.GetFactory(typeName)
		if !ok {
			logger.Warnf("Config factory not found, type: %s", typeName)
			return true
		}
		cns := configNsPrefixComponent + typeName
		out = append(out, awareConfig{
			Name:     name,
			Type:     typeName,
			ConfigNs: cns,
			Config:   ext.ConfigFactory()(cns, config.Map(configKeyInitConfig)),
			Factory:  f,
		})
		return true
	})
	return out
}
