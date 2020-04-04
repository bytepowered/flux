package ext

import (
	"github.com/bytepowered/flux"
	"reflect"
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
	_globalFilter = make([]filterWrapper, 0)
	_scopedFilter = make([]filterWrapper, 0)
)

// AddGlobalFilter 注册全局Filter；
func AddGlobalFilter(v interface{}) {
	f, ok := v.(flux.Filter)
	if !ok || nil == f {
		GetLogger().Panicf("Not a Filter: type=%s, v=%+v", reflect.TypeOf(v), v)
	}
	_globalFilter = append(_globalFilter, filterWrapper{filter: f, order: orderOf(v)})
	sort.Sort(filterArray(_globalFilter))
}

// AddGlobalFilter 注册全局Filter；
func AddFilter(v interface{}) {
	f, ok := v.(flux.Filter)
	if !ok || nil == f {
		GetLogger().Panicf("Not a Filter: type=%s, v=%+v", reflect.TypeOf(v), v)
	}
	_scopedFilter = append(_scopedFilter, filterWrapper{filter: f, order: orderOf(v)})
	sort.Sort(filterArray(_scopedFilter))
}

// GlobalFilters 获取已排序的全局Filter列表
func GlobalFilters() []flux.Filter {
	out := make([]flux.Filter, len(_globalFilter))
	for i, v := range _globalFilter {
		out[i] = v.filter
	}
	return out
}

// GlobalFilters 获取已排序的全局Filter列表
func GetFilter(filterId string) (flux.Filter, bool) {
	for _, f := range _scopedFilter {
		if filterId == f.filter.Id() {
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
