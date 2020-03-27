package flux

type ConfigFactory func(namespace string, data map[string]interface{}) Config

type Config interface {
	IsEmpty() bool
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
