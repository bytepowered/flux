package discovery

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
	zkDiscoveryEndpointPath = "/flux-endpoint"
	zkDiscoveryServicePath  = "/flux-service"
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
func (r *ZookeeperDiscoveryService) OnInit(config *flux.Configuration) error {
	config.SetDefaults(map[string]interface{}{
		zkConfigRootpathEndpoint: zkDiscoveryEndpointPath,
		zkConfigRootpathService:  zkDiscoveryServicePath,
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
		zkconf.SetKeyAlias(map[string]string{
			"address":  "zookeeper.address",
			"password": "zookeeper.password",
			"timeout":  "zookeeper.timeout",
		})
		if len(r.globalAlias) != 0 {
			zkconf.SetKeyAlias(r.globalAlias)
		}
		logger.Infow("ZkEndpointDiscovery start zk discovery", "discovery-id", id)
		if err := r.retrievers[i].OnInit(zkconf); nil != err {
			return err
		}
	}
	return nil
}

// OnEndpointChanged Listen http endpoints events
func (r *ZookeeperDiscoveryService) WatchEndpoints(ctx context.Context, events chan<- flux.EndpointEvent) error {
	const msg = "DISCOVERY:ZOOKEEPER:ENDPOINT:LISTEN_NODE"
	// Watch回调函数
	callback := func(event remoting.NodeEvent) {
		defer func() {
			if r := recover(); nil != r {
				logger.Errorw(msg, "endpoint-event", event, "error", r)
			}
		}()
		if evt, err := NewEndpointEvent(event.Data, event.EventType); nil == err {
			events <- evt
		} else {
			logger.Errorw(msg, "endpoint-event", event, "error", err)
		}
	}
	logger.Infow(msg, "endpoint-path", r.endpointPath)
	return r.onRetrievers(ctx, r.endpointPath, callback)
}

// OnServiceChanged Listen gateway services events
func (r *ZookeeperDiscoveryService) WatchServices(ctx context.Context, events chan<- flux.ServiceEvent) error {
	const msg = "DISCOVERY:ZOOKEEPER:SERVICE:LISTEN_NODE"
	callback := func(event remoting.NodeEvent) {
		defer func() {
			if r := recover(); nil != r {
				logger.Errorw(msg, "endpoint-event", event, "error", r)
			}
		}()
		if evt, ok := NewServiceEvent(event.Data, event.EventType, event.Path); ok {
			events <- evt
		}
	}
	logger.Infow(msg, "endpoint-path", r.servicePath)
	return r.onRetrievers(ctx, r.servicePath, callback)
}

func (r *ZookeeperDiscoveryService) onRetrievers(ctx context.Context, path string, callback func(remoting.NodeEvent)) error {
	for _, retriever := range r.retrievers {
		watcher := func(ret *zk.ZookeeperRetriever, notify chan<- struct{}) {
			if err := r.watch(ret, path, callback); err != nil {
				logger.Errorw("DISCOVERY:ZOOKEEPER:RETRIEVERS:WATCH/Error", "watch-path", path, "error", err)
			} else {
				logger.Infow("DISCOVERY:ZOOKEEPER:RETRIEVERS:WATCH/Success", "watch-path", path)
			}
			notify <- struct{}{}
		}
		notify := make(chan struct{})
		go watcher(retriever, notify)
		select {
		case <-ctx.Done():
			logger.Infow("DISCOVERY:ZOOKEEPER:RETRIEVERS:WATCH/CANCELED", "watch-path", path)
			return nil
		case <-time.After(time.Minute):
			logger.Warnw("DISCOVERY:ZOOKEEPER:RETRIEVERS:WATCH/TIMEOUT", "watch-path", path)
		case <-notify:
			continue
		}
	}
	return nil
}

func (r *ZookeeperDiscoveryService) watch(retriever *zk.ZookeeperRetriever, rootpath string, nodeListener func(remoting.NodeEvent)) error {
	exist, err := retriever.Exists(rootpath)
	if nil != err {
		return fmt.Errorf("check path exists, path: %s, error: %w", rootpath, err)
	}
	if !exist {
		if err := retriever.Create(rootpath); nil != err {
			return fmt.Errorf("init metadata node: %w", err)
		}
	}
	return retriever.AddChildrenNodeChangedListener("", rootpath, func(event remoting.NodeEvent) {
		logger.Infow("DISCOVERY:ZOOKEEPER:RETRIEVERS:WATCH:RECV", "event", event)
		if event.EventType == remoting.EventTypeChildAdd {
			if err := retriever.AddNodeChangedListener("", event.Path, nodeListener); nil != err {
				logger.Warnw("Watch child node data", "error", err)
			}
		}
	})
}

// Startup startup discovery service
func (r *ZookeeperDiscoveryService) OnStartup() error {
	logger.Info("ZkEndpointDiscovery startup")
	for _, retriever := range r.retrievers {
		if err := retriever.OnStartup(); nil != err {
			return err
		}
	}
	return nil
}

// Shutdown shutdown discovery service
func (r *ZookeeperDiscoveryService) OnShutdown(ctx context.Context) error {
	logger.Info("ZkEndpointDiscovery shutdown")
	for _, retriever := range r.retrievers {
		if err := retriever.OnShutdown(ctx); nil != err {
			return err
		}
	}
	return nil
}
