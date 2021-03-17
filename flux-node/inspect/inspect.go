package inspect

import (
	"github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-node/common"
	"github.com/bytepowered/flux/flux-node/ext"
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
			return !ep.IsEmpty() && queryMatch(query, ep.Random().Application)
		}
	}
	endpointFilterFactories[queryKeyProtocol] = func(query string) EndpointFilter {
		return func(ep *flux.MultiEndpoint) bool {
			proto := ep.Random().Service.RpcProto()
			return !ep.IsEmpty() && queryMatch(query, proto)
		}
	}
	httpPatternFilter := func(query string) EndpointFilter {
		return func(ep *flux.MultiEndpoint) bool {
			return !ep.IsEmpty() && queryMatch(query, ep.Random().HttpPattern)
		}
	}
	endpointFilterFactories[queryKeyHttpPattern] = httpPatternFilter
	endpointFilterFactories[queryKeyHttpPattern0] = httpPatternFilter

	endpointFilterFactories[queryKeyInterface] = func(query string) EndpointFilter {
		return func(ep *flux.MultiEndpoint) bool {
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

func ServicesHandler(ctx flux.ServerWebContext) error {
	for _, key := range serviceQueryKeys {
		if id := ctx.QueryVar(key); "" != id {
			service, ok := ext.TransporterServiceById(id)
			if ok {
				return send(ctx, flux.StatusOK, service)
			} else {
				return send(ctx, flux.StatusNotFound, map[string]string{
					"status":     "failed",
					"message":    "service not found",
					"service-id": id,
				})
			}
		}
	}
	return send(ctx, flux.StatusBadRequest, map[string]string{
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

func send(webex flux.ServerWebContext, status int, payload interface{}) error {
	bytes, err := common.SerializeObject(payload)
	if nil != err {
		return err
	}
	return webex.Write(status, flux.MIMEApplicationJSONCharsetUTF8, bytes)
}
