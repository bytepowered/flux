package internal

import (
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"github.com/spf13/viper"
)

const (
	dynConfigKeyDisable = "disable"
	dynConfigKeyTypeId  = "type-id"
)

type AwareConfig struct {
	Id      string
	TypeId  string
	Config  flux.Configuration
	Factory flux.Factory
}

// 动态加载Filter
func dynamicFilters() ([]AwareConfig, error) {
	out := make([]AwareConfig, 0)
	for id := range viper.GetStringMap("FILTER") {
		config := viper.Sub("FILTER." + id)
		if !config.IsSet(dynConfigKeyTypeId) {
			logger.Infof("Filter[%] config without typeId", id)
			continue
		}
		typeId := config.GetString(dynConfigKeyTypeId)
		if config.GetBool(dynConfigKeyDisable) {
			logger.Infof("Filter is DISABLED, typeId: %s, id: %s", typeId, id)
			continue
		}
		factory, ok := ext.GetFactory(typeId)
		if !ok {
			return nil, fmt.Errorf("FilterFactory not found, typeId: %s, name: %s", typeId, id)
		}
		out = append(out, AwareConfig{
			Id:      id,
			TypeId:  typeId,
			Factory: factory,
			Config:  flux.NewConfiguration(config),
		})
	}
	return out, nil
}
