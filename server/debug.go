package server

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"net/http"
	"strings"
)

const (
	queryKeyApplication  = "application"
	queryKeyProtocol     = "protocol"
	queryKeyHttpPattern  = "http-pattern"
	queryKeyHttpPattern0 = "httpPattern"
	queryKeyHttpPattern1 = "httppattern"
	queryKeyInterface    = "interface"
	queryKeyServiceId    = "serviceid"
	queryKeyServiceId0   = "service-id"
	queryKeyServiceId1   = "serviceId"
)

type EndpointFilter func(ep *BindEndpoint) bool

var (
	endpointQueryKeys = []string{queryKeyApplication, queryKeyProtocol,
		queryKeyHttpPattern, queryKeyHttpPattern0, queryKeyHttpPattern1,
		queryKeyInterface,
	}
	serviceQueryKeys = []string{queryKeyServiceId, queryKeyServiceId0, queryKeyServiceId1}
)

var (
	endpointFilterFactories = make(map[string]func(string) EndpointFilter)
)

func init() {
	endpointFilterFactories[queryKeyApplication] = func(query string) EndpointFilter {
		return func(ep *BindEndpoint) bool {
			return queryMatch(query, ep.RandomVersion().Application)
		}
	}
	endpointFilterFactories[queryKeyProtocol] = func(query string) EndpointFilter {
		return func(ep *BindEndpoint) bool {
			proto := ep.RandomVersion().Service.RpcProto
			return queryMatch(query, proto)
		}
	}
	httpPatternFilter := func(query string) EndpointFilter {
		return func(ep *BindEndpoint) bool {
			return queryMatch(query, ep.RandomVersion().HttpPattern)
		}
	}
	endpointFilterFactories[queryKeyHttpPattern] = httpPatternFilter
	endpointFilterFactories[queryKeyHttpPattern0] = httpPatternFilter
	endpointFilterFactories[queryKeyHttpPattern1] = httpPatternFilter

	endpointFilterFactories[queryKeyInterface] = func(query string) EndpointFilter {
		return func(ep *BindEndpoint) bool {
			return queryMatch(query, ep.RandomVersion().Service.Interface)
		}
	}
}

// NewDebugQueryEndpointHandler Endpoint查询
func NewDebugQueryEndpointHandler(datamap map[string]*BindEndpoint) http.HandlerFunc {
	serializer := ext.LoadSerializer(ext.TypeNameSerializerJson)
	return newSerializableHttpHandler(serializer, func(request *http.Request) interface{} {
		return queryEndpoints(datamap, request)
	})
}

// NewDebugQueryServiceHandler Service查询
func NewDebugQueryServiceHandler() http.HandlerFunc {
	serializer := ext.LoadSerializer(ext.TypeNameSerializerJson)
	return newSerializableHttpHandler(serializer, func(request *http.Request) interface{} {
		query := request.URL.Query()
		for _, key := range serviceQueryKeys {
			if id := query.Get(key); "" != id {
				service, ok := ext.LoadBackendService(id)
				if ok {
					return service
				} else {
					return map[string]string{
						"status":     "failed",
						"message":    "service not found",
						"service-id": id,
					}
				}
			}
		}
		return map[string]string{
			"status":  "failed",
			"message": "param is required: serviceId",
		}
	})
}

func queryEndpoints(data map[string]*BindEndpoint, request *http.Request) interface{} {
	filters := make([]EndpointFilter, 0)
	query := request.URL.Query()
	for _, key := range endpointQueryKeys {
		if query := query.Get(key); "" != query {
			if f, ok := endpointFilterFactories[key]; ok {
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
	return queryWithEndpointFilters(data, filters...)
}

func queryWithEndpointFilters(data map[string]*BindEndpoint, filters ...EndpointFilter) []map[string]*flux.Endpoint {
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

func newSerializableHttpHandler(serializer flux.Serializer, queryHandler func(request *http.Request) interface{}) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		data := queryHandler(request)
		if data, err := serializer.Marshal(data); nil != err {
			writer.WriteHeader(http.StatusInternalServerError)
			_, _ = writer.Write([]byte(err.Error()))
		} else {
			writer.WriteHeader(http.StatusOK)
			writer.Header().Set("Content-Type", "application/json;charset=UTF-8")
			_, _ = writer.Write(data)
		}
	}
}

func queryMatch(input, expected string) bool {
	input, expected = strings.ToLower(input), strings.ToLower(expected)
	return input == expected || strings.Contains(expected, input)
}
