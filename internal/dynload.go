package internal

import (
	"github.com/bytepowered/flux"
)

const (
	configKeyDynDisable    = "disable"
	configKeyDynTypeId     = "type-id"
	configKeyDynInitConfig = "InitConfig"
)

type AwareConfig struct {
	Name    string
	TypeId  string
	Factory flux.Factory
}

// dynloadConfig 基于type-id标记的工厂函数，可以生成相同类型的多实例组件
func dynloadConfig() []AwareConfig {
	out := make([]AwareConfig, 0)
	//globals.Foreach(func(name string, v interface{}) bool {
	//	m, is := v.(map[string]interface{})
	//	if !is {
	//		return true
	//	}
	//	config := flux.NewMapConfiguration(m)
	//	if config.IsEmpty() || !(config.Contains(configKeyDynTypeId) && config.Contains(configKeyDynInitConfig)) {
	//		return true
	//	}
	//	typeId := config.String(configKeyDynTypeId)
	//	if config.BooleanOrDefault(configKeyDynDisable, false) {
	//		logger.Infof("Aware is DISABLED, typeId: %s", typeId)
	//		return true
	//	}
	//	factory, ok := ext.GetFactory(typeId)
	//	if !ok {
	//		logger.Warnf("TypeFactory not found, typeId: %s", typeId)
	//		return true
	//	}
	//	ns := configNsPrefixComponent + typeId
	//	out = append(out, AwareConfig{
	//		Name:      name,
	//		TypeId:    typeId,
	//		Namespace: ns,
	//		Config:    ext.ConfigurationFactory()(ns, config.Map(configKeyDynInitConfig)),
	//		Factory:   factory,
	//	})
	//	return true
	//})
	return out
}
