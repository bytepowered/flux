package ext

import "github.com/bytepowered/flux"

var (
	_prepareHooks = make([]flux.PrepareHook, 0)
)

func AddPrepareHook(pf flux.PrepareHook) {
	_prepareHooks = append(_prepareHooks, pf)
}

func PrepareHooks() []flux.PrepareHook {
	d := make([]flux.PrepareHook, len(_prepareHooks))
	copy(d, _prepareHooks)
	return d
}
