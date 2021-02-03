package admin

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
	queryKeyHttpPattern0 = "pattern"
	queryKeyInterface    = "interface"
	queryKeyServiceId0   = "service-id"
	queryKeyServiceId1   = "service"
)

type EndpointFilter func(ep *flux.MultiEndpoint) bool

var (
	endpointQueryKeys = []string{queryKeyApplication, queryKeyProtocol,
		queryKeyHttpPattern, queryKeyHttpPattern0,
		queryKeyInterface,
	}
	serviceQueryKeys = []string{queryKeyServiceId0, queryKeyServiceId1}
)

var (
	endpointFilterFactories = make(map[string]func(string) EndpointFilter)
)

func init() {
	endpointFilterFactories[queryKeyApplication] = func(query string) EndpointFilter {
		return func(ep *flux.MultiEndpoint) bool {
			return queryMatch(query, ep.RandomVersion().Application)
		}
	}
	endpointFilterFactories[queryKeyProtocol] = func(query string) EndpointFilter {
		return func(ep *flux.MultiEndpoint) bool {
			proto := ep.RandomVersion().Service.AttrRpcProto()
			return queryMatch(query, proto)
		}
	}
	httpPatternFilter := func(query string) EndpointFilter {
		return func(ep *flux.MultiEndpoint) bool {
			return queryMatch(query, ep.RandomVersion().HttpPattern)
		}
	}
	endpointFilterFactories[queryKeyHttpPattern] = httpPatternFilter
	endpointFilterFactories[queryKeyHttpPattern0] = httpPatternFilter

	endpointFilterFactories[queryKeyInterface] = func(query string) EndpointFilter {
		return func(ep *flux.MultiEndpoint) bool {
			return queryMatch(query, ep.RandomVersion().Service.Interface)
		}
	}
}

func InspectEndpointsHandler(ctx flux.WebContext) error {
	filters := make([]EndpointFilter, 0)
	for _, key := range endpointQueryKeys {
		if query := ctx.QueryVar(key); "" != query {
			if f, ok := endpointFilterFactories[key]; ok {
				filters = append(filters, f(query))
			}
		}
	}
	data := ext.LoadEndpoints()
	if len(filters) == 0 {
		m := make(map[string]map[string]*flux.Endpoint, 16)
		for k, v := range data {
			m[k] = v.ToSerializable()
		}
		return ctx.Send(ctx, http.Header{}, flux.StatusOK, nil)
	} else {
		return ctx.Send(ctx, http.Header{}, flux.StatusOK, queryWithEndpointFilters(data, filters...))
	}
}

func InspectServicesHandler(ctx flux.WebContext) error {
	noheader := http.Header{}
	for _, key := range serviceQueryKeys {
		if id := ctx.QueryVar(key); "" != id {
			service, ok := ext.GetBackendService(id)
			if ok {
				return ctx.Send(ctx, noheader, flux.StatusOK, service)
			} else {
				return ctx.Send(ctx, noheader, flux.StatusNotFound, map[string]string{
					"status":     "failed",
					"message":    "service not found",
					"service-id": id,
				})
			}
		}
	}
	return ctx.Send(ctx, noheader, flux.StatusBadRequest, map[string]string{
		"status":  "failed",
		"message": "param is required: serviceId",
	})
}

func queryWithEndpointFilters(data map[string]*flux.MultiEndpoint, filters ...EndpointFilter) []map[string]*flux.Endpoint {
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

func queryMatch(input, expected string) bool {
	input, expected = strings.ToLower(input), strings.ToLower(expected)
	return strings.Contains(expected, input)
}
