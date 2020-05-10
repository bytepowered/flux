package ext

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/pkg"
)

var (
	_prepareHooks = make([]flux.PrepareHook, 0)
)

func AddPrepareHook(pf flux.PrepareHook) {
	_prepareHooks = append(_prepareHooks, pkg.RequireNotNil(pf, "PrepareHook is nil").(flux.PrepareHook))
}

func PrepareHooks() []flux.PrepareHook {
	d := make([]flux.PrepareHook, len(_prepareHooks))
	copy(d, _prepareHooks)
	return d
}
