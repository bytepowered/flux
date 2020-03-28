package internal

import (
	"context"
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/pkg"
	"strings"
	"time"
)

type Dispatcher struct {
	activeRegistry flux.Registry
	hooksStartup   []flux.Startuper
	hooksShutdown  []flux.Shutdowner
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		hooksStartup:  make([]flux.Startuper, 0),
		hooksShutdown: make([]flux.Shutdowner, 0),
	}
}

func (d *Dispatcher) Init(globals flux.Config) error {
	logger.Infof("Dispatcher initialing")

	// 组件需要注册生命周期回调钩子
	doRegisterHooks := func(ref interface{}, config flux.Config) error {
		if init, ok := ref.(flux.Initializer); ok {
			if err := init.Init(config); nil != err {
				return err
			}
		}
		if startup, ok := ref.(flux.Startuper); ok {
			d.hooksStartup = append(d.hooksStartup, startup)
		}
		if shutdown, ok := ref.(flux.Shutdowner); ok {
			d.hooksShutdown = append(d.hooksShutdown, shutdown)
		}
		return nil
	}
	// 从配置中动态加载的组件
	for _, item := range dynloadConfig(globals) {
		ref := item.Factory()
		logger.Infof("Load component, name: %s, type: %s, inst.type: %T", item.Name, item.Type, ref)
		if err := doRegisterHooks(ref, item.Config); nil != err {
			return err
		}
	}

	// 静态注册的内核组件

	// Registry
	registryConfig := ext.ConfigFactory()("flux.registry", globals.Map(flux.KeyConfigRootRegistry))
	if activeRegistry, err := registryActiveWith(registryConfig); nil != err {
		return err
	} else {
		d.activeRegistry = activeRegistry
		if err := doRegisterHooks(activeRegistry, registryConfig); nil != err {
			return err
		}
	}
	// Exchanges
	exchangeConfig := ext.ConfigFactory()("flux.exchanges", globals.Map(flux.KeyConfigRootExchanges))
	for proto, ex := range ext.Exchanges() {
		logger.Infof("Load exchange, proto: %s, inst.type: %T", proto, ex)
		protoConfig := ext.ConfigFactory()("flux.exchanges.proto."+proto,
			exchangeConfig.Map(strings.ToUpper(proto)))
		if err := doRegisterHooks(ex, protoConfig); nil != err {
			return err
		}
	}
	// GlobalFilters
	for i, filter := range ext.GlobalFilters() {
		logger.Infof("Load global filter, order: %d, filter.type: %T", i, filter)
		factory := ext.ConfigFactory()
		filterConfig := factory("flux.component."+filter.Id(), make(map[string]interface{}))
		if err := doRegisterHooks(filter, filterConfig); nil != err {
			return err
		}
	}
	return nil
}

func (d *Dispatcher) WatchRegistry(events chan<- flux.EndpointEvent) error {
	// Debug echo registry
	if pkg.IsEnv(pkg.EnvDev) {
		if f, ok := ext.GetRegistryFactory(ext.RegistryIdEcho); ok {
			go func() { pkg.Silently(f().WatchEvents(events)) }()
		}
	}
	return d.activeRegistry.WatchEvents(events)
}

func (d *Dispatcher) Startup() error {
	for _, startup := range sortedStartup(d.hooksStartup) {
		if err := startup.Startup(); nil != err {
			return err
		}
	}
	return nil
}

func (d *Dispatcher) Shutdown(ctx context.Context) error {
	for _, shutdown := range sortedShutdown(d.hooksShutdown) {
		if err := shutdown.Shutdown(ctx); nil != err {
			return err
		}
	}
	return nil
}

func (d *Dispatcher) Dispatch(ctx flux.Context) *flux.InvokeError {
	globalFilters := ext.GlobalFilters()
	selectFilters := make([]flux.Filter, 0)
	for _, selector := range ext.FindSelectors(ctx.RequestHost()) {
		for _, id := range selector.Select(ctx).Filters {
			if f, ok := ext.GetFilter(id); ok {
				selectFilters = append(selectFilters, f)
			} else {
				logger.Warnf("Filter not found on selector, filterId: %s", id)
			}
		}
	}
	return d.walk(func(ctx flux.Context) *flux.InvokeError {
		protoName := ctx.Endpoint().Protocol
		if exchange, ok := ext.GetExchange(protoName); !ok {
			return &flux.InvokeError{
				StatusCode: flux.StatusNotFound,
				Message:    fmt.Sprintf("ROUTE:UNKNOWN_PROTOCOL: %s", protoName)}
		} else {
			start := time.Now()
			ret := exchange.Exchange(ctx)
			elapsed := time.Now().Sub(start)
			ctx.ResponseWriter().AddHeader("X-Exchange-Elapsed", elapsed.String())
			return ret
		}
	}, append(globalFilters, selectFilters...)...)(ctx)
}

func (d *Dispatcher) walk(fi flux.FilterInvoker, filters ...flux.Filter) flux.FilterInvoker {
	for i := len(filters) - 1; i >= 0; i-- {
		fi = filters[i].Invoke(fi)
	}
	return fi
}

func registryActiveWith(config flux.Config) (flux.Registry, error) {
	registryId := config.StringOrDefault(flux.KeyConfigRegistryId, ext.RegistryIdDefault)
	logger.Infof("Active activeRegistry, id: %s", registryId)
	if factory, ok := ext.GetRegistryFactory(registryId); !ok {
		return nil, fmt.Errorf("RegistryFactory not found, id: %s", registryId)
	} else {
		return factory(), nil
	}
}
