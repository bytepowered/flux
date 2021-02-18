package ext

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/pkg"
)

var (
	selectors = make([]flux.Selector, 0, 8)
)

func AddSelector(s flux.Selector) {
	pkg.RequireNotNil(s, "Selector is nil")
	selectors = append(selectors, s)
}

func GetSelectors() []flux.Selector {
	out := make([]flux.Selector, len(selectors))
	copy(out, selectors)
	return out
}
