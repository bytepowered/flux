package flux

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"os"
	"reflect"
	"strings"
	"time"
)

const (
	NamespaceWebListeners = "listeners"
	NamespaceTransporters = "transporters"
	NamespaceDiscoveries  = "discoveries"
)

func MakeConfigurationKey(keys ...string) string {
	if len(keys) == 0 {
		return ""
	}
	for _, key := range keys {
		if key == "" {
			panic("configuration: not allow empty key")
		}
	}
	return strings.Join(keys, ".")
}

// NewConfiguration 根据指定Namespace的配置
func NewConfiguration(namespace string) *Configuration {
	if namespace == "" {
		panic("configuration: not allow empty namespace")
	}
	return &Configuration{
		nspath:   namespace,
		dataID:   namespace,
		registry: viper.GetViper(), // 持有Viper全局实例，通过Namespace来控制查询的Key
		local:    false,
	}
}

// Configuration 封装Viper实例访问接口的配置类
type Configuration struct {
	dataID   string       // 数据ID
	nspath   string       // 配置所属命名空间
	registry *viper.Viper // 实际的配置实例
	local    bool         // 是否使用本地Viper实例
}

func (c *Configuration) makeKey(key string) string {
	if c.local || c.nspath == "" {
		return key
	}
	return MakeConfigurationKey(c.nspath, key)
}

func (c *Configuration) SetDataId(dataID string) {
	c.dataID = dataID
}

func (c *Configuration) DataId() string {
	return c.dataID
}

func (c *Configuration) ToStringMap() map[string]interface{} {
	return cast.ToStringMap(c.registry.Get(c.nspath))
}

func (c *Configuration) Keys() []string {
	return c.registry.Sub(c.nspath).AllKeys()
}

func (c *Configuration) ToConfigurations() []*Configuration {
	return ToConfigurations(c.nspath, c.registry.Get(c.nspath))
}

func (c *Configuration) Sub(subNamespace string) *Configuration {
	return NewConfiguration(c.makeKey(subNamespace))
}

func (c *Configuration) Get(key string) interface{} {
	return c.doget(c.makeKey(key), nil)
}

func (c *Configuration) GetOrDefault(key string, def interface{}) interface{} {
	return c.doget(c.makeKey(key), def)
}

func (c *Configuration) doget(key string, indef interface{}) interface{} {
	val := c.registry.Get(key)
	if expr, ok := val.(string); ok {
		// 动态全局Key和默认值： ${username:yongjia}
		pkey, pdef, ptype := ParseDynamicKey(expr)
		var usedef interface{}
		if indef != nil {
			usedef = indef
		} else {
			usedef = pdef
		}
		switch ptype {
		case DynamicTypeConfig:
			if key == pkey {
				return usedef
			}
			if c.registry.IsSet(pkey) {
				return c.doget(pkey, usedef)
			} else {
				return usedef
			}

		case DynamicTypeEnv:
			if ev, ok := os.LookupEnv(pkey); ok {
				return ev
			} else {
				return usedef
			}

		case DynamicTypeValue:
			return val

		default:
			return val
		}
	}
	if nil == val {
		return indef
	}
	return val
}

// Set 向当前配置实例以覆盖的方式设置Key-Value键值。
func (c *Configuration) Set(key string, value interface{}) {
	c.registry.Set(c.makeKey(key), value)
}

// SetKeyAlias 设置当前配置实例的Key与GlobalAlias的映射
func (c *Configuration) SetKeyAlias(keyAlias map[string]string) {
	for key, alias := range keyAlias {
		c.registry.RegisterAlias(alias, c.makeKey(key))
	}
}

// SetDefault 为当前配置实例设置单个默认值。与Viper的SetDefault一致，作用于当前配置实例。
func (c *Configuration) SetDefault(key string, value interface{}) {
	c.registry.SetDefault(c.makeKey(key), value)
}

// SetDefault 为当前配置实例设置一组默认值。与Viper的SetDefault一致，作用于当前配置实例。
func (c *Configuration) SetDefaults(defaults map[string]interface{}) {
	for key, val := range defaults {
		c.registry.SetDefault(c.makeKey(key), val)
	}
}

// IsSet 判定当前配置实例是否设置指定Key（多个）。与Viper的IsSet一致，查询范围为当前配置实例。
func (c *Configuration) IsSet(keys ...string) bool {
	if len(keys) == 0 {
		return false
	}
	// Any not set, return false
	for _, key := range keys {
		if !c.registry.IsSet(c.makeKey(key)) {
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

// GetConfigurations returns the value associated with the key as a slice of configurations
func (c *Configuration) GetConfigurations(key string) []*Configuration {
	key = c.makeKey(key)
	if !c.registry.IsSet(key) {
		return nil
	}
	v := c.registry.Get(key)
	if v == nil {
		return nil
	}
	return ToConfigurations(key, v)
}

func (c *Configuration) GetStruct(key string, outptr interface{}) error {
	return c.GetStructTag(key, "json", outptr)
}

func (c *Configuration) GetStructTag(key, structTag string, outptr interface{}) error {
	key = c.makeKey(key)
	if !c.registry.IsSet(key) {
		return nil
	}
	return c.registry.UnmarshalKey(key, outptr, func(opt *mapstructure.DecoderConfig) {
		opt.TagName = structTag
	})
}

func ToConfigurations(namespace string, v interface{}) []*Configuration {
	sliceV := reflect.ValueOf(v)
	if sliceV.Kind() != reflect.Slice {
		return nil
	}
	out := make([]*Configuration, 0, sliceV.Len())
	for i := 0; i < sliceV.Len(); i++ {
		sm := cast.ToStringMap(sliceV.Index(i).Interface())
		if len(sm) > 0 {
			out = append(out, &Configuration{
				nspath: namespace + fmt.Sprintf("[%d]", i),
				local:  true,
				registry: func() *viper.Viper {
					r := viper.New()
					for k, v := range sm {
						r.Set(k, v)
					}
					return r
				}(),
			})
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
