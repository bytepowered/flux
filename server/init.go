package server

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/exchange/dubbo"
	"github.com/bytepowered/flux/exchange/echoex"
	"github.com/bytepowered/flux/exchange/http"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/filter"
	"github.com/bytepowered/flux/internal"
	"github.com/bytepowered/flux/registry"
)

func init() {
	// Default logger factory
	ext.SetLoggerFactory(DefaultLoggerFactory)
	// 参数查找函数
	ext.SetArgumentLookupFunc(internal.DefaultArgumentValueLookupFunc)
	// Serializer
	serializer := internal.NewJsonSerializer()
	ext.SetSerializer(ext.TypeNameSerializerDefault, serializer)
	ext.SetSerializer(ext.TypeNameSerializerJson, serializer)
	// Registry
	ext.SetRegistryFactory(ext.RegistryIdDefault, registry.ZookeeperRegistryFactory)
	ext.SetRegistryFactory(ext.RegistryIdZookeeper, registry.ZookeeperRegistryFactory)
	ext.SetRegistryFactory(ext.RegistryIdEcho, registry.EchoRegistryFactory)
	// exchanges
	ext.SetExchange(flux.ProtoEcho, echoex.NewEchoExchange())
	ext.SetExchange(flux.ProtoHttp, http.NewHttpExchange())
	ext.SetExchangeDecoder(flux.ProtoHttp, http.NewHttpExchangeDecoder())
	ext.SetExchange(flux.ProtoDubbo, dubbo.NewDubboExchange())
	ext.SetExchangeDecoder(flux.ProtoDubbo, dubbo.NewDubboExchangeDecoder())
	// dynamic filter factories
	ext.SetFactory(filter.TypeIdJWTVerification, filter.JwtVerificationFilterFactory)
	ext.SetFactory(filter.TypeIdPermissionVerification, filter.PermissionVerificationFactory)
	ext.SetFactory(filter.TypeIdRateLimitFilter, filter.RateLimitFilterFactory)
	ext.SetFactory(filter.TypeIdHystrixFilter, filter.HystrixFilterFactory)
	// global filters
	ext.AddGlobalFilter(filter.NewArgumentValueLookupFilter())
}
