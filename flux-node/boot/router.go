package boot

import (
	"context"
	"fmt"
	flux2 "github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-node/ext"
	"github.com/bytepowered/flux/flux-node/logger"
	"github.com/prometheus/client_golang/prometheus"
	"reflect"
	"sort"
	"time"
)

type Router struct {
	metrics *Metrics
	hooks   []flux2.PrepareHookFunc
}

func NewRouter() *Router {
	return &Router{
		metrics: NewMetrics(),
		hooks:   make([]flux2.PrepareHookFunc, 0, 4),
	}
}

func (r *Router) Prepare() error {
	logger.Info("Router preparing")
	for _, hook := range append(ext.PrepareHooks(), r.hooks...) {
		if err := hook(); nil != err {
			return err
		}
	}
	return nil
}

func (r *Router) Initial() error {
	logger.Info("Router initialing")
	// Backends
	for proto, backend := range ext.BackendTransports() {
		ns := flux2.NamespaceBackendTransports + "." + proto
		logger.Infow("Load backend", "proto", proto, "type", reflect.TypeOf(backend), "config-ns", ns)
		if err := r.AddInitHook(backend, flux2.NewConfigurationOfNS(ns)); nil != err {
			return err
		}
	}
	// 手动注册的单实例Filters
	for _, filter := range append(ext.GlobalFilters(), ext.SelectiveFilters()...) {
		ns := filter.TypeId()
		logger.Infow("Load static-filter", "type", reflect.TypeOf(filter), "config-ns", ns)
		config := flux2.NewConfigurationOfNS(ns)
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
		if filter, ok := filter.(flux2.Filter); ok {
			ext.AddSelectiveFilter(filter)
		}
	}
	return nil
}

func (r *Router) AddInitHook(ref interface{}, config *flux2.Configuration) error {
	if init, ok := ref.(flux2.Initializer); ok {
		if err := init.Init(config); nil != err {
			return err
		}
	}
	ext.AddHookFunc(ref)
	return nil
}

func (r *Router) Startup() error {
	for _, startup := range sortedStartup(ext.StartupHooks()) {
		if err := startup.Startup(); nil != err {
			return err
		}
	}
	return nil
}

func (r *Router) Shutdown(ctx context.Context) error {
	for _, shutdown := range sortedShutdown(ext.ShutdownHooks()) {
		if err := shutdown.Shutdown(ctx); nil != err {
			return err
		}
	}
	return nil
}

func (r *Router) Route(ctx flux2.Context) *flux2.ServeError {
	// 统计异常
	doMetricEndpointFunc := func(err *flux2.ServeError) *flux2.ServeError {
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
	selective := make([]flux2.Filter, 0, 16)
	for _, selector := range ext.FilterSelectors() {
		if selector.Activate(ctx) {
			selective = append(selective, selector.DoSelect(ctx)...)
		}
	}
	ctx.AddMetric("selector", time.Since(ctx.StartAt()))
	transport := func(ctx flux2.Context) *flux2.ServeError {
		select {
		case <-ctx.Context().Done():
			return &flux2.ServeError{
				StatusCode: flux2.StatusOK,
				ErrorCode:  "ROUTE:TRANSPORT/B:CANCELED",
				CauseError: ctx.Context().Err(),
			}
		default:
			break
		}
		defer func() {
			ctx.AddMetric("backend", time.Since(ctx.StartAt()))
		}()
		protoName := ctx.BackendService().AttrRpcProto()
		if backend, ok := ext.BackendTransportByProto(protoName); !ok {
			logger.TraceContext(ctx).Errorw("SERVER:ROUTE:UNSUPPORTED_PROTOCOL",
				"proto", protoName, "service", ctx.Endpoint().Service)
			return &flux2.ServeError{
				StatusCode: flux2.StatusNotFound,
				ErrorCode:  flux2.ErrorCodeRequestNotFound,
				Message:    fmt.Sprintf("ROUTE:UNKNOWN_PROTOCOL:%s", protoName),
			}
		} else {
			// Backend exchange
			timer := prometheus.NewTimer(r.metrics.RouteDuration.WithLabelValues("BackendTransport", protoName))
			err := backend.Exchange(ctx)
			timer.ObserveDuration()
			return err
		}
	}
	// Walk filters
	filters := append(ext.GlobalFilters(), selective...)
	return doMetricEndpointFunc(r.walk(transport, filters)(ctx))
}

func (r *Router) walk(next flux2.FilterHandler, filters []flux2.Filter) flux2.FilterHandler {
	for i := len(filters) - 1; i >= 0; i-- {
		next = filters[i].DoFilter(next)
	}
	return next
}

func isDisabled(config *flux2.Configuration) bool {
	return config.GetBool("disable") || config.GetBool("disabled")
}

func sortedStartup(items []flux2.Startuper) []flux2.Startuper {
	out := make(StartupArray, len(items))
	for i, v := range items {
		out[i] = v
	}
	sort.Sort(out)
	return out
}

func sortedShutdown(items []flux2.Shutdowner) []flux2.Shutdowner {
	out := make(ShutdownArray, len(items))
	for i, v := range items {
		out[i] = v
	}
	sort.Sort(out)
	return out
}
