package pkg

import "github.com/spf13/viper"

func NewConfigurationWith(namespace string) Configuration {
	return Configuration{namespace: namespace}
}

type Configuration struct {
	namespace string
}

func (c Configuration) Has(paths ...string) bool {
	for _, path := range paths {
		if _, ok := c.keyWithNamespace(path); !ok {
			return false
		}
	}
	return true
}

func (c Configuration) GetString(path string) string {
	key, _ := c.keyWithNamespace(path)
	return viper.GetString(key)
}

func (c Configuration) GetStringOr(path string, def string) string {
	if key, ok := c.keyWithNamespace(path); ok {
		return viper.GetString(key)
	} else {
		return def
	}
}

func (c Configuration) GetBoolOr(path string, def bool) bool {
	if key, ok := c.keyWithNamespace(path); ok {
		return viper.GetBool(key)
	} else {
		return def
	}
}

func (c Configuration) GetIntOr(path string, def int) int {
	if key, ok := c.keyWithNamespace(path); ok {
		return viper.GetInt(key)
	} else {
		return def
	}
}

func (c Configuration) GetInt64Or(path string, def int64) int64 {
	if key, ok := c.keyWithNamespace(path); ok {
		return viper.GetInt64(key)
	} else {
		return def
	}
}

func (c Configuration) keyWithNamespace(path string) (string, bool) {
	key := c.namespace + "." + path
	return key, viper.IsSet(key)
}
