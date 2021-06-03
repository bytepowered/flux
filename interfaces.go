package flux

import "context"

// Build version info
type Build struct {
	CommitId string
	Version  string
	Date     string
}

type (
	// Factory 工厂函数，用于动态初始化某些组件实例。
	Factory func() interface{}

	// Preparable 用于介入服务预备阶段的生命周期的Hook。
	Preparable interface {
		OnPrepare() error
	}

	// Startuper 用于介入服务启动生命周期的Hook，通常与 Orderer 接口一起使用。
	Startuper interface {
		// OnStartup 当服务启动时，调用此函数
		OnStartup() error
	}

	// Shutdowner 用于介入服务停止生命周期的Hook，通常与 Orderer 接口一起使用。
	Shutdowner interface {
		// OnShutdown 当服务停止时，调用此函数
		OnShutdown(ctx context.Context) error
	}

	// Initializer 用于介入服务停止生命周期的Hook。
	Initializer interface {
		// OnInit 当服务初始化时，调用此函数
		OnInit(configuration *Configuration) error
	}

	// Orderer 用于定义顺序
	Orderer interface {
		// Order 返回排序顺序
		Order() int
	}
)

type prepareablew func() error

func (wf prepareablew) OnPrepare() error {
	return wf()
}

func WrapPreparable(f func() error) Preparable {
	return prepareablew(f)
}

type startupw func() error

func (wf startupw) OnStartup() error {
	return wf()
}

func WrapStartuper(f func() error) Startuper {
	return startupw(f)
}

type shutdownw func(context.Context) error

func (wf shutdownw) OnShutdown(ctx context.Context) error {
	return wf(ctx)
}

func WrapShutdown(f func(context.Context) error) Shutdowner {
	return shutdownw(f)
}

// LoggerFactory 构建Logger实例
type LoggerFactory func(values context.Context) Logger

// Logger 日志Logger接口定义
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
