package zk

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
	_defaultInvalidFluxEvent = flux.EndpointEvent{}
)

// ZookeeperRegistry 基于ZK节点树实现的Endpoint元数据注册中心
type ZookeeperRegistry struct {
	zkRootPath string
	retriever  *zk.ZookeeperRetriever
}

func ZookeeperRegistryFactory() flux.Registry {
	return &ZookeeperRegistry{
		retriever: zk.NewZkRetriever(),
	}
}

func (r *ZookeeperRegistry) Init(config *flux.Configuration) error {
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
	r.zkRootPath = config.GetString("root-path")
	return r.retriever.InitWith(config)
}

// 监听Metadata配置变化
func (r *ZookeeperRegistry) WatchEvents(outboundEvents chan<- flux.EndpointEvent) error {
	if exists, _ := r.retriever.Exists(r.zkRootPath); !exists {
		if err := r.retriever.Create(r.zkRootPath); nil != err {
			return fmt.Errorf("init metadata node: %w", err)
		}
	}
	logger.Infow("Zookeeper watching metadata node", "path", r.zkRootPath)
	nodeChangeListener := func(event remoting.NodeEvent) {
		if evt, ok := toFluxEvent(event.Data, event.EventType); ok {
			outboundEvents <- evt
		}
	}
	err := r.retriever.WatchChildren("", r.zkRootPath, func(event remoting.NodeEvent) {
		logger.Infow("Receive child change event", "event", event)
		if event.EventType == remoting.EventTypeChildAdd {
			if err := r.retriever.WatchNodeData("", event.Path, nodeChangeListener); nil != err {
				logger.Warnw("Watch node data", "error", err)
			}
		}
	})
	return err
}

func (r *ZookeeperRegistry) Startup() error {
	logger.Info("Startup registry")
	return r.retriever.Startup()
}

func (r *ZookeeperRegistry) Shutdown(ctx context.Context) error {
	logger.Info("Shutdown registry")
	return r.retriever.Shutdown(ctx)
}

func toFluxEvent(zkData []byte, evtType remoting.EventType) (fxEvt flux.EndpointEvent, ok bool) {
	// Check json text
	size := len(zkData)
	if size < len("{\"k\":0}") || (zkData[0] != '[' && zkData[size-1] != '}') {
		return _defaultInvalidFluxEvent, false
	}
	endpoint := flux.Endpoint{}
	json := ext.GetSerializer(ext.TypeNameSerializerJson)
	if err := json.Unmarshal(zkData, &endpoint); nil != err {
		logger.Warnf("Parsing invalid endpoint registry, evt.type: %s, evt.data: %s", evtType, string(zkData), err)
		return _defaultInvalidFluxEvent, false
	}
	logger.Debugf("Parsed endpoint registry, event: %s, method: %s, uri-pattern: %s", evtType, endpoint.HttpMethod, endpoint.HttpPattern)
	if endpoint.HttpPattern == "" {
		logger.Infof("illegal http-pattern, data: %s", string(zkData))
		return _defaultInvalidFluxEvent, false
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
	switch evtType {
	case remoting.EventTypeNodeAdd:
		event.EventType = flux.EndpointEventAdded
	case remoting.EventTypeNodeDelete:
		event.EventType = flux.EndpointEventRemoved
	case remoting.EventTypeNodeUpdate:
		event.EventType = flux.EndpointEventUpdated
	default:
		return _defaultInvalidFluxEvent, false
	}
	return event, true
}

func _initArgumentValue(arg *flux.Argument) {
	arg.HttpValue = flux.NewWrapValue(nil)
	for i := range arg.Fields {
		_initArgumentValue(&arg.Fields[i])
	}
}