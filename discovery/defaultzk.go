package discovery

import (
	"context"
	"errors"
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/remoting"
	"github.com/bytepowered/flux/remoting/zk"
)

const (
	// 在ZK注册的根节点。需要与客户端的注册保持一致。
	zkDiscoveryHttpEndpointPath   = "/flux-endpoint"
	zkDiscoveryBackendServicePath = "/flux-service"
)

var (
	_ flux.EndpointDiscovery = new(DefaultDiscovery)
)

type (
	// Option 配置函数
	Option func(discovery *DefaultDiscovery)
)

// DefaultDiscovery 基于ZK节点树实现的Endpoint元数据注册中心
type DefaultDiscovery struct {
	globalAlias    map[string]string
	endpointPath   string
	servicePath    string
	endpointEvents chan flux.HttpEndpointEvent
	serviceEvents  chan flux.BackendServiceEvent
	retrievers     []*zk.ZookeeperRetriever
}

// WithConfigAlias 配置注册中心的配置别名
func WithConfigAlias(alias map[string]string) Option {
	return func(discovery *DefaultDiscovery) {
		discovery.globalAlias = alias
	}
}

// DefaultDiscoveryFactory Factory func to new a zookeeper discovery
func DefaultDiscoveryFactory() flux.EndpointDiscovery {
	return NewDefaultDiscoveryFactoryWith()()
}

// NewDefaultDiscoveryFactoryWith returns new a zookeeper discovery factory
func NewDefaultDiscoveryFactoryWith(opts ...Option) ext.EndpointDiscoveryFactory {
	return func() flux.EndpointDiscovery {
		r := &DefaultDiscovery{
			endpointEvents: make(chan flux.HttpEndpointEvent, 4),
			serviceEvents:  make(chan flux.BackendServiceEvent, 4),
		}
		for _, opt := range opts {
			opt(r)
		}
		return r
	}
}

// Init init discovery
func (r *DefaultDiscovery) Init(config *flux.Configuration) error {
	config.SetDefaults(map[string]interface{}{
		"endpoint-path": zkDiscoveryHttpEndpointPath,
		"service-path":  zkDiscoveryBackendServicePath,
	})
	active := config.GetStringSlice("discovery-active")
	if len(active) == 0 {
		active = []string{"default"}
	}
	logger.Infow("DefaultZkDiscovery active discovery", "active-ids", active)
	r.endpointPath = config.GetString("endpoint-path")
	r.servicePath = config.GetString("service-path")
	if r.endpointPath == "" || r.servicePath == "" {
		return errors.New("config(endpoint-path, service-path) is empty")
	}
	r.retrievers = make([]*zk.ZookeeperRetriever, len(active))
	for i := range active {
		id := active[i]
		r.retrievers[i] = zk.NewZookeeperRetriever(id)
		zkconf := config.Sub(id)
		zkconf.SetGlobalAlias(map[string]string{
			"address":  "zookeeper.address",
			"password": "zookeeper.password",
			"timeout":  "zookeeper.timeout",
		})
		if len(r.globalAlias) != 0 {
			zkconf.SetGlobalAlias(r.globalAlias)
		}
		logger.Infow("DefaultZkDiscovery start zk discovery", "discovery-id", id)
		if err := r.retrievers[i].Init(zkconf); nil != err {
			return err
		}
	}
	return nil
}

// OnEndpointChanged Listen http endpoints events
func (r *DefaultDiscovery) OnEndpointChanged() (<-chan flux.HttpEndpointEvent, error) {
	listener := func(event remoting.NodeEvent) {
		defer func() {
			if r := recover(); nil != r {
				logger.Errorw("DefaultZkDiscovery node listening", "event", event, "error", r)
			}
		}()
		if evt, ok := NewEndpointEvent(event.Data, event.EventType); ok {
			r.endpointEvents <- evt
		}
	}
	logger.Infow("DefaultZkDiscovery start listen endpoints node", "node-path", r.endpointPath)
	for _, retriever := range r.retrievers {
		if err := r.watch(retriever, r.endpointPath, listener); err != nil {
			return nil, err
		}
	}
	return r.endpointEvents, nil
}

// OnServiceChanged Listen gateway services events
func (r *DefaultDiscovery) OnServiceChanged() (<-chan flux.BackendServiceEvent, error) {
	listener := func(event remoting.NodeEvent) {
		defer func() {
			if r := recover(); nil != r {
				logger.Errorw("DefaultZkDiscovery node listening", "event", event, "error", r)
			}
		}()
		if evt, ok := NewBackendServiceEvent(event.Data, event.EventType); ok {
			r.serviceEvents <- evt
		}
	}
	logger.Infow("DefaultZkDiscovery start listen services node", "node-path", r.servicePath)
	for _, retriever := range r.retrievers {
		if err := r.watch(retriever, r.servicePath, listener); err != nil {
			return nil, err
		}
	}
	return r.serviceEvents, nil
}

func (r *DefaultDiscovery) watch(retriever *zk.ZookeeperRetriever, rootpath string, nodeListener func(remoting.NodeEvent)) error {
	if exist, _ := retriever.Exists(rootpath); !exist {
		if err := retriever.Create(rootpath); nil != err {
			return fmt.Errorf("init metadata node: %w", err)
		}
	}
	logger.Infow("DefaultZkDiscovery watching metadata node", "path", rootpath)
	return retriever.AddChildrenNodeChangedListener("", rootpath, func(event remoting.NodeEvent) {
		logger.Infow("DefaultZkDiscovery receive child change event", "event", event)
		if event.EventType == remoting.EventTypeChildAdd {
			if err := retriever.AddNodeChangedListener("", event.Path, nodeListener); nil != err {
				logger.Warnw("Watch child node data", "error", err)
			}
		}
	})
}

// Startup startup discovery service
func (r *DefaultDiscovery) Startup() error {
	logger.Info("DefaultZkDiscovery startup")
	for _, retriever := range r.retrievers {
		if err := retriever.Startup(); nil != err {
			return err
		}
	}
	return nil
}

// Shutdown shutdown discovery service
func (r *DefaultDiscovery) Shutdown(ctx context.Context) error {
	logger.Info("DefaultZkDiscovery shutdown")
	close(r.endpointEvents)
	for _, retriever := range r.retrievers {
		if err := retriever.Shutdown(ctx); nil != err {
			return err
		}
	}
	return nil
}
