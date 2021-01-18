package registry

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bytepowered/flux"
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
	_ flux.EndpointRegistry = new(ZookeeperMetadataRegistry)
)

// ZookeeperMetadataRegistry 基于ZK节点树实现的Endpoint元数据注册中心
type ZookeeperMetadataRegistry struct {
	globalAlias    map[string]string
	endpointPath   string
	endpointEvents chan flux.HttpEndpointEvent
	servicePath    string
	serviceEvents  chan flux.BackendServiceEvent
	retriever      *zk.ZookeeperRetriever
}

// ZkEndpointRegistryFactory Factory func to new a zookeeper registry
func ZkEndpointRegistryFactory() flux.EndpointRegistry {
	return &ZookeeperMetadataRegistry{
		retriever:      zk.NewZookeeperRetriever(),
		endpointEvents: make(chan flux.HttpEndpointEvent, 4),
		serviceEvents:  make(chan flux.BackendServiceEvent, 4),
	}
}

// ZkEndpointRegistryFactory Factory func to new a zookeeper registry
func ZkEndpointRegistryFactoryWith(globalAlias map[string]string) flux.EndpointRegistry {
	return &ZookeeperMetadataRegistry{
		globalAlias:    globalAlias,
		retriever:      zk.NewZookeeperRetriever(),
		endpointEvents: make(chan flux.HttpEndpointEvent, 4),
		serviceEvents:  make(chan flux.BackendServiceEvent, 4),
	}
}

// Init init registry
func (r *ZookeeperMetadataRegistry) Init(config *flux.Configuration) error {
	config.SetDefaults(map[string]interface{}{
		"endpoint-path": zkRegistryHttpEndpointPath,
		"service-path":  zkRegistryBackendServicePath,
		"timeout":       time.Second * 10,
	})
	config.SetGlobalAlias(map[string]string{
		"host":     "zookeeper.host",
		"port":     "zookeeper.port",
		"address":  "zookeeper.address",
		"password": "zookeeper.password",
		"database": "zookeeper.database",
	})
	if len(r.globalAlias) != 0 {
		config.SetGlobalAlias(r.globalAlias)
	}
	r.endpointPath = config.GetString("endpoint-path")
	r.servicePath = config.GetString("service-path")
	if r.endpointPath == "" || r.servicePath == "" {
		return errors.New("config(endpoint-path, service-path) is empty")
	} else {
		return r.retriever.Init(config)
	}
}

// WatchHttpEndpoints Listen http endpoints events
func (r *ZookeeperMetadataRegistry) WatchHttpEndpoints() (<-chan flux.HttpEndpointEvent, error) {
	listener := func(event remoting.NodeEvent) {
		defer func() {
			if r := recover(); nil != r {
				logger.Errorw("Zookeeper node listening", "event", event, "error", r)
			}
		}()
		if evt, ok := NewEndpointEvent(event.Data, event.EventType); ok {
			r.endpointEvents <- evt
		}
	}
	logger.Infow("Zookeeper start listen endpoints node", "node-path", r.endpointPath)
	if err := r.watch(r.endpointPath, listener); err != nil {
		return nil, err
	} else {
		return r.endpointEvents, err
	}
}

// WatchBackendServices Listen gateway services events
func (r *ZookeeperMetadataRegistry) WatchBackendServices() (<-chan flux.BackendServiceEvent, error) {
	listener := func(event remoting.NodeEvent) {
		defer func() {
			if r := recover(); nil != r {
				logger.Errorw("Zookeeper node listening", "event", event, "error", r)
			}
		}()
		if evt, ok := NewBackendServiceEvent(event.Data, event.EventType); ok {
			r.serviceEvents <- evt
		}
	}
	logger.Infow("Zookeeper start listen services node", "node-path", r.servicePath)
	if err := r.watch(r.servicePath, listener); err != nil {
		return nil, err
	} else {
		return r.serviceEvents, err
	}
}

func (r *ZookeeperMetadataRegistry) watch(rootpath string, nodeListener func(remoting.NodeEvent)) error {
	if exist, _ := r.retriever.Exists(rootpath); !exist {
		if err := r.retriever.Create(rootpath); nil != err {
			return fmt.Errorf("init metadata node: %w", err)
		}
	}
	logger.Infow("Zookeeper watching metadata node", "path", rootpath)
	return r.retriever.AddChildrenNodeChangedListener("", rootpath, func(event remoting.NodeEvent) {
		logger.Infow("Receive child change event", "event", event)
		if event.EventType == remoting.EventTypeChildAdd {
			if err := r.retriever.AddNodeChangedListener("", event.Path, nodeListener); nil != err {
				logger.Warnw("Watch child node data", "error", err)
			}
		}
	})
}

// Startup Startup registry
func (r *ZookeeperMetadataRegistry) Startup() error {
	logger.Info("Startup registry")
	return r.retriever.Startup()
}

// Shutdown Startup registry
func (r *ZookeeperMetadataRegistry) Shutdown(ctx context.Context) error {
	logger.Info("Shutdown registry")
	close(r.endpointEvents)
	return r.retriever.Shutdown(ctx)
}
