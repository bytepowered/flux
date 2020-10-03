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

type RouterEngine struct {
	metrics *Metrics
}

func NewRouteEngine() *RouterEngine {
	return &RouterEngine{
		metrics: NewMetrics(),
	}
}

func (r *RouterEngine) Initial() error {
	logger.Infof("RouterEngine initialing")
	// Backends
	for proto, bk := range ext.Backends() {
		ns := "BACKEND." + proto
		logger.Infow("Load backend", "proto", proto, "type", reflect.TypeOf(bk), "config-ns", ns)
		if err := r.InitialHook(bk, flux.NewConfigurationOf(ns)); nil != err {
			return err
		}
	}
	// 手动注册的单实例Filters
	for _, filter := range append(ext.GlobalFilters(), ext.SelectiveFilters()...) {
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
			ext.AddSelectiveFilter(filter)
		}
	}
	return nil
}

func (r *RouterEngine) InitialHook(ref interface{}, config *flux.Configuration) error {
	if init, ok := ref.(flux.Initializer); ok {
		if err := init.Init(config); nil != err {
			return err
		}
	}
	ext.AddLifecycleHook(ref)
	return nil
}

func (r *RouterEngine) Startup() error {
	for _, startup := range sortedStartup(ext.GetStartupHooks()) {
		if err := startup.Startup(); nil != err {
			return err
		}
	}
	return nil
}

func (r *RouterEngine) Shutdown(ctx context.Context) error {
	for _, shutdown := range sortedShutdown(ext.GetShutdownHooks()) {
		if err := shutdown.Shutdown(ctx); nil != err {
			return err
		}
	}
	return nil
}

func (r *RouterEngine) Route(ctx *WrappedContext) *flux.StateError {
	doMetricEndpoint := func(err *flux.StateError) *flux.StateError {
		// Access Counter: ProtoName, UpstreamUri, UpstreamMethod
		proto, uri, method := ctx.EndpointProto(), ctx.endpoint.UpstreamUri, ctx.endpoint.UpstreamMethod
		r.metrics.EndpointAccess.WithLabelValues(proto, uri, method).Inc()
		if nil != err {
			// Error Counter: ProtoName, UpstreamUri, UpstreamMethod, ErrorCode
			r.metrics.EndpointError.WithLabelValues(proto, uri, method, err.ErrorCode).Inc()
		}
		return err
	}
	// Resolve arguments
	if argumentNeedResolve(ctx, ctx.EndpointArguments()) {
		if err := argumentResolveWith(ext.GetArgumentValueLookupFunc(), ctx.EndpointArguments(), ctx); nil != err {
			return doMetricEndpoint(err)
		}
	}
	// Select filters
	globals := ext.GlobalFilters()
	selectives := make([]flux.Filter, 0)
	host := ctx.Request().Host()
	for _, selector := range ext.FindSelectors(host) {
		for _, typeId := range selector.Select(ctx).Filters {
			if f, ok := ext.GetSelectiveFilter(typeId); ok {
				selectives = append(selectives, f)
			} else {
				logger.Trace(ctx.RequestId()).Warnw("Filter not found on selector", "type-id", typeId)
			}
		}
	}
	// Walk filters
	err := r.walk(func(ctx flux.Context) *flux.StateError {
		protoName := ctx.EndpointProto()
		if backend, ok := ext.GetBackend(protoName); !ok {
			return &flux.StateError{
				StatusCode: flux.StatusNotFound,
				ErrorCode:  flux.ErrorCodeRequestNotFound,
				Message:    fmt.Sprintf("ROUTE:UNKNOWN_PROTOCOL: %s", protoName)}
		} else {
			timer := prometheus.NewTimer(r.metrics.RouteDuration.WithLabelValues("Backend", protoName))
			ret := backend.Exchange(ctx)
			timer.ObserveDuration()
			return ret
		}
	}, append(globals, selectives...)...)(ctx)
	return doMetricEndpoint(err)
}

func (r *RouterEngine) walk(next flux.FilterHandler, filters ...flux.Filter) flux.FilterHandler {
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
