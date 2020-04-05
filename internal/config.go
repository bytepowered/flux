package internal

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/pkg"
)

// 各个组件的配置命名空间前缀
const (
	configNsPrefixComponent     = "flux.component."
	configNsPrefixExchange      = "flux.exchanges"
	configNsPrefixExchangeProto = "flux.exchanges.proto."
	configNsPrefixRegistry      = "flux.registry"
)

type FxConfig struct {
	data map[string]interface{}
}

func NewMapConfig(values map[string]interface{}) flux.Config {
	return &FxConfig{
		data: values,
	}
}

func (c *FxConfig) IsEmpty() bool {
	return 0 == len(c.data)
}

func (c *FxConfig) Foreach(fun func(key string, val interface{}) bool) {
	for k, v := range c.data {
		if !fun(k, v) {
			return
		}
	}
}

func (c *FxConfig) String(name string) string {
	return c.StringOrDefault(name, "")
}

func (c *FxConfig) StringOrDefault(name string, defaultValue string) string {
	if v, ok := c.data[name]; ok {
		return pkg.ToString(v)
	} else {
		return defaultValue
	}
}

func (c *FxConfig) Int64(name string) int64 {
	return c.Int64OrDefault(name, int64(0))
}

func (c *FxConfig) Int64OrDefault(name string, defaultValue int64) int64 {
	if v, ok := c.data[name]; ok {
		if iv, e := pkg.ToInt64(v); nil != e {
			return defaultValue
		} else {
			return iv
		}
	} else {
		return defaultValue
	}
}

func (c *FxConfig) Float64(name string) float64 {
	return c.Float64OrDefault(name, float64(0))
}

func (c *FxConfig) Float64OrDefault(name string, defaultValue float64) float64 {
	if v, ok := c.data[name]; ok {
		if fv, e := pkg.ToFloat64(v); nil != e {
			return defaultValue
		} else {
			return fv
		}
	} else {
		return defaultValue
	}
}

func (c *FxConfig) Boolean(name string) bool {
	return c.BooleanOrDefault(name, false)
}

func (c *FxConfig) BooleanOrDefault(name string, defaultValue bool) bool {
	if v, ok := c.data[name]; ok {
		return pkg.ToBool(v)
	} else {
		return defaultValue
	}
}

func (c *FxConfig) Map(name string) map[string]interface{} {
	if v, ok := c.data[name]; ok {
		return v.(map[string]interface{})
	} else {
		return make(map[string]interface{})
	}
}

func (c *FxConfig) Contains(name string) bool {
	_, ok := c.data[name]
	return ok
}
