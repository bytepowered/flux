package ext

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/pkg"
)

var (
	filterSelectors = make([]flux.FilterSelector, 0, 8)
)

func AddFilterSelector(s flux.FilterSelector) {
	pkg.RequireNotNil(s, "FilterSelector is nil")
	filterSelectors = append(filterSelectors, s)
}

func GetFilterSelectors() []flux.FilterSelector {
	out := make([]flux.FilterSelector, len(filterSelectors))
	copy(out, filterSelectors)
	return out
}
