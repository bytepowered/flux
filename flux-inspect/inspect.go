package fluxinspect

import (
	"github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-node/ext"
)

const (
	queryKeyApplication  = "application"
	queryKeyProtocol     = "protocol"
	queryKeyHttpPattern  = "http-pattern"
	queryKeyHttpPattern0 = "pattern"
	queryKeyInterface    = "interface"
)

type EndpointFilter func(ep *flux.MVCEndpoint) bool

var (
	endpointQueryKeys = []string{queryKeyApplication, queryKeyProtocol,
		queryKeyHttpPattern, queryKeyHttpPattern0,
		queryKeyInterface,
	}
)

var (
	endpointFilterFactories = make(map[string]func(string) EndpointFilter)
)

func init() {
	endpointFilterFactories[queryKeyApplication] = func(query string) EndpointFilter {
		return func(ep *flux.MVCEndpoint) bool {
			return !ep.IsEmpty() && queryMatch(query, ep.Random().Application)
		}
	}
	endpointFilterFactories[queryKeyProtocol] = func(query string) EndpointFilter {
		return func(ep *flux.MVCEndpoint) bool {
			proto := ep.Random().Service.RpcProto()
			return !ep.IsEmpty() && queryMatch(query, proto)
		}
	}
	httpPatternFilter := func(query string) EndpointFilter {
		return func(ep *flux.MVCEndpoint) bool {
			return !ep.IsEmpty() && queryMatch(query, ep.Random().HttpPattern)
		}
	}
	endpointFilterFactories[queryKeyHttpPattern] = httpPatternFilter
	endpointFilterFactories[queryKeyHttpPattern0] = httpPatternFilter

	endpointFilterFactories[queryKeyInterface] = func(query string) EndpointFilter {
		return func(ep *flux.MVCEndpoint) bool {
			return !ep.IsEmpty() && queryMatch(query, ep.Random().Service.Interface)
		}
	}
}

func EndpointsHandler(webex flux.ServerWebContext) error {
	filters := make([]EndpointFilter, 0)
	for _, key := range endpointQueryKeys {
		if query := webex.QueryVar(key); "" != query {
			if f, ok := endpointFilterFactories[key]; ok {
				filters = append(filters, f(query))
			}
		}
	}
	if len(filters) == 0 {
		m := make(map[string]map[string]*flux.Endpoint, 16)
		for k, mep := range ext.Endpoints() {
			m[k] = mep.ToSerializable()
		}
		return send(webex, flux.StatusOK, m)
	} else {
		return send(webex, flux.StatusOK,
			queryWithEndpointFilters(ext.Endpoints(), filters...))
	}
}
