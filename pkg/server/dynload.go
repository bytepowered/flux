package server

import (
	"fmt"
)

import (
	"github.com/bytepowered/fluxgo/pkg/ext"
	"github.com/bytepowered/fluxgo/pkg/flux"
	"github.com/bytepowered/fluxgo/pkg/logger"
)

const (
	dynConfigKeyDynamicFilter = "dynfilter"
	dynConfigKeyFilterId      = "id"
	dynConfigKeyFilterType    = "type"
)

type AwareConfig struct {
	Id      string
	TypeId  string
	Config  *flux.Configuration
	Factory flux.Factory
}

// dynamicFilters 动态加载Filter
func dynamicFilters() ([]AwareConfig, error) {
	out := make([]AwareConfig, 0)
	fconfig := flux.NewConfiguration(dynConfigKeyDynamicFilter)
	for _, config := range fconfig.ToConfigurations() {
		filterID := config.GetString(dynConfigKeyFilterId)
		if filterID == "" {
			logger.Infow("Filter configuration is empty or without filter-id", "filter-id", filterID)
			continue
		}
		filterType := config.GetString(dynConfigKeyFilterType)
		if IsDisabled(config) {
			logger.Infow("Filter is DISABLED", "filter-type", filterType, "filter-id", filterID)
			continue
		}
		factory, ok := ext.FactoryByType(filterType)
		if !ok {
			return nil, fmt.Errorf("FilterFactory not found, filter-type: %s, filter-id: %s", filterType, filterID)
		}
		out = append(out, AwareConfig{
			Id:      filterID,
			TypeId:  filterType,
			Factory: factory,
			Config:  config,
		})
	}
	return out, nil
}
