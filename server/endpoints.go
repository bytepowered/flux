package server

import (
	"github.com/bytepowered/flux/internal"
	"github.com/labstack/echo/v4"
	"strings"
)

const (
	_typeApplication = "application"
	_typeProtocol    = "protocol"
	_typeHttpPattern = "http-pattern"
	_typeUpstreamUri = "upstream-uri"
)

type _filter func(ep *internal.MultiVersionEndpoint) bool

// 支持以下过滤条件
var _typeKeys = []string{"application", "protocol", "http-pattern", "upstream-uri"}

var (
	_filterFactories = make(map[string]func(string) _filter)
)

func init() {
	_filterFactories[_typeApplication] = func(query string) _filter {
		return func(ep *internal.MultiVersionEndpoint) bool {
			return query == ep.RandomVersion().Application
		}
	}
	_filterFactories[_typeProtocol] = func(query string) _filter {
		return func(ep *internal.MultiVersionEndpoint) bool {
			return strings.ToLower(query) == strings.ToLower(ep.RandomVersion().Protocol)
		}
	}
	_filterFactories[_typeHttpPattern] = func(query string) _filter {
		return func(ep *internal.MultiVersionEndpoint) bool {
			return query == ep.RandomVersion().HttpPattern
		}
	}
	_filterFactories[_typeUpstreamUri] = func(query string) _filter {
		return func(ep *internal.MultiVersionEndpoint) bool {
			return query == ep.RandomVersion().UpstreamUri
		}
	}
}

func queryEndpoints(data map[string]*internal.MultiVersionEndpoint, request echo.Context) interface{} {
	filters := make([]_filter, 0)
	for _, key := range _typeKeys {
		if query := request.QueryParam(key); "" != query {
			if f, ok := _filterFactories[key]; ok {
				filters = append(filters, f(query))
			}
		}
	}
	if len(filters) == 0 {
		m := make(map[string]interface{})
		for k, v := range data {
			m[k] = v.ToSerializableMap()
		}
		return m
	}
	return _queryWithFilters(data, filters...)
}

func _queryWithFilters(data map[string]*internal.MultiVersionEndpoint, filters ..._filter) []interface{} {
	items := make([]interface{}, 0)
DataLoop:
	for _, v := range data {
		for _, filter := range filters {
			// 任意Filter返回True
			if filter(v) {
				items = append(items, v.ToSerializableMap())
				continue DataLoop
			}
		}
	}
	return items
}
