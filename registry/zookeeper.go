package registry

import (
	"context"
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/remoting"
	"github.com/bytepowered/flux/remoting/zk"
	"time"
)

const (
	// 在ZK注册的根节点。需要与客户端的注册保持一致。
	zkRegistryRootNodePath = "/flux-metadata"
)

var (
	_                     flux.EndpointRegistry = new(ZookeeperEndpointRegistry)
	_invalidEndpointEvent                       = flux.EndpointEvent{}
)

// ZookeeperEndpointRegistry 基于ZK节点树实现的Endpoint元数据注册中心
type ZookeeperEndpointRegistry struct {
	path      string
	events    chan flux.EndpointEvent
	retriever *zk.ZookeeperRetriever
}

func ZkEndpointRegistryFactory() flux.EndpointRegistry {
	return &ZookeeperEndpointRegistry{
		retriever: zk.NewZookeeperRetriever(),
		events:    make(chan flux.EndpointEvent, 4),
	}
}

func (r *ZookeeperEndpointRegistry) Init(config *flux.Configuration) error {
	config.SetDefaults(map[string]interface{}{
		"root-path": zkRegistryRootNodePath,
		"timeout":   time.Second * 10,
	})
	config.SetGlobalAlias(map[string]string{
		"host":     "zookeeper.host",
		"port":     "zookeeper.port",
		"address":  "zookeeper.address",
		"password": "zookeeper.password",
		"database": "zookeeper.database",
	})
	r.path = config.GetString("root-path")
	return r.retriever.Init(config)
}

func (r *ZookeeperEndpointRegistry) WatchEvents() (<-chan flux.EndpointEvent, error) {
	if exist, _ := r.retriever.Exists(r.path); !exist {
		if err := r.retriever.Create(r.path); nil != err {
			return nil, fmt.Errorf("init metadata node: %w", err)
		}
	}
	logger.Infow("Zookeeper watching metadata node", "path", r.path)
	nodeListener := func(event remoting.NodeEvent) {
		defer func() {
			if r := recover(); nil != r {
				logger.Errorw("Zookeeper node listening", "event", event, "error", r)
			}
		}()
		if evt, ok := toEndpointEvent(event.Data, event.EventType); ok {
			r.events <- evt
		}
	}
	err := r.retriever.AddChildrenNodeChangedListener("", r.path, func(event remoting.NodeEvent) {
		logger.Infow("Receive child change event", "event", event)
		if event.EventType == remoting.EventTypeChildAdd {
			if err := r.retriever.AddNodeChangedListener("", event.Path, nodeListener); nil != err {
				logger.Warnw("WatchEvents node data", "error", err)
			}
		}
	})
	if err != nil {
		return nil, err
	} else {
		return r.events, err
	}
}

func (r *ZookeeperEndpointRegistry) Startup() error {
	logger.Info("Startup registry")
	return r.retriever.Startup()
}

func (r *ZookeeperEndpointRegistry) Shutdown(ctx context.Context) error {
	logger.Info("Shutdown registry")
	close(r.events)
	return r.retriever.Shutdown(ctx)
}

func toEndpointEvent(bytes []byte, etype remoting.EventType) (fxEvt flux.EndpointEvent, ok bool) {
	// Check json text
	size := len(bytes)
	if size < len("{\"k\":0}") || (bytes[0] != '[' && bytes[size-1] != '}') {
		logger.Infow("Invalid endpoint event data.size", "data", string(bytes))
		return _invalidEndpointEvent, false
	}
	endpoint := flux.Endpoint{}
	json := ext.GetSerializer(ext.TypeNameSerializerJson)
	if err := json.Unmarshal(bytes, &endpoint); nil != err {
		logger.Warnw("Parsing invalid endpoint registry",
			"event-type: ", etype, "data: %s", etype, string(bytes), "error", err)
		return _invalidEndpointEvent, false
	}
	logger.Infow("Received endpoint event",
		"event-type", etype, "method", endpoint.HttpMethod, "pattern", endpoint.HttpPattern, "data", string(bytes))
	if endpoint.HttpPattern == "" || endpoint.HttpMethod == "" {
		logger.Infof("illegal http-pattern, data: %s", string(bytes))
		return _invalidEndpointEvent, false
	}
	event := flux.EndpointEvent{
		HttpMethod:  endpoint.HttpMethod,
		HttpPattern: endpoint.HttpPattern,
		Endpoint:    endpoint,
	}
	switch etype {
	case remoting.EventTypeNodeAdd:
		event.EventType = flux.EndpointEventAdded
	case remoting.EventTypeNodeDelete:
		event.EventType = flux.EndpointEventRemoved
	case remoting.EventTypeNodeUpdate:
		event.EventType = flux.EndpointEventUpdated
	default:
		return _invalidEndpointEvent, false
	}
	return event, true
}
