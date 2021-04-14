package fluxinspect

import (
	"github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-node/ext"
)

const (
	epQueryKeyApplication = "application"
	epQueryKeyProtocol    = "protocol"
	epQueryKeyPattern     = "pattern"
	epQueryKeyInterface   = "interface"
)

type EndpointFilter func(ep *flux.MVCEndpoint) bool

var (
	endpointQueryKeys = []string{epQueryKeyApplication, epQueryKeyProtocol, epQueryKeyPattern, epQueryKeyInterface}
	endpointFilters   = make(map[string]func(string) EndpointFilter)
)

func init() {
	endpointFilters[epQueryKeyApplication] = func(query string) EndpointFilter {
		return func(ep *flux.MVCEndpoint) bool {
			return !ep.IsEmpty() && queryMatch(query, ep.Random().Application)
		}
	}
	endpointFilters[epQueryKeyProtocol] = func(query string) EndpointFilter {
		return func(ep *flux.MVCEndpoint) bool {
			proto := ep.Random().Service.RpcProto()
			return !ep.IsEmpty() && queryMatch(query, proto)
		}
	}
	endpointFilters[epQueryKeyPattern] = func(query string) EndpointFilter {
		return func(ep *flux.MVCEndpoint) bool {
			return !ep.IsEmpty() && queryMatch(query, ep.Random().HttpPattern)
		}
	}
	endpointFilters[epQueryKeyInterface] = func(query string) EndpointFilter {
		return func(ep *flux.MVCEndpoint) bool {
			return !ep.IsEmpty() && queryMatch(query, ep.Random().Service.Interface)
		}
	}
}

func DoQueryEndpoints(args func(key string) string) []*flux.Endpoint {
	filters := make([]EndpointFilter, 0)
	for _, key := range endpointQueryKeys {
		if value := args(key); value != "" {
			if f, ok := endpointFilters[key]; ok {
				filters = append(filters, f(value))
			}
		}
	}
	endpoints := ext.Endpoints()
	if len(filters) == 0 {
		out := make([]*flux.Endpoint, 0, len(endpoints))
		for _, mep := range endpoints {
			out = append(out, mep.Endpoints()...)
		}
		return out
	}
	return queryEndpointByFilters(endpoints, filters...)
}

func EndpointsHandler(webex flux.ServerWebContext) error {
	m := DoQueryEndpoints(func(key string) string {
		return webex.QueryVar(key)
	})
	return send(webex, flux.StatusOK, m)
}

func queryEndpointByFilters(endpoints map[string]*flux.MVCEndpoint, filters ...EndpointFilter) []*flux.Endpoint {
	out := make([]*flux.Endpoint, 0, 16)
	for _, ep := range endpoints {
		passed := true
		for _, filter := range filters {
			passed = filter(ep)
			if !passed {
				break
			}
		}
		if passed {
			out = append(out, ep.Endpoints()...)
		}
	}
	return out
}
