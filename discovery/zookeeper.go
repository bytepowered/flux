package discovery

import (
	"context"
	"errors"
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/remoting"
	"github.com/bytepowered/flux/remoting/zk"
)

const (
	// 在ZK注册的根节点。需要与客户端的注册保持一致。
	zkDiscoveryHttpEndpointPath   = "/flux-endpoint"
	zkDiscoveryBackendServicePath = "/flux-service"
)

const (
	ZookeeperId = "zookeeper"
)

const (
	zkConfigRootpathEndpoint = "rootpath_endpoint"
	zkConfigRootpathService  = "rootpath_service"
	zkConfigRegistrySelector = "registry_selector"
)

var _ flux.EndpointDiscovery = new(ZookeeperDiscoveryService)

type (
	// ZookeeperOption 配置函数
	ZookeeperOption func(discovery *ZookeeperDiscoveryService)
)

// ZookeeperDiscoveryService 基于ZK节点树实现的Endpoint元数据注册中心
type ZookeeperDiscoveryService struct {
	id           string
	globalAlias  map[string]string
	endpointPath string
	servicePath  string
	retrievers   []*zk.ZookeeperRetriever
}

// WithGlobalAlias 配置注册中心的配置别名
func WithGlobalAlias(alias map[string]string) ZookeeperOption {
	return func(discovery *ZookeeperDiscoveryService) {
		discovery.globalAlias = alias
	}
}

// NewZookeeperServiceWith returns new a zookeeper discovery factory
func NewZookeeperServiceWith(id string, opts ...ZookeeperOption) *ZookeeperDiscoveryService {
	r := &ZookeeperDiscoveryService{
		id: id,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *ZookeeperDiscoveryService) Id() string {
	return r.id
}

// Init init discovery
func (r *ZookeeperDiscoveryService) Init(config *flux.Configuration) error {
	config.SetDefaults(map[string]interface{}{
		zkConfigRootpathEndpoint: zkDiscoveryHttpEndpointPath,
		zkConfigRootpathService:  zkDiscoveryBackendServicePath,
	})
	selected := config.GetStringSlice(zkConfigRegistrySelector)
	if len(selected) == 0 {
		selected = []string{"default"}
	}
	logger.Infow("ZkEndpointDiscovery selected discovery", "selected-ids", selected)
	r.endpointPath = config.GetString(zkConfigRootpathEndpoint)
	r.servicePath = config.GetString(zkConfigRootpathService)
	if r.endpointPath == "" || r.servicePath == "" {
		return errors.New("config(rootpath_endpoint, rootpath_service) is empty")
	}
	r.retrievers = make([]*zk.ZookeeperRetriever, len(selected))
	registries := config.Sub("registry_centers")
	for i := range selected {
		id := selected[i]
		r.retrievers[i] = zk.NewZookeeperRetriever(id)
		zkconf := registries.Sub(id)
		zkconf.SetGlobalAlias(map[string]string{
			"address":  "zookeeper.address",
			"password": "zookeeper.password",
			"timeout":  "zookeeper.timeout",
		})
		if len(r.globalAlias) != 0 {
			zkconf.SetGlobalAlias(r.globalAlias)
		}
		logger.Infow("ZkEndpointDiscovery start zk discovery", "discovery-id", id)
		if err := r.retrievers[i].Init(zkconf); nil != err {
			return err
		}
	}
	return nil
}

// OnEndpointChanged Listen http endpoints events
func (r *ZookeeperDiscoveryService) WatchEndpoints(events chan<- flux.HttpEndpointEvent) error {
	const msg = "DISCOVERY:ZOOKEEPER:ENDPOINT:LISTEN_NODE"
	listener := func(event remoting.NodeEvent) {
		defer func() {
			if r := recover(); nil != r {
				logger.Errorw(msg, "event", event, "error", r)
			}
		}()
		if evt, ok := NewEndpointEvent(event.Data, event.EventType, event.Path); ok {
			events <- evt
		}
	}
	logger.Infow(msg, "node-path", r.endpointPath)
	for _, retriever := range r.retrievers {
		if err := r.watch(retriever, r.endpointPath, listener); err != nil {
			return err
		}
	}
	return nil
}

// OnServiceChanged Listen gateway services events
func (r *ZookeeperDiscoveryService) WatchServices(events chan<- flux.BackendServiceEvent) error {
	const msg = "DISCOVERY:ZOOKEEPER:SERVICE:LISTEN_NODE"
	listener := func(event remoting.NodeEvent) {
		defer func() {
			if r := recover(); nil != r {
				logger.Errorw(msg, "event", event, "error", r)
			}
		}()
		if evt, ok := NewBackendServiceEvent(event.Data, event.EventType, event.Path); ok {
			events <- evt
		}
	}
	logger.Infow(msg, "node-path", r.servicePath)
	for _, retriever := range r.retrievers {
		if err := r.watch(retriever, r.servicePath, listener); err != nil {
			return err
		}
	}
	return nil
}

func (r *ZookeeperDiscoveryService) watch(retriever *zk.ZookeeperRetriever, rootpath string, nodeListener func(remoting.NodeEvent)) error {
	if exist, _ := retriever.Exists(rootpath); !exist {
		if err := retriever.Create(rootpath); nil != err {
			return fmt.Errorf("init metadata node: %w", err)
		}
	}
	logger.Infow("DISCOVERY:ZOOKEEPER:META:WATCH", "path", rootpath)
	return retriever.AddChildrenNodeChangedListener("", rootpath, func(event remoting.NodeEvent) {
		logger.Infow("DISCOVERY:ZOOKEEPER:META:RECEIVED", "event", event)
		if event.EventType == remoting.EventTypeChildAdd {
			if err := retriever.AddNodeChangedListener("", event.Path, nodeListener); nil != err {
				logger.Warnw("Watch child node data", "error", err)
			}
		}
	})
}

// Startup startup discovery service
func (r *ZookeeperDiscoveryService) Startup() error {
	logger.Info("ZkEndpointDiscovery startup")
	for _, retriever := range r.retrievers {
		if err := retriever.Startup(); nil != err {
			return err
		}
	}
	return nil
}

// Shutdown shutdown discovery service
func (r *ZookeeperDiscoveryService) Shutdown(ctx context.Context) error {
	logger.Info("ZkEndpointDiscovery shutdown")
	for _, retriever := range r.retrievers {
		if err := retriever.Shutdown(ctx); nil != err {
			return err
		}
	}
	return nil
}
