package fluxinspect

import (
	"github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-node/ext"
)

const (
	srvQueryKeyServiceId = "id"
	srvQueryKeyInterface = "interface"
)

type ServiceFilter func(srv *flux.Service) bool

var (
	serviceQueryKeys = []string{srvQueryKeyServiceId, srvQueryKeyInterface}
	serviceFilters   = make(map[string]func(string) ServiceFilter)
)

func init() {
	serviceFilters[srvQueryKeyServiceId] = func(query string) ServiceFilter {
		return func(srv *flux.Service) bool {
			return srv.IsValid() && queryMatch(query, srv.ServiceID())
		}
	}
	serviceFilters[srvQueryKeyInterface] = func(query string) ServiceFilter {
		return func(srv *flux.Service) bool {
			return srv.IsValid() && queryMatch(query, srv.Interface)
		}
	}
}

func DoQueryServices(args func(key string) string) []flux.Service {
	filters := make([]ServiceFilter, 0)
	for _, key := range serviceQueryKeys {
		if value := args(key); value != "" {
			if f, ok := serviceFilters[key]; ok {
				filters = append(filters, f(value))
			}
		}
	}
	services := ext.Services()
	if len(filters) == 0 {
		out := make([]flux.Service, 0, len(services))
		for _, srv := range services {
			out = append(out, srv)
		}
		return out
	}
	return queryServiceByFilters(services, filters...)
}

func ServicesHandler(ctx flux.ServerWebContext) error {
	services := DoQueryServices(func(key string) string {
		return ctx.QueryVar(key)
	})
	return send(ctx, flux.StatusOK, services)
}

func queryServiceByFilters(data map[string]flux.Service, filters ...ServiceFilter) []flux.Service {
	outs := make([]flux.Service, 0, 16)
	for _, srv := range data {
		passed := true
		for _, filter := range filters {
			passed = filter(&srv)
			if !passed {
				break
			}
		}
		if passed {
			outs = append(outs, srv)
		}
	}
	return outs
}
