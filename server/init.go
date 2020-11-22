package server

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/backend/dubbo"
	"github.com/bytepowered/flux/backend/http"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/registry"
	"github.com/bytepowered/flux/support"
)

func init() {
	// Default logger factory
	ext.SetLoggerFactory(DefaultLoggerFactory)
	// 参数查找与解析函数
	ext.SetArgumentValueLookupFunc(support.DefaultArgumentValueLookupFunc)
	ext.SetArgumentValueResolveFunc(support.DefaultArgumentValueResolveFunc)
	// Serializer
	// Default: JSON
	serializer := flux.NewJsonSerializer()
	ext.SetSerializer(ext.TypeNameSerializerDefault, serializer)
	ext.SetSerializer(ext.TypeNameSerializerJson, serializer)
	// Endpoint registry
	// Default: ZK
	ext.SetEndpointRegistryFactory(ext.EndpointRegistryIdDefault, registry.ZkEndpointRegistryFactory)
	ext.SetEndpointRegistryFactory(ext.EndpointRegistryIdZookeeper, registry.ZkEndpointRegistryFactory)
	// Backends
	ext.SetBackend(flux.ProtoHttp, http.NewHttpBackend())
	ext.SetBackendResponseDecoder(flux.ProtoHttp, http.NewHttpBackendResponseDecoder())
	ext.SetBackend(flux.ProtoDubbo, dubbo.NewDubboBackend())
	ext.SetBackendResponseDecoder(flux.ProtoDubbo, dubbo.NewDubboBackendResponseDecoder())
	// Server
	SetServerWriterSerializer(serializer)
	SetServerResponseContentType(flux.MIMEApplicationJSONCharsetUTF8)
}
