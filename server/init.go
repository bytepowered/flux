package server

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/discovery"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/support"
)

func init() {
	// Default logger factory
	ext.StoreLoggerFactory(DefaultLoggerFactory)
	// 参数查找与解析函数
	ext.StoreArgumentLookupFunc(support.DefaultArgumentValueLookupFunc)
	// Serializer
	// Default: JSON
	serializer := flux.NewJsonSerializer()
	ext.StoreSerializer(ext.TypeNameSerializerDefault, serializer)
	ext.StoreSerializer(ext.TypeNameSerializerJson, serializer)
	// Endpoint discovery
	// Default: ZK
	ext.StoreEndpointDiscoveryFactory(ext.EndpointDiscoveryProtoDefault, discovery.DefaultDiscoveryFactory)
	ext.StoreEndpointDiscoveryFactory(ext.EndpointDiscoveryProtoZookeeper, discovery.DefaultDiscoveryFactory)
	// Server
	SetServerWriterSerializer(serializer)
	SetServerResponseContentType(flux.MIMEApplicationJSONCharsetUTF8)
}
