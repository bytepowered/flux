package server

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"net/http"
	"strings"
)

const (
	_typeApplication  = "application"
	_typeProtocol     = "protocol"
	_typeHttpPattern  = "http-pattern"
	_typeHttpPattern0 = "httpPattern"
	_typeHttpPattern1 = "httppattern"
	_typeInterface    = "interface"
)

type _filter func(ep *BindEndpoint) bool

// 支持以下过滤条件
var _typeKeys = []string{_typeApplication, _typeProtocol,
	_typeHttpPattern, _typeHttpPattern0, _typeHttpPattern1,
	_typeInterface,
}

var (
	_filterFactories = make(map[string]func(string) _filter)
)

func DebugQueryEndpoint(datamap map[string]*BindEndpoint) http.HandlerFunc {
	// Endpoint查询
	serializer := ext.LoadSerializer(ext.TypeNameSerializerJson)
	return func(writer http.ResponseWriter, request *http.Request) {
		data := _queryEndpoints(datamap, request)
		if data, err := serializer.Marshal(data); nil != err {
			writer.WriteHeader(http.StatusInternalServerError)
			_, _ = writer.Write([]byte(err.Error()))
		} else {
			writer.WriteHeader(http.StatusOK)
			_, _ = writer.Write(data)
		}
	}
}

func init() {
	_filterFactories[_typeApplication] = func(query string) _filter {
		return func(ep *BindEndpoint) bool {
			return _queryMatch(query, ep.RandomVersion().Application)
		}
	}
	_filterFactories[_typeProtocol] = func(query string) _filter {
		return func(ep *BindEndpoint) bool {
			proto := ep.RandomVersion().Service.RpcProto
			return _queryMatch(query, proto)
		}
	}
	httpPatternFilter := func(query string) _filter {
		return func(ep *BindEndpoint) bool {
			return _queryMatch(query, ep.RandomVersion().HttpPattern)
		}
	}
	_filterFactories[_typeHttpPattern] = httpPatternFilter
	_filterFactories[_typeHttpPattern0] = httpPatternFilter
	_filterFactories[_typeHttpPattern1] = httpPatternFilter

	_filterFactories[_typeInterface] = func(query string) _filter {
		return func(ep *BindEndpoint) bool {
			return _queryMatch(query, ep.RandomVersion().Service.Interface)
		}
	}
}

func _queryMatch(input, expected string) bool {
	input, expected = strings.ToLower(input), strings.ToLower(expected)
	return input == expected || strings.Contains(expected, input)
}

func _queryEndpoints(data map[string]*BindEndpoint, request *http.Request) interface{} {
	filters := make([]_filter, 0)
	query := request.URL.Query()
	for _, key := range _typeKeys {
		if query := query.Get(key); "" != query {
			if f, ok := _filterFactories[key]; ok {
				filters = append(filters, f(query))
			}
		}
	}
	if len(filters) == 0 {
		m := make(map[string]map[string]*flux.Endpoint, 16)
		for k, v := range data {
			m[k] = v.ToSerializable()
		}
		return m
	}
	return _queryWithFilters(data, filters...)
}

func _queryWithFilters(data map[string]*BindEndpoint, filters ..._filter) []map[string]*flux.Endpoint {
	items := make([]map[string]*flux.Endpoint, 0, 16)
DataLoop:
	for _, v := range data {
		for _, filter := range filters {
			// 任意Filter返回True
			if filter(v) {
				items = append(items, v.ToSerializable())
				continue DataLoop
			}
		}
	}
	return items
}
