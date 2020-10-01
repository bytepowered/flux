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
		retriever: zk.NewZkRetriever(),
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
	return r.retriever.InitWith(config)
}

func (r *ZookeeperEndpointRegistry) Watch() (<-chan flux.EndpointEvent, error) {
	if exist, _ := r.retriever.Exists(r.path); !exist {
		if err := r.retriever.Create(r.path); nil != err {
			return nil, fmt.Errorf("init metadata node: %w", err)
		}
	}
	logger.Infow("Zookeeper watching metadata node", "path", r.path)
	nodeListener := func(event remoting.NodeEvent) {
		if evt, ok := toEndpointEvent(event.Data, event.EventType); ok {
			r.events <- evt
		}
	}
	err := r.retriever.WatchChildren("", r.path, func(event remoting.NodeEvent) {
		logger.Infow("Receive child change event", "event", event)
		if event.EventType == remoting.EventTypeChildAdd {
			if err := r.retriever.WatchNodeData("", event.Path, nodeListener); nil != err {
				logger.Warnw("Watch node data", "error", err)
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

func toEndpointEvent(data []byte, etype remoting.EventType) (fxEvt flux.EndpointEvent, ok bool) {
	// Check json text
	size := len(data)
	if size < len("{\"k\":0}") || (data[0] != '[' && data[size-1] != '}') {
		return _invalidEndpointEvent, false
	}
	endpoint := flux.Endpoint{}
	json := ext.GetSerializer(ext.TypeNameSerializerJson)
	if err := json.Unmarshal(data, &endpoint); nil != err {
		logger.Warnf("Parsing invalid endpoint registry, evt.type: %s, evt.data: %s", etype, string(data), err)
		return _invalidEndpointEvent, false
	}
	logger.Debugf("Parsed endpoint registry, event: %s, method: %s, uri-pattern: %s", etype, endpoint.HttpMethod, endpoint.HttpPattern)
	if endpoint.HttpPattern == "" {
		logger.Infof("illegal http-pattern, data: %s", string(data))
		return _invalidEndpointEvent, false
	}
	// Init arg value
	for i := range endpoint.Arguments {
		_initArgumentValue(&endpoint.Arguments[i])
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

func _initArgumentValue(arg *flux.Argument) {
	arg.HttpValue = flux.NewWrapValue(nil)
	for i := range arg.Fields {
		_initArgumentValue(&arg.Fields[i])
	}
}
