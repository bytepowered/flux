package server

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"net/http"
	"strings"
)

const (
	_typeApplication = "application"
	_typeProtocol    = "protocol"
	_typeHttpPattern = "http-pattern"
	_typeInterface   = "interface"
)

type _filter func(ep *MultiVersionEndpoint) bool

// 支持以下过滤条件
var _typeKeys = []string{"application", "protocol", "http-pattern", "interface"}

var (
	_filterFactories = make(map[string]func(string) _filter)
)

func DebugQueryEndpoint(datamap map[string]*MultiVersionEndpoint) http.HandlerFunc {
	// Endpoint查询
	serializer := ext.GetSerializer(ext.TypeNameSerializerJson)
	return func(writer http.ResponseWriter, request *http.Request) {
		data := queryEndpoints(datamap, request)
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
		return func(ep *MultiVersionEndpoint) bool {
			return query == ep.RandomVersion().Application
		}
	}
	_filterFactories[_typeProtocol] = func(query string) _filter {
		return func(ep *MultiVersionEndpoint) bool {
			proto := ep.RandomVersion().Service.RpcProto
			return strings.ToLower(query) == strings.ToLower(proto)
		}
	}
	_filterFactories[_typeHttpPattern] = func(query string) _filter {
		return func(ep *MultiVersionEndpoint) bool {
			return query == ep.RandomVersion().HttpPattern
		}
	}
	_filterFactories[_typeInterface] = func(query string) _filter {
		return func(ep *MultiVersionEndpoint) bool {
			return query == ep.RandomVersion().Service.Interface
		}
	}
}

func queryEndpoints(data map[string]*MultiVersionEndpoint, request *http.Request) interface{} {
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

func _queryWithFilters(data map[string]*MultiVersionEndpoint, filters ..._filter) []map[string]*flux.Endpoint {
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
