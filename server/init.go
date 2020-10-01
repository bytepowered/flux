package server

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/exchange/dubbo"
	"github.com/bytepowered/flux/exchange/http"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/filter"
	"github.com/bytepowered/flux/registry"
)

func init() {
	// Default logger factory
	ext.SetLoggerFactory(DefaultLoggerFactory)
	// 参数查找函数
	ext.SetEndpointArgumentValueLookupFunc(flux.DefaultEndpointArgumentValueLookup)
	// Serializer
	// Default: JSON
	serializer := flux.NewJsonSerializer()
	ext.SetSerializer(ext.TypeNameSerializerDefault, serializer)
	ext.SetSerializer(ext.TypeNameSerializerJson, serializer)
	// Endpoint registry
	// Default: ZK
	ext.SetEndpointRegistryFactory(ext.EndpointRegistryIdDefault, registry.ZkEndpointRegistryFactory)
	ext.SetEndpointRegistryFactory(ext.EndpointRegistryIdZookeeper, registry.ZkEndpointRegistryFactory)
	// Exchanges
	ext.SetExchange(flux.ProtoHttp, http.NewHttpExchange())
	ext.SetExchangeResponseDecoder(flux.ProtoHttp, http.NewHttpExchangeDecoder())
	ext.SetExchange(flux.ProtoDubbo, dubbo.NewDubboExchange())
	ext.SetExchangeResponseDecoder(flux.ProtoDubbo, dubbo.NewDubboExchangeDecoder())
	// Dynamic factories
	ext.SetTypedFactory(filter.TypeIdJWTVerification, filter.JwtVerificationFilterFactory)
	ext.SetTypedFactory(filter.TypeIdPermissionVerification, filter.PermissionVerificationFactory)
	ext.SetTypedFactory(filter.TypeIdHystrixFilter, filter.HystrixFilterFactory)
}
