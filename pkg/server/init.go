package server

import (
	common "github.com/bytepowered/fluxgo/pkg/common"
	discovery2 "github.com/bytepowered/fluxgo/pkg/discovery"
	ext "github.com/bytepowered/fluxgo/pkg/ext"
	"github.com/bytepowered/fluxgo/pkg/flux"
	logger "github.com/bytepowered/fluxgo/pkg/logger"
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
	ext.RegisterEndpointDiscovery(discovery2.NewZookeeperEndpointDiscovery(discovery2.ZookeeperId))
	ext.RegisterEndpointDiscovery(discovery2.NewResourceEndpointDiscovery(discovery2.ResourceId))
}
