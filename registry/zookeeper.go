package registry

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
	zkRegistryHttpEndpointPath   = "/flux-endpoint"
	zkRegistryBackendServicePath = "/flux-service"
)

var (
	_ flux.EndpointRegistry = new(ZkEndpointRegistry)
)

// ZkEndpointRegistry 基于ZK节点树实现的Endpoint元数据注册中心
type ZkEndpointRegistry struct {
	globalAlias    map[string]string
	endpointPath   string
	endpointEvents chan flux.HttpEndpointEvent
	servicePath    string
	serviceEvents  chan flux.BackendServiceEvent
	retrievers     []*zk.ZookeeperRetriever
}

// ZkEndpointRegistryFactory Factory func to new a zookeeper registry
func ZkEndpointRegistryFactory() flux.EndpointRegistry {
	return &ZkEndpointRegistry{
		endpointEvents: make(chan flux.HttpEndpointEvent, 4),
		serviceEvents:  make(chan flux.BackendServiceEvent, 4),
	}
}

// NewZkEndpointRegistryFactoryWith returns new a zookeeper registry factory
func NewZkEndpointRegistryFactoryWith(globalAlias map[string]string) ext.EndpointRegistryFactory {
	return func() flux.EndpointRegistry {
		return &ZkEndpointRegistry{
			globalAlias:    globalAlias,
			endpointEvents: make(chan flux.HttpEndpointEvent, 4),
			serviceEvents:  make(chan flux.BackendServiceEvent, 4),
		}
	}
}

// Init init registry
func (r *ZkEndpointRegistry) Init(config *flux.Configuration) error {
	config.SetDefaults(map[string]interface{}{
		"endpoint-path": zkRegistryHttpEndpointPath,
		"service-path":  zkRegistryBackendServicePath,
	})
	active := config.GetStringSlice("registry-active")
	if len(active) == 0 {
		active = []string{"default"}
	}
	logger.Infow("ZookeeperRegistry active registry", "active-ids", active)
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
		logger.Infow("ZookeeperRegistry start zk registry", "registry-id", id)
		if err := r.retrievers[i].Init(zkconf); nil != err {
			return err
		}
	}
	return nil
}

// WatchHttpEndpoints Listen http endpoints events
func (r *ZkEndpointRegistry) WatchHttpEndpoints() (<-chan flux.HttpEndpointEvent, error) {
	listener := func(event remoting.NodeEvent) {
		defer func() {
			if r := recover(); nil != r {
				logger.Errorw("ZookeeperRegistry node listening", "event", event, "error", r)
			}
		}()
		if evt, ok := NewEndpointEvent(event.Data, event.EventType); ok {
			r.endpointEvents <- evt
		}
	}
	logger.Infow("ZookeeperRegistry start listen endpoints node", "node-path", r.endpointPath)
	for _, retriever := range r.retrievers {
		if err := r.watch(retriever, r.endpointPath, listener); err != nil {
			return nil, err
		}
	}
	return r.endpointEvents, nil
}

// WatchBackendServices Listen gateway services events
func (r *ZkEndpointRegistry) WatchBackendServices() (<-chan flux.BackendServiceEvent, error) {
	listener := func(event remoting.NodeEvent) {
		defer func() {
			if r := recover(); nil != r {
				logger.Errorw("ZookeeperRegistry node listening", "event", event, "error", r)
			}
		}()
		if evt, ok := NewBackendServiceEvent(event.Data, event.EventType); ok {
			r.serviceEvents <- evt
		}
	}
	logger.Infow("ZookeeperRegistry start listen services node", "node-path", r.servicePath)
	for _, retriever := range r.retrievers {
		if err := r.watch(retriever, r.servicePath, listener); err != nil {
			return nil, err
		}
	}
	return r.serviceEvents, nil
}

func (r *ZkEndpointRegistry) watch(retriever *zk.ZookeeperRetriever, rootpath string, nodeListener func(remoting.NodeEvent)) error {
	if exist, _ := retriever.Exists(rootpath); !exist {
		if err := retriever.Create(rootpath); nil != err {
			return fmt.Errorf("init metadata node: %w", err)
		}
	}
	logger.Infow("ZookeeperRegistry watching metadata node", "path", rootpath)
	return retriever.AddChildrenNodeChangedListener("", rootpath, func(event remoting.NodeEvent) {
		logger.Infow("ZookeeperRegistry receive child change event", "event", event)
		if event.EventType == remoting.EventTypeChildAdd {
			if err := retriever.AddNodeChangedListener("", event.Path, nodeListener); nil != err {
				logger.Warnw("Watch child node data", "error", err)
			}
		}
	})
}

// Startup Startup registry
func (r *ZkEndpointRegistry) Startup() error {
	logger.Info("ZookeeperRegistry startup")
	for _, retriever := range r.retrievers {
		if err := retriever.Startup(); nil != err {
			return err
		}
	}
	return nil
}

// Shutdown Startup registry
func (r *ZkEndpointRegistry) Shutdown(ctx context.Context) error {
	logger.Info("ZookeeperRegistry shutdown")
	close(r.endpointEvents)
	for _, retriever := range r.retrievers {
		if err := retriever.Shutdown(ctx); nil != err {
			return err
		}
	}
	return nil
}
