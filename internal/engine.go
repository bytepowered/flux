package internal

import (
	"context"
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"reflect"
)

var (
	defaultMetricNamespace = "flux"
	defaultMetricSubsystem = "http"
	defaultMetricBuckets   = []float64{
		0.0005,
		0.001, // 1ms
		0.002,
		0.005,
		0.01, // 10ms
		0.02,
		0.05,
		0.1, // 100 ms
		0.2,
		0.5,
		1.0, // 1s
		2.0,
		5.0,
		10.0, // 10s
		15.0,
		20.0,
		30.0,
	}
)

type RouterEngine struct {
	metricEndpointAccess *prometheus.CounterVec
	metricEndpointError  *prometheus.CounterVec
	metricRouteDuration  *prometheus.HistogramVec
}

func NewRouteEngine() *RouterEngine {
	return &RouterEngine{
		metricEndpointAccess: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: defaultMetricNamespace,
			Subsystem: defaultMetricSubsystem,
			Name:      "endpoint_access_total",
			Help:      "Number of endpoint access",
		}, []string{"ProtoName", "UpstreamUri", "UpstreamMethod"}),
		metricEndpointError: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: defaultMetricNamespace,
			Subsystem: defaultMetricSubsystem,
			Name:      "endpoint_error_total",
			Help:      "Number of endpoint access errors",
		}, []string{"ProtoName", "UpstreamUri", "UpstreamMethod", "ErrorCode"}),
		metricRouteDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: defaultMetricNamespace,
			Subsystem: defaultMetricSubsystem,
			Name:      "endpoint_route_duration",
			Help:      "Spend time by processing a endpoint",
			Buckets:   defaultMetricBuckets,
		}, []string{"ComponentType", "TypeId"}),
	}
}

func (r *RouterEngine) Initial() error {
	logger.Infof("RouterEngine initialing")
	// Exchanges
	for proto, ex := range ext.Exchanges() {
		ns := "EXCHANGE." + proto
		logger.Infow("Load exchange", "proto", proto, "type", reflect.TypeOf(ex), "config-ns", ns)
		if err := r.InitialHook(ex, flux.NewConfigurationOf(ns)); nil != err {
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

func (r *RouterEngine) Route(ctx *ContextWrapper) *flux.StateError {
	doMetricEndpoint := func(err *flux.StateError) *flux.StateError {
		// Access Counter: ProtoName, UpstreamUri, UpstreamMethod
		proto, uri, method := ctx.EndpointProtoName(), ctx.endpoint.UpstreamUri, ctx.endpoint.UpstreamMethod
		r.metricEndpointAccess.WithLabelValues(proto, uri, method).Inc()
		if nil != err {
			// Error Counter: ProtoName, UpstreamUri, UpstreamMethod, ErrorCode
			r.metricEndpointError.WithLabelValues(proto, uri, method, err.ErrorCode).Inc()
		}
		return err
	}
	// Resolve arguments
	if shouldResolve(ctx, ctx.EndpointArguments()) {
		if err := resolveArguments(ext.GetArgumentLookupFunc(), ctx.EndpointArguments(), ctx); nil != err {
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
		protoName := ctx.EndpointProtoName()
		if exchange, ok := ext.GetExchange(protoName); !ok {
			return &flux.StateError{
				StatusCode: flux.StatusNotFound,
				ErrorCode:  flux.ErrorCodeRequestNotFound,
				Message:    fmt.Sprintf("ROUTE:UNKNOWN_PROTOCOL: %s", protoName)}
		} else {
			timer := prometheus.NewTimer(r.metricRouteDuration.WithLabelValues("Exchange", protoName))
			ret := exchange.Exchange(ctx)
			timer.ObserveDuration()
			return ret
		}
	}, append(globals, selectives...)...)(ctx)
	return doMetricEndpoint(err)
}

func (r *RouterEngine) walk(next flux.FilterInvoker, filters ...flux.Filter) flux.FilterInvoker {
	for i := len(filters) - 1; i >= 0; i-- {
		timer := prometheus.NewTimer(r.metricRouteDuration.WithLabelValues("Filter", filters[i].TypeId()))
		next = filters[i].Invoke(next)
		timer.ObserveDuration()
	}
	return next
}

func _isDisabled(config *flux.Configuration) bool {
	return config.GetBool("disable") || config.GetBool("disabled")
}
