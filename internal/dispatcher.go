package internal

import (
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/extension"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/pkg"
	"os"
	"strings"
	"time"
)

type Dispatcher struct {
	registry  flux.Registry
	startups  []flux.Startuper
	shutdowns []flux.Shutdowner
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{}
}

func (d *Dispatcher) Init(globals flux.Config) error {
	logger.Infof("Dispatcher initialing")
	register := func(ref interface{}, config flux.Config) error {
		if init, ok := ref.(flux.Initializer); ok {
			if err := init.Init(config); nil != err {
				return err
			}
		}
		if startup, ok := ref.(flux.Startuper); ok {
			d.startups = append(d.startups, startup)
		}
		if shutdown, ok := ref.(flux.Shutdowner); ok {
			d.shutdowns = append(d.shutdowns, shutdown)
		}
		return nil
	}
	// 从配置中动态加载的组件
	for _, item := range dynloadConfig(globals) {
		ref := item.Factory()
		logger.Infof("Load component, name: %s, type: %s, inst.type: %T", item.Name, item.Type, ref)
		if err := register(ref, item.Config); nil != err {
			return err
		}
	}

	// 静态注册的内核组件

	// Registry
	registryConfig := globals.Config(flux.KeyConfigRootRegistry)
	if registry, err := activeRegistry(registryConfig); nil != err {
		return err
	} else {
		d.registry = registry
		if err := register(registry, registryConfig); nil != err {
			return err
		}
	}
	// Exchanges
	exchangeConfig := globals.Config("Exchanges")
	for proto, ex := range extension.Exchanges() {
		logger.Infof("Load exchange, proto: %s, inst.type: %T", proto, ex)
		if err := register(ex, exchangeConfig.Config(strings.ToUpper(proto))); nil != err {
			return err
		}
	}
	// GlobalFilters
	for i, gf := range extension.GetGlobalFilter() {
		logger.Infof("Load global filter, idx: %d, inst.type: %T", i, gf)
		if err := register(gf, map[string]interface{}{}); nil != err {
			return err
		}
	}
	return nil
}

func (d *Dispatcher) WatchRegistry(events chan<- flux.EndpointEvent) error {
	// Debug echo registry
	if "dev" == os.Getenv("runtime.env") {
		if f, ok := extension.GetRegistryFactory(extension.TypeNameRegistryEcho); ok {
			go func() { pkg.Silently(f().WatchEvents(events)) }()
		}
	}
	return d.registry.WatchEvents(events)
}

func (d *Dispatcher) Startup() error {
	for _, startup := range sortedStartup(d.startups) {
		if err := startup.Startup(); nil != err {
			return err
		}
	}
	return nil
}

func (d *Dispatcher) Shutdown() error {
	for _, shutdown := range sortedShutdown(d.shutdowns) {
		if err := shutdown.Shutdown(); nil != err {
			return err
		}
	}
	return nil
}

func (d *Dispatcher) Dispatch(ctx flux.Context) *flux.InvokeError {
	globalFilter := extension.GetGlobalFilter()
	// TODO Select Context Filter
	return d.walk(func(ctx flux.Context) *flux.InvokeError {
		protocol := ctx.Endpoint().Protocol
		if exchange, ok := extension.GetExchange(protocol); !ok {
			return &flux.InvokeError{
				StatusCode: flux.StatusNotFound,
				Message:    fmt.Sprintf("ROUTE:UNKNOWN_PROTOCOL: %s", protocol)}
		} else {
			start := time.Now()
			ret := exchange.Exchange(ctx)
			elapsed := time.Now().Sub(start)
			ctx.ResponseWriter().AddHeader("X-Exchange-Elapsed", elapsed.String())
			return ret
		}
	}, globalFilter...)(ctx)
}

func (d *Dispatcher) walk(fi flux.FilterInvoker, filters ...flux.Filter) flux.FilterInvoker {
	for i := len(filters) - 1; i >= 0; i-- {
		fi = filters[i].Invoke(fi)
	}
	return fi
}

func activeRegistry(config flux.Config) (flux.Registry, error) {
	registry := config.StringOrDefault(flux.KeyConfigRegistryProtocol, extension.TypeNameRegistryActive)
	logger.Infof("Active endpoint registry: %s", registry)
	if factory, ok := extension.GetRegistryFactory(registry); !ok {
		return nil, fmt.Errorf("RegistryFactory not found, name: %s", registry)
	} else {
		return factory(), nil
	}
}
