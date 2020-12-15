package server

import (
	"context"
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/support"
	"github.com/prometheus/client_golang/prometheus"
	"reflect"
	"sort"
)

type Router struct {
	metrics *Metrics
}

func NewRouter() *Router {
	return &Router{
		metrics: NewMetrics(),
	}
}

func (r *Router) Initial() error {
	logger.Infof("Router initialing")
	// Backends
	for proto, backend := range ext.LoadBackendTransports() {
		ns := "BACKEND." + proto
		logger.Infow("Load backend", "proto", proto, "type", reflect.TypeOf(backend), "config-ns", ns)
		if err := r.InitialHook(backend, flux.NewConfigurationOf(ns)); nil != err {
			return err
		}
	}
	// 手动注册的单实例Filters
	for _, filter := range append(ext.LoadGlobalFilters(), ext.LoadSelectiveFilters()...) {
		ns := filter.TypeId()
		logger.Infow("Load static-filter", "type", reflect.TypeOf(filter), "config-ns", ns)
		config := flux.NewConfigurationOf(ns)
		if _isDisabled(config) {
			logger.Infow("Set static-filter DISABLED", "filter-id", filter.TypeId())
			continue
		}
		if err := r.InitialHook(filter, config); nil != err {
			return err
		}
	}
	// 加载和注册，动态多实例Filter
	dynFilters, err := dynamicFilters()
	if nil != err {
		return err
	}
	for _, item := range dynFilters {
		filter := item.Factory()
		logger.Infow("Load dynamic-filter", "filter-id", item.Id, "type-id", item.TypeId, "type", reflect.TypeOf(filter))
		if _isDisabled(item.Config) {
			logger.Infow("Set dynamic-filter DISABLED", "filter-id", item.Id, "type-id", item.TypeId)
			continue
		}
		if err := r.InitialHook(filter, item.Config); nil != err {
			return err
		}
		if filter, ok := filter.(flux.Filter); ok {
			ext.StoreSelectiveFilter(filter)
		}
	}
	return nil
}

func (r *Router) InitialHook(ref interface{}, config *flux.Configuration) error {
	if init, ok := ref.(flux.Initializer); ok {
		if err := init.Init(config); nil != err {
			return err
		}
	}
	ext.StoreHookFunc(ref)
	return nil
}

func (r *Router) Startup() error {
	for _, startup := range sortedStartup(ext.LoadStartupHooks()) {
		if err := startup.Startup(); nil != err {
			return err
		}
	}
	return nil
}

func (r *Router) Shutdown(ctx context.Context) error {
	for _, shutdown := range sortedShutdown(ext.LoadShutdownHooks()) {
		if err := shutdown.Shutdown(ctx); nil != err {
			return err
		}
	}
	return nil
}

func (r *Router) Route(ctx *WrappedContext) *flux.ServeError {
	// 统计异常
	doMetricEndpointFunc := func(err *flux.ServeError) *flux.ServeError {
		// Access Counter: ProtoName, Interface, Method
		proto, _, uri, method := ctx.ServiceInterface()
		r.metrics.EndpointAccess.WithLabelValues(proto, uri, method).Inc()
		if nil != err {
			// Error Counter: ProtoName, Interface, Method, ErrorCode
			r.metrics.EndpointError.WithLabelValues(proto, uri, method, err.GetErrorCode()).Inc()
		}
		return err
	}
	// Select filters
	globals := ext.LoadGlobalFilters()
	selective := make([]flux.Filter, 0, 16)
	for _, selector := range ext.FindSelectors(ctx.Request().Host()) {
		for _, typeId := range selector.Select(ctx).FilterId {
			if f, ok := ext.LoadSelectiveFilter(typeId); ok {
				selective = append(selective, f)
			} else {
				logger.TraceContext(ctx).Warnw("Filter not found on selector", "type-id", typeId)
			}
		}
	}
	// Walk filters
	err := r.walk(func(ctx flux.Context) *flux.ServeError {
		protoName := ctx.ServiceProto()
		if backend, ok := ext.LoadBackend(protoName); !ok {
			logger.TraceContext(ctx).Warnw("Route, unsupported protocol", "proto", protoName, "service", ctx.Endpoint().Service)
			return &flux.ServeError{
				StatusCode: flux.StatusNotFound,
				ErrorCode:  flux.ErrorCodeRequestNotFound,
				Message:    fmt.Sprintf("ROUTE:UNKNOWN_PROTOCOL:%s", protoName)}
		} else {
			// Backend exchange
			timer := prometheus.NewTimer(r.metrics.RouteDuration.WithLabelValues("BackendTransport", protoName))
			ret := backend.Exchange(ctx)
			timer.ObserveDuration()
			return ret
		}
	}, append(globals, selective...)...)(ctx)
	return doMetricEndpointFunc(err)
}

func (r *Router) walk(next flux.FilterHandler, filters ...flux.Filter) flux.FilterHandler {
	for i := len(filters) - 1; i >= 0; i-- {
		timer := prometheus.NewTimer(r.metrics.RouteDuration.WithLabelValues("Filter", filters[i].TypeId()))
		next = filters[i].DoFilter(next)
		timer.ObserveDuration()
	}
	return next
}

func _isDisabled(config *flux.Configuration) bool {
	return config.GetBool("disable") || config.GetBool("disabled")
}

func sortedStartup(items []flux.Startuper) []flux.Startuper {
	out := make(support.StartupArray, len(items))
	for i, v := range items {
		out[i] = v
	}
	sort.Sort(out)
	return out
}

func sortedShutdown(items []flux.Shutdowner) []flux.Shutdowner {
	out := make(support.ShutdownArray, len(items))
	for i, v := range items {
		out[i] = v
	}
	sort.Sort(out)
	return out
}
