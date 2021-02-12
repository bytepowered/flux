package boot

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/backend"
	"github.com/bytepowered/flux/discovery"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
)

func init() {
	// Default logger factory
	ext.SetLoggerFactory(logger.DefaultFactory)
	// 参数查找与解析函数
	ext.SetArgumentLookupFunc(backend.DefaultArgumentLookupFunc)
	// Serializer
	// Default: JSON
	serializer := flux.NewJsonSerializer()
	ext.SetSerializer(ext.TypeNameSerializerDefault, serializer)
	ext.SetSerializer(ext.TypeNameSerializerJson, serializer)
	// Endpoint discovery
	ext.SetEndpointDiscovery(discovery.NewZookeeperServiceWith(discovery.ZookeeperId))
	ext.SetEndpointDiscovery(discovery.NewResourceServiceWith(discovery.ResourceId))
}
