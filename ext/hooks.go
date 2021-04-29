package ext

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/fluxkit"
)

var (
	hooksPrepare  = make([]flux.PrepareHookFunc, 0, 16)
	hooksStartup  = make([]flux.Startuper, 0, 16)
	hooksShutdown = make([]flux.Shutdowner, 0, 16)
)

// AddHookFunc 添加生命周期启动与停止的钩子接口
func AddHookFunc(hook interface{}) {
	fluxkit.MustNotNil(hook, "Hook is nil")
	if startup, ok := hook.(flux.Startuper); ok {
		hooksStartup = append(hooksStartup, startup)
	}
	if shutdown, ok := hook.(flux.Shutdowner); ok {
		hooksShutdown = append(hooksShutdown, shutdown)
	}
}

// AddPrepareHook 添加预备阶段钩子函数
func AddPrepareHook(pf flux.PrepareHookFunc) {
	hooksPrepare = append(hooksPrepare, fluxkit.MustNotNil(pf, "PrepareHookFunc is nil").(flux.PrepareHookFunc))
}

func PrepareHooks() []flux.PrepareHookFunc {
	dst := make([]flux.PrepareHookFunc, len(hooksPrepare))
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
