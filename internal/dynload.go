package internal

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/extension"
	"github.com/bytepowered/flux/logger"
)

type configurable struct {
	Name    string
	Type    string
	Config  flux.Config
	Factory flux.Factory
}

func dynloadConfig(globals flux.Config) []configurable {
	out := make([]configurable, 0)
	globals.Foreach(func(name string, v interface{}) bool {
		m, is := v.(map[string]interface{})
		if !is {
			return true
		}
		config := NewMapConfig(m)
		if config.IsEmpty() || !(config.Contains("type") && config.Contains("InitConfig")) {
			return true
		}
		typeName := config.String("type")
		if config.BooleanOrDefault("disable", false) {
			logger.Infof("Component is DISABLED, type: %s", typeName)
			return true
		}
		f, ok := extension.GetFactory(typeName)
		if !ok {
			logger.Warnf("Config factory not found, type: %s", typeName)
			return true
		}
		out = append(out, configurable{
			Name:    name,
			Type:    typeName,
			Config:  extension.ConfigFactory()("flux.component."+typeName, config.Map("InitConfig")),
			Factory: f,
		})
		return true
	})
	return out
}
