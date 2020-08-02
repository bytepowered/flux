package server

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/exchange/dubbo"
	"github.com/bytepowered/flux/exchange/http"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/filter"
	"github.com/bytepowered/flux/internal"
	"github.com/bytepowered/flux/registry/zk"
)

func init() {
	// Default logger factory
	ext.SetLoggerFactory(DefaultLoggerFactory)
	// 参数查找函数
	ext.SetArgumentLookupFunc(internal.DefaultArgumentValueLookupFunc)
	// Serializer
	// Default: JSON
	serializer := internal.NewJsonSerializer()
	ext.SetSerializer(ext.TypeNameSerializerDefault, serializer)
	ext.SetSerializer(ext.TypeNameSerializerJson, serializer)
	// Registry
	// Default: ZK
	ext.SetRegistryFactory(ext.RegistryIdDefault, zk.ZookeeperRegistryFactory)
	ext.SetRegistryFactory(ext.RegistryIdZookeeper, zk.ZookeeperRegistryFactory)
	// Exchanges
	ext.SetExchange(flux.ProtoHttp, http.NewHttpExchange())
	ext.SetExchangeDecoder(flux.ProtoHttp, http.NewHttpExchangeDecoder())
	ext.SetExchange(flux.ProtoDubbo, dubbo.NewDubboExchange())
	ext.SetExchangeDecoder(flux.ProtoDubbo, dubbo.NewDubboExchangeDecoder())
	// Dynamic factories
	ext.SetFactory(filter.TypeIdJWTVerification, filter.JwtVerificationFilterFactory)
	ext.SetFactory(filter.TypeIdPermissionVerification, filter.PermissionVerificationFactory)
	ext.SetFactory(filter.TypeIdHystrixFilter, filter.HystrixFilterFactory)
}
