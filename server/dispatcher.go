package server

import (
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/internal"
	"github.com/bytepowered/flux/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cast"
	"time"
)

type Dispatcher struct {
	metrics                *Metrics
	writer                 flux.ServeResponseWriter
	onBeforeFilterHooks    []flux.OnBeforeFilterHookFunc
	onBeforeTransportHooks []flux.OnBeforeTransportHookFunc
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		metrics:                NewMetrics(),
		writer:                 new(internal.JSONServeResponseWriter),
		onBeforeFilterHooks:    make([]flux.OnBeforeFilterHookFunc, 0, 4),
		onBeforeTransportHooks: make([]flux.OnBeforeTransportHookFunc, 0, 4),
	}
}

func (d *Dispatcher) dispatch(ctx *flux.Context) *flux.ServeError {
	// 统计异常
	doMetricEndpointFunc := func(err *flux.ServeError) *flux.ServeError {
		// Access Counter: ProtoName, Interface, Method
		service := ctx.Service()
		proto, uri, method := service.RpcProto(), service.Interface, service.Method
		d.metrics.EndpointAccess.WithLabelValues(proto, uri, method).Inc()
		if flux.NotNil(err) {
			// Error Counter: ProtoName, Interface, Method, ErrorCode
			d.metrics.EndpointError.WithLabelValues(proto, uri, method, cast.ToString(err.ErrorCode)).Inc()
		}
		return err
	}
	// Metric: Route
	defer func() {
		ctx.AddMetric("dispatch", time.Since(ctx.StartAt()))
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
			return &flux.ServeError{StatusCode: flux.StatusBadRequest,
				ErrorCode: "DISPATCHER:TRANSPORT:CANCELED/100", CauseError: ctx.Context().Err(),
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
			return &flux.ServeError{StatusCode: flux.StatusNotFound,
				ErrorCode: flux.ErrorCodeRequestNotFound,
				Message:   fmt.Sprintf("SERVER:ROUTE:ILLEGAL_PROTOCOL/%s", proto),
			}
		}
		// Transporter invoke
		for _, hook := range d.onBeforeTransportHooks {
			hook(ctx, transporter)
		}
		timer := prometheus.NewTimer(d.metrics.RouteDuration.WithLabelValues("Transporter", proto))
		invret, inverr := transporter.DoInvoke(ctx, ctx.Service())
		timer.ObserveDuration()
		select {
		case <-ctx.Context().Done():
			return &flux.ServeError{StatusCode: flux.StatusBadRequest,
				ErrorCode: "DISPATCHER:TRANSPORT:CANCELED/200", CauseError: ctx.Context().Err(),
			}
		default:
			break
		}
		// Write response
		if flux.NotNil(inverr) {
			d.writer.WriteError(ctx, inverr)
		} else {
			for k, v := range invret.Attachments {
				ctx.SetAttribute(k, v)
			}
			d.writer.Write(ctx, invret)
		}
		return nil
	}
	// Walk filters
	filters := append(ext.GlobalFilters(), selective...)
	for _, hook := range d.onBeforeFilterHooks {
		hook(ctx, filters)
	}
	return doMetricEndpointFunc(d.walk(transport, filters)(ctx))
}

func (d *Dispatcher) walk(next flux.FilterInvoker, filters []flux.Filter) flux.FilterInvoker {
	for i := len(filters) - 1; i >= 0; i-- {
		next = filters[i].DoFilter(next)
	}
	return next
}

func (d *Dispatcher) setResponseWriter(w flux.ServeResponseWriter) {
	d.writer = w
}

func (d *Dispatcher) addOnBeforeFilterHook(h flux.OnBeforeFilterHookFunc) {
	d.onBeforeFilterHooks = append(d.onBeforeFilterHooks, h)
}

func (d *Dispatcher) addOnBeforeTransportHook(h flux.OnBeforeTransportHookFunc) {
	d.onBeforeTransportHooks = append(d.onBeforeTransportHooks, h)
}
