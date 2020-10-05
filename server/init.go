package server

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/backend/dubbo"
	"github.com/bytepowered/flux/backend/http"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/filter"
	"github.com/bytepowered/flux/registry"
)

func init() {
	// Default logger factory
	ext.SetLoggerFactory(DefaultLoggerFactory)
	// 参数查找函数
	ext.SetArgumentValueLookupFunc(flux.DefaultArgumentValueLookupFunc)
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
	// Dynamic factories
	ext.SetTypedFactory(filter.TypeIdJWTVerification, filter.JwtVerificationFilterFactory)
	ext.SetTypedFactory(filter.TypeIdEndpointPermission, filter.EndpointPermissionFactory)
	ext.SetTypedFactory(filter.TypeIdHystrixFilter, filter.HystrixFilterFactory)
	// Server
	SetServerWriterSerializer(serializer)
	SetServerResponseContentType(flux.MIMEApplicationJSONCharsetUTF8)
}
