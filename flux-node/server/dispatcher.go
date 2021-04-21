package server

import (
	"context"
	"fmt"
	"github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-node/ext"
	"github.com/bytepowered/flux/flux-node/logger"
	"github.com/prometheus/client_golang/prometheus"
	"reflect"
	"sort"
	"time"
)

type Dispatcher struct {
	metrics *Metrics
	hooks   []flux.PrepareHookFunc
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		metrics: NewMetrics(),
		hooks:   make([]flux.PrepareHookFunc, 0, 4),
	}
}

func (r *Dispatcher) Prepare() error {
	logger.Info("Dispatcher preparing")
	for _, hook := range append(ext.PrepareHooks(), r.hooks...) {
		if err := hook(); nil != err {
			return err
		}
	}
	return nil
}

func (r *Dispatcher) Initial() error {
	logger.Info("Dispatcher initialing")
	// Transporter
	for proto, transporter := range ext.Transporters() {
		ns := flux.NamespaceTransporters + "." + proto
		logger.Infow("Load transporter", "proto", proto, "type", reflect.TypeOf(transporter), "config-ns", ns)
		if err := r.AddInitHook(transporter, flux.NewConfiguration(ns)); nil != err {
			return err
		}
	}
	// 手动注册的单实例Filters
	for _, filter := range append(ext.GlobalFilters(), ext.SelectiveFilters()...) {
		ns := filter.FilterId()
		logger.Infow("Load static-filter", "type", reflect.TypeOf(filter), "config-ns", ns)
		config := flux.NewConfiguration(ns)
		if IsDisabled(config) {
			logger.Infow("Set static-filter DISABLED", "filter-id", filter.FilterId())
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
		if IsDisabled(item.Config) {
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

func (r *Dispatcher) AddInitHook(ref interface{}, config *flux.Configuration) error {
	if init, ok := ref.(flux.Initializer); ok {
		if err := init.Init(config); nil != err {
			return err
		}
	}
	ext.AddHookFunc(ref)
	return nil
}

func (r *Dispatcher) Startup() error {
	for _, startup := range sortedStartup(ext.StartupHooks()) {
		if err := startup.Startup(); nil != err {
			return err
		}
	}
	return nil
}

func (r *Dispatcher) Shutdown(ctx context.Context) error {
	for _, shutdown := range sortedShutdown(ext.ShutdownHooks()) {
		if err := shutdown.Shutdown(ctx); nil != err {
			return err
		}
	}
	return nil
}

func (r *Dispatcher) Route(ctx *flux.Context) *flux.ServeError {
	// 统计异常
	doMetricEndpointFunc := func(err *flux.ServeError) *flux.ServeError {
		// Access Counter: ProtoName, Interface, Method
		service := ctx.Service()
		proto, uri, method := service.RpcProto(), service.Interface, service.Method
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
	for _, selector := range ext.FilterSelectors() {
		if selector.Activate(ctx) {
			selective = append(selective, selector.DoSelect(ctx)...)
		}
	}
	ctx.AddMetric("selector", time.Since(ctx.StartAt()))
	transport := func(ctx *flux.Context) *flux.ServeError {
		select {
		case <-ctx.Context().Done():
			return &flux.ServeError{
				StatusCode: flux.StatusOK,
				ErrorCode:  "ROUTE:TRANSPORT/B:CANCELED",
				CauseError: ctx.Context().Err(),
			}
		default:
			break
		}
		defer func() {
			ctx.AddMetric("transporter", time.Since(ctx.StartAt()))
		}()
		proto := ctx.Service().RpcProto()
		transporter, ok := ext.TransporterByProto(proto)
		if !ok {
			logger.TraceContext(ctx).Errorw("SERVER:ROUTE:UNSUPPORTED_PROTOCOL",
				"proto", proto, "service", ctx.Endpoint().Service)
			return &flux.ServeError{
				StatusCode: flux.StatusNotFound,
				ErrorCode:  flux.ErrorCodeRequestNotFound,
				Message:    fmt.Sprintf("ROUTE:UNKNOWN_PROTOCOL:%s", proto),
			}
		}
		// Transporter exchange
		timer := prometheus.NewTimer(r.metrics.RouteDuration.WithLabelValues("Transporter", proto))
		transporter.Transport(ctx)
		timer.ObserveDuration()
		return nil
	}
	// Walk filters
	filters := append(ext.GlobalFilters(), selective...)
	return doMetricEndpointFunc(r.walk(transport, filters)(ctx))
}

func (r *Dispatcher) walk(next flux.FilterInvoker, filters []flux.Filter) flux.FilterInvoker {
	for i := len(filters) - 1; i >= 0; i-- {
		next = filters[i].DoFilter(next)
	}
	return next
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

type StartupArray []flux.Startuper

func (s StartupArray) Len() int           { return len(s) }
func (s StartupArray) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s StartupArray) Less(i, j int) bool { return orderOf(s[i]) < orderOf(s[j]) }

type ShutdownArray []flux.Shutdowner

func (s ShutdownArray) Len() int           { return len(s) }
func (s ShutdownArray) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s ShutdownArray) Less(i, j int) bool { return orderOf(s[i]) < orderOf(s[j]) }

func orderOf(v interface{}) int {
	if v, ok := v.(flux.Orderer); ok {
		return v.Order()
	} else {
		return 0
	}
}
