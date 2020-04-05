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

type AwareConfig struct {
	Name     string
	TypeId   string
	ConfigNs string
	Config   flux.Config
	Factory  flux.Factory
}

func dynloadConfig(globals flux.Config) []AwareConfig {
	out := make([]AwareConfig, 0)
	globals.Foreach(func(name string, v interface{}) bool {
		m, is := v.(map[string]interface{})
		if !is {
			return true
		}
		config := NewMapConfig(m)
		if config.IsEmpty() || !(config.Contains(configKeyTypeId) && config.Contains(configKeyInitConfig)) {
			return true
		}
		typeId := config.String(configKeyTypeId)
		if config.BooleanOrDefault(configKeyDisable, false) {
			logger.Infof("Component is DISABLED, type: %s", typeId)
			return true
		}
		factory, ok := ext.GetFactory(typeId)
		if !ok {
			logger.Warnf("Config factory not found, type: %s", typeId)
			return true
		}
		ns := configNsPrefixComponent + typeId
		out = append(out, AwareConfig{
			Name:     name,
			TypeId:   typeId,
			ConfigNs: ns,
			Config:   ext.ConfigFactory()(ns, config.Map(configKeyInitConfig)),
			Factory:  factory,
		})
		return true
	})
	return out
}
