package boot

import (
	"context"
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"github.com/prometheus/client_golang/prometheus"
	"reflect"
	"sort"
	"time"
)

type Router struct {
	metrics *Metrics
	hooks   []flux.PrepareHookFunc
}

func NewRouter() *Router {
	return &Router{
		metrics: NewMetrics(),
		hooks:   make([]flux.PrepareHookFunc, 0, 4),
	}
}

func (r *Router) Prepare() error {
	logger.Info("Router preparing")
	for _, hook := range append(ext.GetPrepareHooks(), r.hooks...) {
		if err := hook(); nil != err {
			return err
		}
	}
	return nil
}

func (r *Router) Initial() error {
	logger.Info("Router initialing")
	// Backends
	for proto, backend := range ext.GetBackendTransports() {
		ns := flux.NamespaceBackendTransports + "." + proto
		logger.Infow("Load backend", "proto", proto, "type", reflect.TypeOf(backend), "config-ns", ns)
		if err := r.AddInitHook(backend, flux.NewConfigurationOfNS(ns)); nil != err {
			return err
		}
	}
	// 手动注册的单实例Filters
	for _, filter := range append(ext.GetGlobalFilters(), ext.GetSelectiveFilters()...) {
		ns := filter.TypeId()
		logger.Infow("Load static-filter", "type", reflect.TypeOf(filter), "config-ns", ns)
		config := flux.NewConfigurationOfNS(ns)
		if isDisabled(config) {
			logger.Infow("Set static-filter DISABLED", "filter-id", filter.TypeId())
			continue
		}
		if err := r.AddInitHook(filter, config); nil != err {
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
		if isDisabled(item.Config) {
			logger.Infow("Set dynamic-filter DISABLED", "filter-id", item.Id, "type-id", item.TypeId)
			continue
		}
		if err := r.AddInitHook(filter, item.Config); nil != err {
			return err
		}
		if filter, ok := filter.(flux.Filter); ok {
			ext.AddSelectiveFilter(filter)
		}
	}
	return nil
}

func (r *Router) AddInitHook(ref interface{}, config *flux.Configuration) error {
	if init, ok := ref.(flux.Initializer); ok {
		if err := init.Init(config); nil != err {
			return err
		}
	}
	ext.AddHookFunc(ref)
	return nil
}

func (r *Router) Startup() error {
	for _, startup := range sortedStartup(ext.GetStartupHooks()) {
		if err := startup.Startup(); nil != err {
			return err
		}
	}
	return nil
}

func (r *Router) Shutdown(ctx context.Context) error {
	for _, shutdown := range sortedShutdown(ext.GetShutdownHooks()) {
		if err := shutdown.Shutdown(ctx); nil != err {
			return err
		}
	}
	return nil
}

func (r *Router) Route(ctx flux.Context) *flux.ServeError {
	// 统计异常
	doMetricEndpointFunc := func(err *flux.ServeError) *flux.ServeError {
		// Access Counter: ProtoName, Interface, Method
		backend := ctx.BackendService()
		proto, uri, method := backend.AttrRpcProto(), backend.Interface, backend.Method
		r.metrics.EndpointAccess.WithLabelValues(proto, uri, method).Inc()
		if nil != err {
			// Error Counter: ProtoName, Interface, Method, ErrorCode
			r.metrics.EndpointError.WithLabelValues(proto, uri, method, err.GetErrorCode()).Inc()
		}
		return err
	}
	// Metric: Route
	defer func() {
		ctx.AddMetric("route", time.Since(ctx.StartAt()))
	}()
	// Select filters
	selective := make([]flux.Filter, 0, 16)
	for _, selector := range ext.GetSelectors() {
		if selector.Activate(ctx) {
			selective = append(selective, selector.DoSelect(ctx)...)
		}
	}
	ctx.AddMetric("selector", time.Since(ctx.StartAt()))
	// Walk filters
	filters := append(ext.GetGlobalFilters(), selective...)
	err := r.walk(func(ctx flux.Context) *flux.ServeError {
		defer func() {
			ctx.AddMetric("backend", time.Since(ctx.StartAt()))
		}()
		protoName := ctx.BackendService().AttrRpcProto()
		if backend, ok := ext.GetBackendTransport(protoName); !ok {
			logger.TraceContext(ctx).Errorw("SERVER:ROUTE:UNSUPPORTED_PROTOCOL",
				"proto", protoName, "service", ctx.Endpoint().Service)
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
	}, filters)(ctx)
	return doMetricEndpointFunc(err)
}

func (r *Router) walk(next flux.FilterHandler, filters []flux.Filter) flux.FilterHandler {
	for i := len(filters) - 1; i >= 0; i-- {
		next = filters[i].DoFilter(next)
	}
	return next
}

func isDisabled(config *flux.Configuration) bool {
	return config.GetBool("disable") || config.GetBool("disabled")
}

func sortedStartup(items []flux.Startuper) []flux.Startuper {
	out := make(StartupArray, len(items))
	for i, v := range items {
		out[i] = v
	}
	sort.Sort(out)
	return out
}

func sortedShutdown(items []flux.Shutdowner) []flux.Shutdowner {
	out := make(ShutdownArray, len(items))
	for i, v := range items {
		out[i] = v
	}
	sort.Sort(out)
	return out
}
