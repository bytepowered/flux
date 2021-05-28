package ext

import (
	"github.com/bytepowered/flux"
)

var (
	hooksPrepare  = make([]flux.Preparable, 0, 16)
	hooksStartup  = make([]flux.Startuper, 0, 16)
	hooksShutdown = make([]flux.Shutdowner, 0, 16)
)

// AddPrepareHook 添加预备阶段钩子函数
func AddPrepareHook(hook interface{}) {
	flux.MustNotNil(hook, "Prepare hook is nil")
	if prepare, ok := hook.(flux.Preparable); ok {
		hooksPrepare = append(hooksPrepare, prepare)
	}
}

// AddStartuper 添加生命周期启动钩子接口
func AddShutdownHook(hook interface{}) {
	flux.MustNotNil(hook, "Shutdown hook is nil")
	if shutdown, ok := hook.(flux.Shutdowner); ok {
		hooksShutdown = append(hooksShutdown, shutdown)
	}
}

// AddStartuper 添加生命周期停止的钩子接口
func AddStartupHook(hook interface{}) {
	flux.MustNotNil(hook, "Startup hook is nil")
	if startup, ok := hook.(flux.Startuper); ok {
		hooksStartup = append(hooksStartup, startup)
	}
}

func PrepareHooks() []flux.Preparable {
	dst := make([]flux.Preparable, len(hooksPrepare))
	copy(dst, hooksPrepare)
	return dst
}

func StartupHooks() []flux.Startuper {
	dst := make([]flux.Startuper, len(hooksStartup))
	copy(dst, hooksStartup)
	return dst
}

func ShutdownHooks() []flux.Shutdowner {
	dst := make([]flux.Shutdowner, len(hooksShutdown))
	copy(dst, hooksShutdown)
	return dst
}
