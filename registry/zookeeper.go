package registry

import (
	"context"
	"errors"
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/remoting"
	"github.com/bytepowered/flux/remoting/zk"
	"time"
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
	endpointRootPath string
	endpointEvents   chan flux.HttpEndpointEvent
	serviceRootPath  string
	serviceEvents    chan flux.BackendServiceEvent
	retriever        *zk.ZookeeperRetriever
}

func ZkEndpointRegistryFactory() flux.EndpointRegistry {
	return &ZookeeperMetadataRegistry{
		retriever:      zk.NewZookeeperRetriever(),
		endpointEvents: make(chan flux.HttpEndpointEvent, 4),
	}
}

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
	r.endpointRootPath = config.GetString("endpoint-path")
	r.serviceRootPath = config.GetString("service-path")
	if r.endpointRootPath == "" || r.serviceRootPath == "" {
		return errors.New("config(endpoint-path, service-path) is empty")
	} else {
		return r.retriever.Init(config)
	}
}

func (r *ZookeeperMetadataRegistry) WatchHttpEndpoints() (<-chan flux.HttpEndpointEvent, error) {
	listener := func(event remoting.NodeEvent) {
		defer func() {
			if r := recover(); nil != r {
				logger.Errorw("Zookeeper node listening", "event", event, "error", r)
			}
		}()
		if evt, ok := toEndpointEvent(event.Data, event.EventType); ok {
			r.endpointEvents <- evt
		}
	}
	if err := r.watch(r.endpointRootPath, listener); err != nil {
		return nil, err
	} else {
		return r.endpointEvents, err
	}
}

func (r *ZookeeperMetadataRegistry) WatchBackendServices() (<-chan flux.BackendServiceEvent, error) {
	listener := func(event remoting.NodeEvent) {
		defer func() {
			if r := recover(); nil != r {
				logger.Errorw("Zookeeper node listening", "event", event, "error", r)
			}
		}()
		if evt, ok := toBackendServiceEvent(event.Data, event.EventType); ok {
			r.serviceEvents <- evt
		}
	}
	if err := r.watch(r.serviceRootPath, listener); err != nil {
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

func (r *ZookeeperMetadataRegistry) Startup() error {
	logger.Info("Startup registry")
	return r.retriever.Startup()
}

func (r *ZookeeperMetadataRegistry) Shutdown(ctx context.Context) error {
	logger.Info("Shutdown registry")
	close(r.endpointEvents)
	return r.retriever.Shutdown(ctx)
}
