package pkg

import "github.com/spf13/viper"

func NewConfigurationWith(namespace string) Configuration {
	return Configuration{namespace: namespace}
}

type Configuration struct {
	namespace string
}

func (c Configuration) Has(keys ...string) bool {
	for _, key := range keys {
		if _, ok := c.checkedKey(key); !ok {
			return false
		}
	}
	return true
}

func (c Configuration) GetString(key string) string {
	return viper.GetString(c.toPath(key))
}

func (c Configuration) GetStringOr(key string, def string) string {
	if key, ok := c.checkedKey(key); ok {
		return viper.GetString(key)
	} else {
		return def
	}
}

func (c Configuration) GetBoolOr(key string, def bool) bool {
	if key, ok := c.checkedKey(key); ok {
		return viper.GetBool(key)
	} else {
		return def
	}
}

func (c Configuration) GetIntOr(key string, def int) int {
	if key, ok := c.checkedKey(key); ok {
		return viper.GetInt(key)
	} else {
		return def
	}
}

func (c Configuration) GetInt64Or(key string, def int64) int64 {
	if key, ok := c.checkedKey(key); ok {
		return viper.GetInt64(key)
	} else {
		return def
	}
}

func (c Configuration) checkedKey(key string) (string, bool) {
	path := c.toPath(key)
	return path, viper.IsSet(path)
}

func (c Configuration) toPath(key string) string {
	return c.namespace + "." + key
}
