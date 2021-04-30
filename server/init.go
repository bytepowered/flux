package server

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/common"
	"github.com/bytepowered/flux/discovery"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
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
