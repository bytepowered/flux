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

// MakeConfigurationKey 根据Key列表，构建Configuration的查询Key。
// Note: Key列表任意单个Key不允许为空字符。
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

// NewConfigurationByKeys 根据命名空间和Key列表，构建Configuration实例
// Note：构建参数必须指定Namespace和Keys，否则报错Panic。
func NewConfigurationByKeys(namespaceAndKeys ...string) *Configuration {
	if len(namespaceAndKeys) == 0 {
		panic("namespace and keys is require")
	}
	return NewConfiguration(MakeConfigurationKey(namespaceAndKeys...))
}

// NewRootConfiguration 构建获取根配置对象。
func NewRootConfiguration() *Configuration {
	return newGlobalRefConfiguration("")
}

// NewConfiguration 根据指定Namespace的配置
func NewConfiguration(namespace string) *Configuration {
	if namespace == "" {
		panic("configuration: not allow empty namespace")
	}
	return newGlobalRefConfiguration(namespace)
}

func NewVarsConfiguration(vars map[string]interface{}) *Configuration {
	locals := viper.New()
	for k, v := range vars {
		locals.Set(k, v)
	}
	return newSpecifiedRefConfiguration("", locals, true)
}

// NewConfiguration 根据指定Namespace的配置
func newGlobalRefConfiguration(namespace string) *Configuration {
	// 持有Viper全局实例，通过Namespace来控制查询的Key
	return newSpecifiedRefConfiguration(namespace, viper.GetViper(), false)
}

func newSpecifiedRefConfiguration(namespace string, viperRef *viper.Viper, local bool) *Configuration {
	return &Configuration{
		namespace:  namespace,
		dataID:     namespace,
		root:       viperRef,
		isLocalRef: local,
		alias:      make(map[string]string),
	}
}

// Configuration 封装Viper实例访问接口的配置类
// 根据Namespace指向不同的配置路径，可以从全局配置中读取指定域的配置数据
type Configuration struct {
	dataID     string            // 数据ID
	namespace  string            // 配置所属命名空间
	root       *viper.Viper      // 实际的配置实例
	isLocalRef bool              // 是否使用本地Viper实例
	alias      map[string]string // 本地Key别名
	watchStop  chan struct{}
}

// SetDataId 设置当前配置实例的 dataId
func (c *Configuration) SetDataId(dataID string) {
	c.dataID = dataID
}

// DataId 返回当前配置实例的 dataId
func (c *Configuration) DataId() string {
	return c.dataID
}

// ToStringMap 将当前配置实例（命名空间）下所有配置，转换成 map[string]any 类型的字典。
func (c *Configuration) ToStringMap() map[string]interface{} {
	if "" == c.namespace {
		return c.root.AllSettings()
	}
	return cast.ToStringMap(c.root.Get(c.namespace))
}

// Keys 获取当前配置实例（命名空间）下所有配置的键列表
func (c *Configuration) Keys() []string {
	v := c.root.Sub(c.namespace)
	if v != nil {
		return v.AllKeys()
	}
	return []string{}
}

// ToConfigurations 将当前配置实例（命名空间）下所有配置，转换成 Configuration 类型的列表。
func (c *Configuration) ToConfigurations() []*Configuration {
	if "" == c.namespace {
		return ToConfigurations("", []interface{}{c.root.AllSettings()})
	}
	return ToConfigurations(c.namespace, c.root.Get(c.namespace))
}

func (c *Configuration) Sub(subNamespace string) *Configuration {
	return newSpecifiedRefConfiguration(c.makeKey(subNamespace), c.root, false)
}

func (c *Configuration) Get(key string) interface{} {
	return c.doget(c.makeKey(key), nil)
}

func (c *Configuration) GetOrDefault(key string, def interface{}) interface{} {
	return c.doget(c.makeKey(key), def)
}

// Set 向当前配置实例以覆盖的方式设置Key-Value键值。
func (c *Configuration) Set(key string, value interface{}) {
	c.root.Set(c.makeKey(key), value)
}

