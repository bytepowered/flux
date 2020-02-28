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

func dynloadConfig() []configurable {
	out := make([]configurable, 0)
	for name, v := range extension.GetGlobalConfig() {
		m, is := v.(map[string]interface{})
		if !is {
			continue
		}
		config := flux.ToConfig(m)
		if config.IsEmpty() || !(config.Contains("type") && config.Contains("InitConfig")) {
			continue
		}
		typeName := config.String("type")
		if config.BooleanOrDefault("disable", false) {
			logger.Infof("Component is DISABLED, type: %s", typeName)
			continue
		}
		f, ok := extension.GetFactory(typeName)
		if !ok {
			logger.Warnf("Config factory not found, type: %s", typeName)
			continue
		}
		out = append(out, configurable{
			Name:    name,
			Type:    typeName,
			Config:  config.Config("InitConfig"),
			Factory: f,
		})
	}
	return out
}
