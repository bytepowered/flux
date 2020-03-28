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
	// Config factory
	ext.SetConfigFactory(func(ns string, m map[string]interface{}) flux.Config {
		return internal.NewMapConfig(m)
	})
	// Serializer
	serializer := internal.NewJsonSerializer()
	ext.SetSerializer(ext.TypeNameSerializerDefault, serializer)
	ext.SetSerializer(ext.TypeNameSerializerJson, serializer)
	// Registry
	ext.SetRegistryFactory(ext.RegistryIdDefault, registry.ZookeeperRegistryFactory)
	ext.SetRegistryFactory(ext.RegistryIdZookeeper, registry.ZookeeperRegistryFactory)
	ext.SetRegistryFactory(ext.RegistryIdEcho, registry.EchoRegistryFactory)
	// exchanges
	ext.SetExchange(flux.ProtocolEcho, echoex.NewEchoExchange())
	ext.SetExchange(flux.ProtocolHttp, http.NewHttpExchange())
	ext.SetExchange(flux.ProtocolDubbo, dubbo.NewDubboExchange())
	// filters
	ext.SetFactory(filter.FilterIdJWTVerification, filter.JwtVerificationFilterFactory)
	ext.SetFactory(filter.FilterIdPermissionVerification, filter.PermissionVerificationFactory)
	// global filters
	ext.AddGlobalFilter(filter.NewParameterParsingFilter())
}
