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

type FxDispatcher struct {
	activeRegistry flux.Registry
	hooksStartup   []flux.Startuper
	hooksShutdown  []flux.Shutdowner
}

func NewDispatcher() *FxDispatcher {
	return &FxDispatcher{
		hooksStartup:  make([]flux.Startuper, 0),
		hooksShutdown: make([]flux.Shutdowner, 0),
	}
}

func (d *FxDispatcher) Init(globals flux.Configuration) error {
	logger.Infof("Dispatcher initialing")
	// 组件生命周期回调钩子
	initRegisterHook := func(ref interface{}, config flux.Configuration) error {
		if init, ok := ref.(flux.Initializer); ok {
			if err := init.Init(config); nil != err {
				return err
			}
		}
		d.AddLifecycleHook(ref)
		return nil
	}
	// 静态注册的单实例内核组件
	// Registry
	registryConfig := ext.ConfigurationFactory()(configNsPrefixRegistry, globals.Map(flux.KeyConfigRootRegistry))
	if activeRegistry, err := registryActiveWith(registryConfig); nil != err {
		return err
	} else {
		d.activeRegistry = activeRegistry
		if err := initRegisterHook(activeRegistry, registryConfig); nil != err {
			return err
		}
	}
	// Exchanges
	exchangeConfig := flux.NewMapConfiguration(globals.Map(flux.KeyConfigRootExchanges))
	for proto, ex := range ext.Exchanges() {
		ns := configNsPrefixExchangeProto + proto
		logger.Infof("Load exchange, proto: %s, inst.type: %T, config.ns: %s", proto, ex, ns)
		protoConfig := ext.ConfigurationFactory()(ns, exchangeConfig.Map(strings.ToUpper(proto)))
		if err := initRegisterHook(ex, protoConfig); nil != err {
			return err
		}
	}
	// Filters
	for _, filter := range append(ext.GlobalFilters(), ext.ScopedFilters()...) {
		ns := configNsPrefixComponent + filter.TypeId()
		filterConfig := ext.ConfigurationFactory()(ns, globals.Map(filter.TypeId()))
		logger.Infof("Load filter, filter.type: %T, config.ns: %s", filter, ns)
		if err := initRegisterHook(filter, filterConfig); nil != err {
			return err
		}
	}
	// 加载和注册，动态多实例组件
	for _, item := range dynloadConfig(globals) {
		aware := item.Factory()
		logger.Infof("Load aware, name: %s, type: %s, aware.type: %T, config.ns: %s", item.Name, item.TypeId, aware, item.ConfigNs)
		if err := initRegisterHook(aware, item.Config); nil != err {
			return err
		}
		// 目前只支持Filter动态注册
		// 其它未知组件，只做动态启动生命周期
		if filter, ok := aware.(flux.Filter); ok {
			ext.AddFilter(filter)
		}
	}
	return nil
}

func (d *FxDispatcher) AddLifecycleHook(hook interface{}) {
	if startup, ok := hook.(flux.Startuper); ok {
		d.hooksStartup = append(d.hooksStartup, startup)
	}
	if shutdown, ok := hook.(flux.Shutdowner); ok {
		d.hooksShutdown = append(d.hooksShutdown, shutdown)
	}
}

func (d *FxDispatcher) WatchRegistry(events chan<- flux.EndpointEvent) error {
	// Debug echo registry
	if pkg.IsEnv(pkg.EnvDev) {
		if f, ok := ext.GetRegistryFactory(ext.RegistryIdEcho); ok {
			go func() { pkg.Silently(f().WatchEvents(events)) }()
		}
	}
	return d.activeRegistry.WatchEvents(events)
}

func (d *FxDispatcher) Startup() error {
	for _, startup := range sortedStartup(d.hooksStartup) {
		if err := startup.Startup(); nil != err {
			return err
		}
	}
	return nil
}

func (d *FxDispatcher) Shutdown(ctx context.Context) error {
	for _, shutdown := range sortedShutdown(d.hooksShutdown) {
		if err := shutdown.Shutdown(ctx); nil != err {
			return err
		}
	}
	return nil
}

func (d *FxDispatcher) Dispatch(ctx flux.Context) *flux.InvokeError {
	globalFilters := ext.GlobalFilters()
	selectFilters := make([]flux.Filter, 0)
	for _, selector := range ext.FindSelectors(ctx.RequestHost()) {
		for _, typeId := range selector.Select(ctx).Filters {
			if f, ok := ext.GetFilter(typeId); ok {
				selectFilters = append(selectFilters, f)
			} else {
				logger.Warnf("Filter not found on selector, filter.typeId: %s", typeId)
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

func (d *FxDispatcher) walk(fi flux.FilterInvoker, filters ...flux.Filter) flux.FilterInvoker {
	for i := len(filters) - 1; i >= 0; i-- {
		fi = filters[i].Invoke(fi)
	}
	return fi
}

func registryActiveWith(config flux.Configuration) (flux.Registry, error) {
	registryId := config.StringOrDefault(flux.KeyConfigRegistryId, ext.RegistryIdDefault)
	logger.Infof("Active registry, id: %s", registryId)
	if factory, ok := ext.GetRegistryFactory(registryId); !ok {
		return nil, fmt.Errorf("RegistryFactory not found, id: %s", registryId)
	} else {
		return factory(), nil
	}
}
