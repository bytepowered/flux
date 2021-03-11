package ext

import (
	flux2 "github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-pkg"
)

var (
	hooksPrepare  = make([]flux2.PrepareHookFunc, 0, 16)
	hooksStartup  = make([]flux2.Startuper, 0, 16)
	hooksShutdown = make([]flux2.Shutdowner, 0, 16)
)

// AddHookFunc 添加生命周期启动与停止的钩子接口
func AddHookFunc(hook interface{}) {
	fluxpkg.MustNotNil(hook, "Hook is nil")
	if startup, ok := hook.(flux2.Startuper); ok {
		hooksStartup = append(hooksStartup, startup)
	}
	if shutdown, ok := hook.(flux2.Shutdowner); ok {
		hooksShutdown = append(hooksShutdown, shutdown)
	}
}

// AddPrepareHook 添加预备阶段钩子函数
func AddPrepareHook(pf flux2.PrepareHookFunc) {
	hooksPrepare = append(hooksPrepare, fluxpkg.MustNotNil(pf, "PrepareHookFunc is nil").(flux2.PrepareHookFunc))
}

func PrepareHooks() []flux2.PrepareHookFunc {
	dst := make([]flux2.PrepareHookFunc, len(hooksPrepare))
	copy(dst, hooksPrepare)
	return dst
}

func StartupHooks() []flux2.Startuper {
	dst := make([]flux2.Startuper, len(hooksStartup))
	copy(dst, hooksStartup)
	return dst
}

func ShutdownHooks() []flux2.Shutdowner {
	dst := make([]flux2.Shutdowner, len(hooksShutdown))
	copy(dst, hooksShutdown)
	return dst
}
