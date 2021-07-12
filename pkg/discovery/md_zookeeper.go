package discovery

import (
	"context"
	"errors"
	"fmt"
	"time"
)

import (
	"github.com/bytepowered/fluxgo/pkg/flux"
	"github.com/bytepowered/fluxgo/pkg/logger"
	"github.com/bytepowered/fluxgo/pkg/remoting"
	"github.com/bytepowered/fluxgo/pkg/remoting/zk"
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

var _ flux.MetadataDiscovery = new(ZookeeperMetadataDiscovery)

type (
	// ZookeeperDiscoveryOption 配置函数
	ZookeeperDiscoveryOption func(discovery *ZookeeperMetadataDiscovery)
)

// ZookeeperMetadataDiscovery 基于ZK节点树实现的Endpoint元数据注册中心
type ZookeeperMetadataDiscovery struct {
	id                 string
	endpointPath       string
	servicePath        string
	retrievers         []*zk.ZookeeperRetriever
	decodeServiceFunc  DecodeServiceFunc
	decodeEndpointFunc DecodeEndpointFunc
	serviceFilter      ServiceFilter
	endpointFilter     EndpointFilter
}

func WithZookeeperDecodeServiceFunc(f DecodeServiceFunc) ZookeeperDiscoveryOption {
	return func(discovery *ZookeeperMetadataDiscovery) {
		discovery.decodeServiceFunc = f
	}
}

func WithZookeeperDecodeEndpointFunc(f DecodeEndpointFunc) ZookeeperDiscoveryOption {
	return func(discovery *ZookeeperMetadataDiscovery) {
		discovery.decodeEndpointFunc = f
	}
}

func WithZookeeperEndpointFilter(f EndpointFilter) ZookeeperDiscoveryOption {
	return func(discovery *ZookeeperMetadataDiscovery) {
		discovery.endpointFilter = f
	}
}

func WithZookeeperServiceFilter(f ServiceFilter) ZookeeperDiscoveryOption {
	return func(discovery *ZookeeperMetadataDiscovery) {
		discovery.serviceFilter = f
	}
}

// NewZookeeperMetadataDiscovery returns new a zookeeper discovery factory
func NewZookeeperMetadataDiscovery(id string, opts ...ZookeeperDiscoveryOption) *ZookeeperMetadataDiscovery {
	d := &ZookeeperMetadataDiscovery{
		id:                 id,
		decodeEndpointFunc: DecodeEndpoint,
		decodeServiceFunc:  DecodeService,
		endpointFilter: func(event remoting.NodeEvent, data *flux.EndpointSpec) bool {
			return true
		},
		serviceFilter: func(event remoting.NodeEvent, data *flux.ServiceSpec) bool {
			return true
		},
	}
	for _, opt := range opts {
		opt(d)
	}
	return d
}

func (d *ZookeeperMetadataDiscovery) Id() string {
	return d.id
}

// OnInit init discovery
func (d *ZookeeperMetadataDiscovery) OnInit(config *flux.Configuration) error {
	config.SetDefaults(map[string]interface{}{
		zkConfigRootpathEndpoint: zkDiscoveryEndpointPath,
		zkConfigRootpathService:  zkDiscoveryServicePath,
	})
	selected := config.GetStringSlice(zkConfigRegistrySelector)
	if len(selected) == 0 {
		selected = []string{"default"}
	}
	logger.Infow("METADISCOVERY:ZOOKEEPER:INIT", "selected-registry", selected)
	d.endpointPath = config.GetString(zkConfigRootpathEndpoint)
	d.servicePath = config.GetString(zkConfigRootpathService)
	if d.endpointPath == "" || d.servicePath == "" {
		return errors.New("config(rootpath_endpoint, rootpath_service) is empty")
	}
	d.retrievers = make([]*zk.ZookeeperRetriever, len(selected))
	registries := config.Sub("registry_centers")
	for i := range selected {
		id := selected[i]
		d.retrievers[i] = zk.NewZookeeperRetriever(id)
		zkconf := registries.Sub(id)
		logger.Infow("METADISCOVERY:ZOOKEEPER:INIT/start-eds", "discovery-id", id)
		if err := d.retrievers[i].OnInit(zkconf); nil != err {
			return err
		}
	}
	return nil
}

func (d *ZookeeperMetadataDiscovery) SubscribeEndpoints(ctx context.Context, events chan<- flux.EndpointEvent) error {
	callback := func(event *remoting.NodeEvent) (err error) {
		defer func() {
			if r := recover(); nil != r {
				err = fmt.Errorf("discovery(zk.endpoint) callback panic: %+v", r)
			}
		}()
		srv, err := d.decodeEndpointFunc(event.Data)
		if nil != err {
			return err
		}
		evt, err := ToEndpointEvent(&srv, event.Event)
		if nil == err {
			if !d.endpointFilter(*event, &evt.Endpoint) {
				return fmt.Errorf("skip by filter")
			}
			events <- evt
		}
		return err
	}
	logger.Infow("METADISCOVERY:ZOOKEEPER:ENDPOINT/watch", "ep-path", d.endpointPath)
	return d.onRetrievers(ctx, d.endpointPath, func(event remoting.NodeEvent) {
		if err := callback(&event); err != nil {
			logger.Warnw("METADISCOVERY:ZOOKEEPER:ENDPOINT/failed", "ep-event", event, "error", err)
		}
	})
}

func (d *ZookeeperMetadataDiscovery) SubscribeServices(ctx context.Context, events chan<- flux.ServiceEvent) error {
	callback := func(event *remoting.NodeEvent) (err error) {
		defer func() {
			if r := recover(); nil != r {
				err = fmt.Errorf("discovery(zk.service) callback panic: %+v", r)
			}
		}()
		srv, err := d.decodeServiceFunc(event.Data)
		if nil != err {
			return err
		}
		evt, err := ToServiceEvent(&srv, event.Event)
		if nil == err {
			if !d.serviceFilter(*event, &evt.Service) {
				return fmt.Errorf("skip by filter")
			}
			events <- evt
		}
		return err
	}
	logger.Infow("METADISCOVERY:ZOOKEEPER:SERVICE/watch", "service-path", d.servicePath)
	return d.onRetrievers(ctx, d.servicePath, func(event remoting.NodeEvent) {
		if err := callback(&event); err != nil {
			logger.Warnw("METADISCOVERY:ZOOKEEPER:SERVICE/failed", "service-event", event, "error", err)
		}
	})
}

func (d *ZookeeperMetadataDiscovery) onRetrievers(ctx context.Context, path string, callback func(remoting.NodeEvent)) error {
	for _, retriever := range d.retrievers {
		watcher := func(ret *zk.ZookeeperRetriever, notify chan<- struct{}) {
			if err := d.watch(ret, path, callback); err != nil {
				logger.Errorw("METADISCOVERY:ZOOKEEPER:WATCH/error", "watch-path", path, "error", err)
			} else {
				logger.Infow("METADISCOVERY:ZOOKEEPER:WATCH/success", "watch-path", path)
			}
			notify <- struct{}{}
		}
		notify := make(chan struct{})
		go watcher(retriever, notify)
		select {
		case <-ctx.Done():
			logger.Infow("METADISCOVERY:ZOOKEEPER:WATCH/canceled", "watch-path", path)
			return nil
		case <-time.After(time.Minute):
			logger.Warnw("METADISCOVERY:ZOOKEEPER:WATCH/timeout", "watch-path", path)
		case <-notify:
			continue
		}
	}
	return nil
}

func (d *ZookeeperMetadataDiscovery) watch(retriever *zk.ZookeeperRetriever, rootpath string, nodeListener func(remoting.NodeEvent)) error {
	exist, err := retriever.Exists(rootpath)
	if nil != err {
		return fmt.Errorf("check path exists, path: %s, error: %w", rootpath, err)
	}
	if !exist {
		if err := retriever.Create(rootpath); nil != err {
			return fmt.Errorf("init metadata node: %w", err)
		}
	}
	return retriever.AddChildChangedListener("", rootpath, func(event remoting.NodeEvent) {
		logger.Infow("METADISCOVERY:ZOOKEEPER:WATCH/recv", "event", event)
		if event.Event == remoting.EventTypeChildAdd {
			if err := retriever.AddChangedListener("", event.Path, nodeListener); nil != err {
				logger.Warnw("METADISCOVERY:ZOOKEEPER:WATCH/node", "path", event.Path, "error", err)
			}
		}
	})
}

// OnStartup startup discovery service
func (d *ZookeeperMetadataDiscovery) OnStartup() error {
	logger.Info("METADISCOVERY:ZOOKEEPER:STARTUP")
	for _, retriever := range d.retrievers {
		if err := retriever.OnStartup(); nil != err {
			return err
		}
	}
	return nil
}

// OnShutdown shutdown discovery service
func (d *ZookeeperMetadataDiscovery) OnShutdown(ctx context.Context) error {
	logger.Info("METADISCOVERY:ZOOKEEPER:SHUTDOWN")
	for _, retriever := range d.retrievers {
		if err := retriever.OnShutdown(ctx); nil != err {
			return err
		}
	}
	return nil
}
