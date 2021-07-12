package ext

import (
	"github.com/bytepowered/fluxgo/pkg/flux"
	"sort"
)

type filterw struct {
	filter flux.Filter
	order  int
}

type filters []filterw

func (a filters) Len() int           { return len(a) }
func (a filters) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a filters) Less(i, j int) bool { return a[i].order < a[j].order }

var (
	globalFilter    = make([]filterw, 0, 16)
	selectiveFilter = make([]filterw, 0, 16)
	filterSelectors = make([]flux.FilterSelector, 0, 8)
)

// AddGlobalFilter 注册全局Filter；
func AddGlobalFilter(v interface{}) {
	flux.AssertNotNil(v, "<filter> must not nil")
	globalFilter = _checkedAppendFilter(v, globalFilter)
	sort.Sort(filters(globalFilter))
}

// AddSelectiveFilter 注册可选Filter；
func AddSelectiveFilter(v interface{}) {
	flux.AssertNotNil(v, "<filter> must not nil")
	selectiveFilter = _checkedAppendFilter(v, selectiveFilter)
	sort.Sort(filters(selectiveFilter))
}

func _checkedAppendFilter(v interface{}, in []filterw) (out []filterw) {
	f := flux.MustNotNil(v, "Not a valid Filter").(flux.Filter)
	return append(in, filterw{filter: f, order: orderOfFilter(f)})
}

// SelectiveFilters 获取已排序的Filter列表
func SelectiveFilters() []flux.Filter {
	return getFilters(selectiveFilter)
}

// GlobalFilters 获取已排序的全局Filter列表
func GlobalFilters() []flux.Filter {
	return getFilters(globalFilter)
}

func AddFilterSelector(s flux.FilterSelector) {
	flux.MustNotNil(s, "<filter-selector> must not nil")
	filterSelectors = append(filterSelectors, s)
}

func FilterSelectors() []flux.FilterSelector {
	out := make([]flux.FilterSelector, len(filterSelectors))
	copy(out, filterSelectors)
	return out
}

// SelectiveFilterById 获取已排序的可选Filter列表
func SelectiveFilterById(filterId string) (flux.Filter, bool) {
	filterId = flux.MustNotEmpty(filterId, "<filter-id> must not empty")
	for _, f := range selectiveFilter {
		if filterId == f.filter.FilterId() {
			return f.filter, true
		}
	}
	return nil, false
}

func getFilters(in []filterw) []flux.Filter {
	out := make([]flux.Filter, len(in))
	for i, v := range in {
		out[i] = v.filter
	}
	return out
}

func orderOfFilter(v flux.Filter) int {
	if v, ok := v.(flux.Orderer); ok {
		return v.Order()
	} else {
		return 0
	}
}
