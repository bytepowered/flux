package extension

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
)

// SetGlobalFilter 注册全局Filter；此函数会自动注册生命周期Hook
func SetGlobalFilter(v interface{}) {
	m, ok := v.(flux.Filter)
	if !ok {
		GetLogger().Panicf("Not a Filter: type=%s, v=%+v", reflect.TypeOf(v), v)
	}
	_globalFilter = append(_globalFilter, filterWrapper{
		filter: m,
		order:  orderOf(v),
	})
	sort.Sort(filterArray(_globalFilter))
}

// GetGlobalFilter 获取已排序的全局Filter列表
func GetGlobalFilter() []flux.Filter {
	out := make([]flux.Filter, len(_globalFilter))
	for i, v := range _globalFilter {
		out[i] = v.filter
	}
	return out
}

func orderOf(v interface{}) int {
	if v, ok := v.(flux.Orderer); ok {
		return v.Order()
	} else {
		return 0
	}
}
