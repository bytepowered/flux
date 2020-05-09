package flux

import "github.com/bytepowered/flux/pkg"

// ConfigurationFactory 根据指定ns和map数据构建Configuration对象。可用于生成多数据源配置对象。
type ConfigurationFactory func(namespace string, data map[string]interface{}) Configuration

// Configuration 支持按类型读取数值的配置接口
type Configuration interface {
	IsEmpty() bool
	GetValue(name string) (interface{}, bool)
	String(name string) string
	StringOrDefault(name string, defaultValue string) string
	Int64(name string) int64
	Int64OrDefault(name string, defaultValue int64) int64
	Float64(name string) float64
	Float64OrDefault(name string, defaultValue float64) float64
	Boolean(name string) bool
	BooleanOrDefault(name string, defaultValue bool) bool
	Map(name string) map[string]interface{}
	Contains(name string) bool
	Foreach(f func(key string, value interface{}) bool)
}

// MapConfiguration实现
func NewMapConfiguration(values map[string]interface{}) Configuration {
	return &MapConfiguration{values}
}

type MapConfiguration struct {
	values map[string]interface{}
}

func (c *MapConfiguration) GetValue(name string) (interface{}, bool) {
	v, ok := c.values[name]
	return v, ok
}

func (c *MapConfiguration) IsEmpty() bool {
	return 0 == len(c.values)
}

func (c *MapConfiguration) Foreach(fun func(key string, val interface{}) bool) {
	for k, v := range c.values {
		if !fun(k, v) {
			return
		}
	}
}

func (c *MapConfiguration) String(name string) string {
	return c.StringOrDefault(name, "")
}

func (c *MapConfiguration) StringOrDefault(name string, defaultValue string) string {
	if v, ok := c.GetValue(name); ok {
		return pkg.ToString(v)
	} else {
		return defaultValue
	}
}

func (c *MapConfiguration) Int64(name string) int64 {
	return c.Int64OrDefault(name, int64(0))
}

func (c *MapConfiguration) Int64OrDefault(name string, defaultValue int64) int64 {
	if v, ok := c.GetValue(name); ok {
		if iv, e := pkg.ToInt64(v); nil != e {
			return defaultValue
		} else {
			return iv
		}
	} else {
		return defaultValue
	}
}

func (c *MapConfiguration) Float64(name string) float64 {
	return c.Float64OrDefault(name, float64(0))
}

func (c *MapConfiguration) Float64OrDefault(name string, defaultValue float64) float64 {
	if v, ok := c.GetValue(name); ok {
		if fv, e := pkg.ToFloat64(v); nil != e {
			return defaultValue
		} else {
			return fv
		}
	} else {
		return defaultValue
	}
}

func (c *MapConfiguration) Boolean(name string) bool {
	return c.BooleanOrDefault(name, false)
}

func (c *MapConfiguration) BooleanOrDefault(name string, defaultValue bool) bool {
	if v, ok := c.GetValue(name); ok {
		return pkg.ToBool(v)
	} else {
		return defaultValue
	}
}

func (c *MapConfiguration) Map(name string) map[string]interface{} {
	if v, ok := c.GetValue(name); ok {
		return v.(map[string]interface{})
	} else {
		return make(map[string]interface{})
	}
}

func (c *MapConfiguration) Contains(name string) bool {
	_, ok := c.values[name]
	return ok
}
