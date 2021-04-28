package server

import (
	"github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-node/common"
	"github.com/bytepowered/flux/flux-node/discovery"
	"github.com/bytepowered/flux/flux-node/ext"
	"github.com/bytepowered/flux/flux-node/logger"
)

func init() {
	// Default logger factory
	ext.SetLoggerFactory(logger.DefaultFactory)
	// 参数查找与解析函数
	ext.SetLookupFunc(common.LookupMTValue)
	// Serializer
	// Default: JSON
	serializer := flux.NewJsonSerializer()
	ext.RegisterSerializer(ext.TypeNameSerializerDefault, serializer)
	ext.RegisterSerializer(ext.TypeNameSerializerJson, serializer)
	// Endpoint discovery
	ext.RegisterEndpointDiscovery(discovery.NewZookeeperServiceWith(discovery.ZookeeperId))
	ext.RegisterEndpointDiscovery(discovery.NewResourceServiceWith(discovery.ResourceId))
}
