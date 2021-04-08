package flux

import (
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"os"
	"reflect"
	"strings"
	"time"
)

const (
	NamespaceWebListeners              = "web_listeners"
	NamespaceTransporters              = "transporters"
	NamespaceEndpointDiscoveryServices = "endpoint_discovery_services"
)

// NewGlobalConfiguration 创建全局Viper实例的配置对象
func NewGlobalConfiguration() *Configuration {
	return NewConfigurationOfViper(viper.GetViper())
}

// NewEmptyConfiguration 创建空的Viper实例的配置对象
func NewEmptyConfiguration() *Configuration {
	return NewConfigurationOfViper(viper.New())
}

// NewConfigurationOfMap 根据指定Map实例来构建。
func NewConfigurationOfMap(config map[string]interface{}) *Configuration {
	v := viper.New()
	for key, val := range config {
		v.Set(key, val)
	}
	return NewConfigurationOfViper(v)
}

// NewConfigurationOfNS 根据指定Namespace，在Viper全局配置中查找配置实例。如果NS不存在，新建一个空配置实例。
func NewConfigurationOfNS(namespace string) *Configuration {
	v := viper.Sub(namespace)
	if v == nil {
		v = viper.New()
	}
	return NewConfigurationOfViper(v)
}

// NewConfigurationOfViper 根据指定Viper实例来构建。如果Viper实例为nil，新建一个空配置实例。
func NewConfigurationOfViper(in *viper.Viper) *Configuration {
	if nil == in {
		in = viper.New()
	}
	return &Configuration{instance: in}
}

// Configuration 封装Viper实例访问接口的配置类
type Configuration struct {
	instance    *viper.Viper      // 实际的配置实例
	globalAlias map[string]string // 全局配置别名
}

// Reference 返回Viper实例
func (c *Configuration) Reference() *viper.Viper {
	return c.instance
}

// Sub 获取当前实例的子级配置对象
func (c *Configuration) Sub(name string) *Configuration {
	return NewConfigurationOfViper(c.instance.Sub(name))
}

func (c *Configuration) Get(key string) interface{} {
	return c.GetOrDefault(key, nil)
}

// GetOrDefault 查找指定Key的配置值。
// 从当前NS查询不到配置时，
// 1. 如果Value为动态Key，则根据动态Key读取全局配置；
// 2. 如果配置globalAlias映射，则根据AliasKey读取全局配置。
// 与Viper的Alias不同的是，Configuration的GlobalAlias是作用于局部命名空间下的别名映射。
// 当然，这不影响原有Viper的Alias功能。
func (c *Configuration) GetOrDefault(key string, def interface{}) interface{} {
	v := c.instance.Get(key)
	// 动态全局Key和默认值： ${username:yongjia}
	if strv, ok := v.(string); ok {
		dkey, defv, typ := ParseDynamicKey(strv)
		switch typ {
		case DynamicTypeConfig:
			if viper.IsSet(dkey) {
				return viper.Get(dkey)
			} else {
				return defv
			}
		case DynamicTypeEnv:
			if ev, ok := os.LookupEnv(dkey); ok {
				return ev
			} else {
				return defv
			}

		default:
			return v
		}
	}
	// GlobalAlias优先级低一些
	if nil == v && c.globalAlias != nil {
		if alias, ok := c.globalAlias[key]; ok {
			v = viper.Get(alias)
		}
	}
	if nil == v {
		v = def
	}
	return v
}

// Set 向当前配置实例以覆盖的方式设置Key-Value键值。
func (c *Configuration) Set(key string, value interface{}) {
	c.instance.Set(key, value)
}

// SetGlobalAlias 设置当前配置实例的Key与GlobalAlias的映射
// GlobalAlias 映射的Key是针对当前Configuration下的Key列表的映射。
// 如果在当前Configuration实例中查找不到值是时，将尝试使用GlobalAlias映射的Key，在全局对象中查找。
func (c *Configuration) SetGlobalAlias(globalAlias map[string]string) {
	c.globalAlias = globalAlias
}

// SetDefault 为当前配置实例设置单个默认值。与Viper的SetDefault一致，作用于当前配置实例。
func (c *Configuration) SetDefault(key string, value interface{}) {
	c.instance.SetDefault(key, value)
}

// SetDefault 为当前配置实例设置一组默认值。与Viper的SetDefault一致，作用于当前配置实例。
func (c *Configuration) SetDefaults(defaults map[string]interface{}) {
	for k, v := range defaults {
		c.instance.SetDefault(k, v)
	}
}

