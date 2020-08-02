package internal

import (
	"context"
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"reflect"
	"time"
)

type ServerDispatcher struct {
	activeRegistry flux.Registry
}

func NewDispatcher() *ServerDispatcher {
	return &ServerDispatcher{}
}

func (d *ServerDispatcher) Initial() error {
	logger.Infof("Dispatcher initialing")
	// 组件生命周期回调钩子
	initRegisterHook := func(ref interface{}, config *flux.Configuration) error {
		if init, ok := ref.(flux.Initializer); ok {
			if err := init.Init(config); nil != err {
				return err
			}
		}
		ext.AddLifecycleHook(ref)
		return nil
	}
	// 静态注册的单实例内核组件
	// Registry
	if registry, config, err := findActiveRegistry(); nil != err {
		return err
	} else {
		d.activeRegistry = registry
		if err := initRegisterHook(registry, config); nil != err {
			return err
		}
	}
	// Exchanges
	for proto, ex := range ext.Exchanges() {
		ns := "EXCHANGE." + proto
		logger.Infow("Load exchange", "proto", proto, "type", reflect.TypeOf(ex), "config-ns", ns)
		if err := initRegisterHook(ex, flux.NewConfigurationOf(ns)); nil != err {
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
		if err := initRegisterHook(filter, config); nil != err {
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
		if err := initRegisterHook(filter, item.Config); nil != err {
			return err
		}
		if filter, ok := filter.(flux.Filter); ok {
			ext.AddSelectiveFilter(filter)
		}
	}
	return nil
}

func (d *ServerDispatcher) WatchRegistry(events chan<- flux.EndpointEvent) error {
	return d.activeRegistry.WatchEvents(events)
}

func (d *ServerDispatcher) Startup() error {
	for _, startup := range sortedStartup(ext.GetStartupHooks()) {
		if err := startup.Startup(); nil != err {
			return err
		}
	}
	return nil
}

func (d *ServerDispatcher) Shutdown(ctx context.Context) error {
	for _, shutdown := range sortedShutdown(ext.GetShutdownHooks()) {
		if err := shutdown.Shutdown(ctx); nil != err {
			return err
		}
	}
	return nil
}

func (d *ServerDispatcher) Dispatch(ctx flux.Context) *flux.InvokeError {
	globalFilters := ext.GlobalFilters()
	selectFilters := make([]flux.Filter, 0)
	for _, selector := range ext.FindSelectors(ctx.RequestHost()) {
		for _, typeId := range selector.Select(ctx).Filters {
			if f, ok := ext.GetSelectiveFilter(typeId); ok {
				selectFilters = append(selectFilters, f)
			} else {
				logger.Trace(ctx.RequestId()).Warnw("Filter not found on selector", "type-id", typeId)
			}
		}
	}
	// Metrics
	metrics := make(map[string]string, len(globalFilters)+len(selectFilters)+1)
	defer func() {
		for k, v := range metrics {
			ctx.ResponseWriter().AddHeader(k, v)
		}
	}()
	return d.walk(metrics, func(ctx flux.Context) *flux.InvokeError {
		protoName := ctx.Endpoint().Protocol
		if exchange, ok := ext.GetExchange(protoName); !ok {
			return &flux.InvokeError{
				StatusCode: flux.StatusNotFound,
				ErrorCode:  flux.ErrorCodeRequestNotFound,
				Message:    fmt.Sprintf("ROUTE:UNKNOWN_PROTOCOL: %s", protoName)}
		} else {
			start := time.Now()
			ret := exchange.Exchange(ctx)
			metrics["X-Metric-Exchange"] = time.Since(start).String()
			return ret
		}
	}, append(globalFilters, selectFilters...)...)(ctx)
}

func (d *ServerDispatcher) walk(metrics map[string]string, next flux.FilterInvoker, filters ...flux.Filter) flux.FilterInvoker {
	for i := len(filters) - 1; i >= 0; i-- {
		start := time.Now()
		next = filters[i].Invoke(next)
		metrics["X-Metric-"+filters[i].TypeId()] = time.Since(start).String()
	}
	return next
}

func _isDisabled(config *flux.Configuration) bool {
	return config.GetBool("disable") || config.GetBool("disabled")
}

func findActiveRegistry() (flux.Registry, *flux.Configuration, error) {
	config := flux.NewConfigurationOf(flux.KeyConfigRootRegistry)
	config.SetDefault(flux.KeyConfigRegistryId, ext.RegistryIdDefault)
	registryId := config.GetString(flux.KeyConfigRegistryId)
	logger.Infow("Active metadata registry", "registry-id", registryId)
	if factory, ok := ext.GetRegistryFactory(registryId); !ok {
		return nil, config, fmt.Errorf("RegistryFactory not found, id: %s", registryId)
	} else {
		return factory(), config, nil
	}
}
