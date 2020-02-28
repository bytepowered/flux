package flux

import "github.com/bytepowered/flux/pkg"

type Config map[string]interface{}

func ToConfig(values map[string]interface{}) Config {
	return Config(values)
}

func (c Config) IsEmpty() bool {
	return 0 == len(c)
}

func (c Config) String(name string) string {
	return c.StringOrDefault(name, "")
}

func (c Config) StringOrDefault(name string, defaultValue string) string {
	if v, ok := c[name]; ok {
		return pkg.ToString(v)
	} else {
		return defaultValue
	}
}

func (c Config) Int64(name string) int64 {
	return c.Int64OrDefault(name, int64(0))
}

func (c Config) Int64OrDefault(name string, defaultValue int64) int64 {
	if v, ok := c[name]; ok {
		if iv, e := pkg.ToInt64(v); nil != e {
			return defaultValue
		} else {
			return iv
		}
	} else {
		return defaultValue
	}
}

func (c Config) Float64(name string) float64 {
	return c.Float64OrDefault(name, float64(0))
}

func (c Config) Float64OrDefault(name string, defaultValue float64) float64 {
	if v, ok := c[name]; ok {
		if fv, e := pkg.ToFloat64(v); nil != e {
			return defaultValue
		} else {
			return fv
		}
	} else {
		return defaultValue
	}
}

func (c Config) Boolean(name string) bool {
	return c.BooleanOrDefault(name, false)
}

func (c Config) BooleanOrDefault(name string, defaultValue bool) bool {
	if v, ok := c[name]; ok {
		return pkg.ToBool(v)
	} else {
		return defaultValue
	}
}

func (c Config) Config(name string) Config {
	if v, ok := c[name]; ok {
		return ToConfig(v.(map[string]interface{}))
	} else {
		return make(Config, 0)
	}
}

func (c Config) Contains(name string) bool {
	_, ok := c[name]
	return ok
}
