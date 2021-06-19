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
	ext.SetLookupScopedValueFunc(common.LookupValueByScoped)
	// Serializer
	// Default: JSON
	serializer := flux.NewJsonSerializer()
	ext.RegisterSerializer(ext.TypeNameSerializerDefault, serializer)
	ext.RegisterSerializer(ext.TypeNameSerializerJson, serializer)
	// Endpoint discovery
	ext.RegisterEndpointDiscovery(discovery.NewZookeeperEndpointDiscovery(discovery.ZookeeperId))
	ext.RegisterEndpointDiscovery(discovery.NewResourceEndpointDiscovery(discovery.ResourceId))
}
