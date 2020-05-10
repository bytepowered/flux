package flux

import "github.com/spf13/viper"

func NewConfigurationOf(namespace string) Configuration {
	v := viper.Sub(namespace)
	if v != nil {
		v = viper.New()
	}
	return Configuration{Viper: v}
}

func NewConfiguration(in *viper.Viper) Configuration {
	if nil == in {
		in = viper.New()
	}
	return Configuration{Viper: in}
}

type Configuration struct {
	*viper.Viper
}

func (c Configuration) IsSetKeys(keys ...string) bool {
	for _, key := range keys {
		if !c.IsSet(key) {
			return false
		}
	}
	return true
}

func (c Configuration) GetStringDefault(key string, def string) string {
	c.setDefaultIfAbsent(key, def)
	return c.GetString(key)
}

func (c Configuration) GetBoolDefault(key string, def bool) bool {
	c.setDefaultIfAbsent(key, def)
	return c.GetBool(key)
}

func (c Configuration) GetIntDefault(key string, def int) int {
	c.setDefaultIfAbsent(key, def)
	return c.GetInt(key)
}

func (c Configuration) GetInt64Default(key string, def int64) int64 {
	c.setDefaultIfAbsent(key, def)
	return c.GetInt64(key)
}

func (c Configuration) setDefaultIfAbsent(key string, def interface{}) {
	if !c.IsSet(key) {
		c.SetDefault(key, def)
	}
}
