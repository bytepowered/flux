package server

import (
	"github.com/bytepowered/fluxgo/pkg/common"
	"github.com/bytepowered/fluxgo/pkg/discovery"
	"github.com/bytepowered/fluxgo/pkg/ext"
	"github.com/bytepowered/fluxgo/pkg/flux"
	"github.com/bytepowered/fluxgo/pkg/logger"
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
