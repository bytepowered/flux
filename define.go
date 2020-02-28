package flux

// StringMap 定义一个KV字典
type StringMap map[string]interface{}

type (
	// Factory 用于动态初始化
	Factory func() interface{}
	// Startuper 用于介入服务启动生命周期的Hook，通常与 Orderer 接口一起使用。
	Startuper interface {
		Startup() error // 当服务启动时，调用此函数
	}
	// Shutdowner 用于介入服务停止生命周期的Hook，通常与 Orderer 接口一起使用。
	Shutdowner interface {
		Shutdown() error // 当服务停止时，调用此函数
	}
	// Initializer 用于介入服务停止生命周期的Hook，通常与 Orderer 接口一起使用。
	Initializer interface {
		Init(config Config) error // 当服务初始化时，调用此函数
	}
	// Orderer 用于定义顺序
	Orderer interface {
		Order() int // 返回排序顺序
	}
	// Valuer 用于定义值对象的读写接口
	Valuer interface {
		Value() interface{}
		SetValue(interface{})
	}
)
