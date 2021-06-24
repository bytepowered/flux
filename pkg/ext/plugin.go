package ext

import (
	"github.com/bytepowered/fluxgo/pkg/flux"
	"sort"
)

type pluginw struct {
	plugin flux.Plugin
	order  int
}

type plugins []pluginw

func (a plugins) Len() int           { return len(a) }
func (a plugins) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a plugins) Less(i, j int) bool { return a[i].order < a[j].order }

var (
	globalPlugin    = make([]pluginw, 0, 16)
	selectivePlugin = make([]pluginw, 0, 16)
	pluginSelectors = make([]flux.PluginSelector, 0, 8)
)

// AddGlobalPlugin 注册全局Plugin；
func AddGlobalPlugin(v interface{}) {
	flux.AssertNotNil(v, "<plugin> must not nil")
	globalPlugin = _checkedAppendPlugin(v, globalPlugin)
	sort.Sort(plugins(globalPlugin))
}

// AddSelectivePlugin 注册可选Plugin；
func AddSelectivePlugin(v interface{}) {
	flux.AssertNotNil(v, "<plugin> must not nil")
	selectivePlugin = _checkedAppendPlugin(v, selectivePlugin)
	sort.Sort(plugins(selectivePlugin))
}

func _checkedAppendPlugin(v interface{}, in []pluginw) (out []pluginw) {
	flux.AssertNotNil(v, "<plugin> must not nil")
	p := v.(flux.Plugin)
	return append(in, pluginw{plugin: p, order: orderOfPlugin(p)})
}

// SelectivePlugins 获取已排序的Plugin列表
func SelectivePlugins() []flux.Plugin {
	return getPlugins(selectivePlugin)
}

// GlobalPlugins 获取已排序的全局Plugin列表
func GlobalPlugins() []flux.Plugin {
	return getPlugins(globalPlugin)
}

func AddPluginSelector(s flux.PluginSelector) {
	flux.MustNotNil(s, "PluginSelector is nil")
	pluginSelectors = append(pluginSelectors, s)
}

func PluginSelectors() []flux.PluginSelector {
	out := make([]flux.PluginSelector, len(pluginSelectors))
	copy(out, pluginSelectors)
	return out
}

// SelectivePluginById 获取已排序的可选Plugin列表
func SelectivePluginById(pluginId string) (flux.Plugin, bool) {
	pluginId = flux.MustNotEmpty(pluginId, "pluginId is empty")
	for _, f := range selectivePlugin {
		if pluginId == f.plugin.PluginId() {
			return f.plugin, true
		}
	}
	return nil, false
}

func getPlugins(in []pluginw) []flux.Plugin {
	out := make([]flux.Plugin, len(in))
	for i, v := range in {
		out[i] = v.plugin
	}
	return out
}

func orderOfPlugin(v flux.Plugin) int {
	if v, ok := v.(flux.Orderer); ok {
		return v.Order()
	} else {
		return 0
	}
}