// IsSet 判定当前配置实例是否设置指定Key（多个）。与Viper的IsSet一致，查询范围为当前配置实例。
func (c *Configuration) IsSet(keys ...string) bool {
	if len(keys) == 0 {
		return false
	}
	// Any not set, return false
	for _, key := range keys {
		if !c.instance.IsSet(key) {
			return false
		}
	}
	return true
}

// GetString returns the value associated with the key as a string.
func (c *Configuration) GetString(key string) string {
	return cast.ToString(c.Get(key))
}

// GetBool returns the value associated with the key as a boolean.
func (c *Configuration) GetBool(key string) bool {
	return cast.ToBool(c.Get(key))
}

// GetInt returns the value associated with the key as an integer.
func (c *Configuration) GetInt(key string) int {
	return cast.ToInt(c.Get(key))
}

// GetInt32 returns the value associated with the key as an integer.
func (c *Configuration) GetInt32(key string) int32 {
	return cast.ToInt32(c.Get(key))
}

// GetInt64 returns the value associated with the key as an integer.
func (c *Configuration) GetInt64(key string) int64 {
	return cast.ToInt64(c.Get(key))
}

// GetUint returns the value associated with the key as an unsigned integer.
func (c *Configuration) GetUint(key string) uint {
	return cast.ToUint(c.Get(key))
}

// GetUint32 returns the value associated with the key as an unsigned integer.
func (c *Configuration) GetUint32(key string) uint32 {
	return cast.ToUint32(c.Get(key))
}

// GetUint64 returns the value associated with the key as an unsigned integer.
func (c *Configuration) GetUint64(key string) uint64 {
	return cast.ToUint64(c.Get(key))
}

// GetFloat64 returns the value associated with the key as a float64.
func (c *Configuration) GetFloat64(key string) float64 {
	return cast.ToFloat64(c.Get(key))
}

// GetTime returns the value associated with the key as time.
func (c *Configuration) GetTime(key string) time.Time {
	return cast.ToTime(c.Get(key))
}

// GetDuration returns the value associated with the key as a duration.
func (c *Configuration) GetDuration(key string) time.Duration {
	return cast.ToDuration(c.Get(key))
}

// GetIntSlice returns the value associated with the key as a slice of int values.
func (c *Configuration) GetIntSlice(key string) []int {
	return cast.ToIntSlice(c.Get(key))
}

// GetStringSlice returns the value associated with the key as a slice of strings.
func (c *Configuration) GetStringSlice(key string) []string {
	return cast.ToStringSlice(c.Get(key))
}

// GetStringMap returns the value associated with the key as a map of interfaces.
func (c *Configuration) GetStringMap(key string) map[string]interface{} {
	return cast.ToStringMap(c.Get(key))
}

// GetStringMapString returns the value associated with the key as a map of strings.
func (c *Configuration) GetStringMapString(key string) map[string]string {
	return cast.ToStringMapString(c.Get(key))
}

// GetConfigurationSlice returns the value associated with the key as a slice of configurations
func (c *Configuration) GetConfigurationSlice(key string) []*Configuration {
	if !c.IsSet(key) {
		return nil
	}
	v := c.Get(key)
	if v == nil {
		return nil
	}
	sliceV := reflect.ValueOf(v)
	if sliceV.Kind() != reflect.Slice {
		return nil
	}
	out := make([]*Configuration, 0, sliceV.Len())
	for i := 0; i < sliceV.Len(); i++ {
		sm := cast.ToStringMap(sliceV.Index(i))
		if len(sm) > 0 {
			out = append(out, NewConfigurationOfMap(sm))
		}
	}
	return out
}

const (
	DynamicTypeValue  = 0
	DynamicTypeConfig = 1
	DynamicTypeEnv    = 2
)

// 解析动态值：配置参数：${key:defaultV}，环境变量：#{key:defaultV}
func ParseDynamicKey(pattern string) (key string, def string, typ int) {
	pattern = strings.TrimSpace(pattern)
	size := len(pattern)
	if size <= 3 {
		return pattern, "", DynamicTypeValue
	}
	dyn := "${" == pattern[:2]
	env := "#{" == pattern[:2]
	if (dyn || env) && '}' == pattern[size-1] {
		values := strings.TrimSpace(pattern[2 : size-1])
		idx := strings.IndexByte(values, ':')
		key = values
		if idx > 0 {
			key = values[:idx]
			def = values[idx+1:]
		}
		if env {
			return key, def, DynamicTypeEnv
		} else {
			return key, def, DynamicTypeConfig
		}
	}
	return pattern, "", DynamicTypeValue
}
