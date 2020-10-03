package ext

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/pkg"
	"sort"
)

type filterWrapper struct {
	filter flux.Filter
	order  int
}

type filterArray []filterWrapper

func (s filterArray) Len() int           { return len(s) }
func (s filterArray) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s filterArray) Less(i, j int) bool { return s[i].order < s[j].order }

var (
	_globalFilter    = make([]filterWrapper, 0, 16)
	_selectiveFilter = make([]filterWrapper, 0, 16)
)

// AddGlobalFilter 注册全局Filter；
func AddGlobalFilter(v interface{}) {
	_globalFilter = _checkedAppendFilter(v, _globalFilter)
	sort.Sort(filterArray(_globalFilter))
}

// AddSelectiveFilter 注册可选Filter；
func AddSelectiveFilter(v interface{}) {
	_selectiveFilter = _checkedAppendFilter(v, _selectiveFilter)
	sort.Sort(filterArray(_selectiveFilter))
}

func _checkedAppendFilter(v interface{}, in []filterWrapper) (out []filterWrapper) {
	f := pkg.RequireNotNil(v, "Not a filter").(flux.Filter)
	return append(in, filterWrapper{filter: f, order: orderOf(v)})
}

// SelectiveFilters 获取已排序的Filter列表
func SelectiveFilters() []flux.Filter {
	return _getFilters(_selectiveFilter)
}

// GlobalFilters 获取已排序的全局Filter列表
func GlobalFilters() []flux.Filter {
	return _getFilters(_globalFilter)
}

func _getFilters(in []filterWrapper) []flux.Filter {
	out := make([]flux.Filter, len(in))
	for i, v := range in {
		out[i] = v.filter
	}
	return out
}

// GetSelectiveFilter 获取已排序的可选Filter列表
func GetSelectiveFilter(filterId string) (flux.Filter, bool) {
	for _, f := range _selectiveFilter {
		if filterId == f.filter.TypeId() {
			return f.filter, true
		}
	}
	return nil, false
}

func orderOf(v interface{}) int {
	if v, ok := v.(flux.Orderer); ok {
		return v.Order()
	} else {
		return 0
	}
}
