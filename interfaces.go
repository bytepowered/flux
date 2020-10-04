package flux

import "context"

// Build version info
type BuildInfo struct {
	CommitId string
	Version  string
	Date     string
}

type (
	// Factory 用于动态初始化
	Factory func() interface{}
	// PrepareHookFunc 在初始化调用前的预备函数
	PrepareHookFunc func() error
	// Startuper 用于介入服务启动生命周期的Hook，通常与 Orderer 接口一起使用。
	Startuper interface {
		Startup() error // 当服务启动时，调用此函数
	}
	// Shutdowner 用于介入服务停止生命周期的Hook，通常与 Orderer 接口一起使用。
	Shutdowner interface {
		Shutdown(ctx context.Context) error // 当服务停止时，调用此函数
	}
	// Initializer 用于介入服务停止生命周期的Hook，通常与 Orderer 接口一起使用。
	Initializer interface {
		Init(configuration *Configuration) error // 当服务初始化时，调用此函数
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

// 日志Logger接口定义
type Logger interface {
	// uses fmt.Sprint to construct and log a message.
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Debug(args ...interface{})
	Panic(args ...interface{})
	// uses fmt.Sprintf to log a templated message.
	Infof(fmt string, args ...interface{})
	Warnf(fmt string, args ...interface{})
	Errorf(fmt string, args ...interface{})
	Debugf(fmt string, args ...interface{})
	Panicf(fmt string, args ...interface{})
	// logs a message with some additional context. The variadic key-value
	// pairs are treated as they are in With.
	Infow(msg string, keyAndValues ...interface{})
	Warnw(msg string, keyAndValues ...interface{})
	Errorw(msg string, keyAndValues ...interface{})
	Debugw(msg string, keyAndValues ...interface{})
	Panicw(msg string, keyAndValues ...interface{})
}

type LoggerFactory func(values context.Context) Logger