// SetKeyAlias 设置当前配置实例的Key与GlobalAlias的映射
func (c *Configuration) SetKeyAlias(keyAlias map[string]string) {
	for key, alias := range keyAlias {
		c.alias[c.makeKey(key)] = alias
	}
}

// SetDefault 为当前配置实例设置单个默认值。与Viper的SetDefault一致，作用于当前配置实例。
func (c *Configuration) SetDefault(key string, value interface{}) {
	c.root.SetDefault(c.makeKey(key), value)
}

// SetDefaults 为当前配置实例设置一组默认值。与Viper的SetDefault一致，作用于当前配置实例。
func (c *Configuration) SetDefaults(defaults map[string]interface{}) {
	for key, val := range defaults {
		c.root.SetDefault(c.makeKey(key), val)
	}
}

// IsSet 判定当前配置实例是否设置指定Key（多个）。与Viper的IsSet一致，查询范围为当前配置实例。
func (c *Configuration) IsSet(keys ...string) bool {
	if len(keys) == 0 {
		return false
	}
	// Any not set, return false
	for _, key := range keys {
		if !c.root.IsSet(c.makeKey(key)) {
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
	if !c.root.IsSet(key) {
		return nil
	}
	v := c.root.Get(key)
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
	if !c.root.IsSet(key) {
		return nil
	}
	return c.root.UnmarshalKey(key, outptr, func(opt *mapstructure.DecoderConfig) {
		opt.TagName = structTag
	})
}

func (c *Configuration) StartWatch(notify func(key string, value interface{})) {
	if c.watchStop != nil {
		return
	}
	c.watchStop = make(chan struct{}, 1)
	values := make(map[string]interface{}, 16)
	for _, k := range c.Keys() {
		values[k] = c.Get(k)
	}
	go func() {
		watch := time.NewTicker(time.Second)
		defer func() {
			watch.Stop()
		}()
		c.watchStop = nil
		for {
			select {
			case <-watch.C:
				for _, key := range c.Keys() {
					newV := c.Get(key)
					if oldV := values[key]; !reflect.DeepEqual(oldV, newV) {
						values[key] = newV
						notify(key, newV)
					}
				}

			case <-c.watchStop:
				return
			}
		}
	}()
}

func (c *Configuration) StopWatch() {
	if c.watchStop == nil {
		return
	}
	select {
	case c.watchStop <- struct{}{}:
	default:
		return
	}
}

func (c *Configuration) makeKey(key string) string {
	if c.isLocalRef || c.namespace == "" {
		return key
	}
	return MakeConfigurationKey(c.namespace, key)
}

func (c *Configuration) doget(key string, indef interface{}) interface{} {
	val := c.root.Get(key)
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
		case DynamicTypeLookupConfig:
			// check circle key
			if key == pkey {
				return usedef
			}
			if c.root.IsSet(pkey) {
				return c.doget(pkey, usedef)
			} else {
				return usedef
			}

		case DynamicTypeLookupEnv:
			if ev, ok := os.LookupEnv(pkey); ok {
				return ev
			} else {
				return usedef
			}

		case DynamicTypeStaticValue:
			return val

		default:
			return val
		}
	}
	// check local alias
	if nil == val {
		if alias, ok := c.alias[key]; ok {
			val = c.root.Get(alias)
		}
	}
	if nil == val {
		return indef
	}
	return val
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
				namespace:  namespace + fmt.Sprintf("[%d]", i),
				isLocalRef: true,
				root: func() *viper.Viper {
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
	DynamicTypeStaticValue  = 0 << iota
	DynamicTypeLookupConfig = 1
	DynamicTypeLookupEnv    = 2
)

// ParseDynamicKey 解析动态值：配置参数：${key:defaultV}，环境变量：#{key:defaultV}
func ParseDynamicKey(pattern string) (key string, def string, typ int) {
	pattern = strings.TrimSpace(pattern)
	size := len(pattern)
	if size <= len("${}") {
		return pattern, "", DynamicTypeStaticValue
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
			return key, def, DynamicTypeLookupEnv
		} else {
			return key, def, DynamicTypeLookupConfig
		}
	}
	return pattern, "", DynamicTypeStaticValue
}
