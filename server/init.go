package server

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/backend"
	"github.com/bytepowered/flux/discovery"
	"github.com/bytepowered/flux/ext"
)

func init() {
	// Default logger factory
	ext.StoreLoggerFactory(DefaultLoggerFactory)
	// 参数查找与解析函数
	ext.StoreArgumentLookupFunc(backend.DefaultArgumentValueLookupFunc)
	// Serializer
	// Default: JSON
	serializer := flux.NewJsonSerializer()
	ext.StoreSerializer(ext.TypeNameSerializerDefault, serializer)
	ext.StoreSerializer(ext.TypeNameSerializerJson, serializer)
	// Endpoint discovery
	ext.StoreEndpointDiscovery(discovery.NewZookeeperServiceWith(discovery.ZookeeperId))
	ext.StoreEndpointDiscovery(discovery.NewResourceServiceWith(discovery.ResourceId))
	// Server
	SetServerWriterSerializer(serializer)
	SetServerResponseContentType(flux.MIMEApplicationJSONCharsetUTF8)
}
