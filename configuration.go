package flux

import (
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"time"
)

func NewConfigurationOf(namespace string) *Configuration {
	v := viper.Sub(namespace)
	if v == nil {
		v = viper.New()
	}
	return &Configuration{ref: v}
}

func NewConfiguration(in *viper.Viper) *Configuration {
	if nil == in {
		in = viper.New()
	}
	return &Configuration{ref: in}
}

type Configuration struct {
	ref         *viper.Viper
	globalAlias map[string]string
}

func (c *Configuration) Get(key string) interface{} {
	v := c.ref.Get(key)
	if nil == v && c.globalAlias != nil {
		if akey, ok := c.globalAlias[key]; ok {
			return viper.Get(akey)
		}
	}
	return v
}

func (c *Configuration) Set(key string, value interface{}) {
	c.ref.Set(key, value)
}

func (c *Configuration) SetGlobalAlias(alias map[string]string) {
	c.globalAlias = alias
}

func (c *Configuration) SetDefault(key string, value interface{}) {
	c.ref.SetDefault(key, value)
}

func (c *Configuration) SetDefaults(defaults map[string]interface{}) {
	for k, v := range defaults {
		c.ref.SetDefault(k, v)
	}
}

func (c *Configuration) IsSet(keys ...string) bool {
	for _, key := range keys {
		if !c.ref.IsSet(key) {
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
