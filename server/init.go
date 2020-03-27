package server

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/exchange/dubbo"
	"github.com/bytepowered/flux/exchange/echoex"
	"github.com/bytepowered/flux/exchange/http"
	"github.com/bytepowered/flux/extension"
	"github.com/bytepowered/flux/filter"
	"github.com/bytepowered/flux/internal"
	"github.com/bytepowered/flux/registry"
)

func init() {
	// Config factory
	extension.SetConfigFactory(func(_ string, m map[string]interface{}) flux.Config {
		return internal.NewMapConfig(m)
	})
	// Serializer
	defaultSerializer := internal.NewJsonSerializer()
	extension.SetSerializer(extension.TypeNameSerializerDefault, defaultSerializer)
	extension.SetSerializer(extension.TypeNameSerializerJson, defaultSerializer)
	// Registry
	extension.SetRegistryFactory(extension.RegistryIdDefault, registry.ZookeeperRegistryFactory)
	extension.SetRegistryFactory(extension.RegistryIdZookeeper, registry.ZookeeperRegistryFactory)
	extension.SetRegistryFactory(extension.RegistryIdEcho, registry.EchoRegistryFactory)
	// exchanges
	extension.SetExchange(flux.ProtocolEcho, echoex.NewEchoExchange())
	extension.SetExchange(flux.ProtocolHttp, http.NewHttpExchange())
	extension.SetExchange(flux.ProtocolDubbo, dubbo.NewDubboExchange())
	// filters
	extension.SetFactory(filter.TypeNameFilterJWTVerification, filter.JwtVerificationFilterFactory)
	extension.SetFactory(filter.TypeNameFilterPermissionVerification, filter.PermissionVerificationFactory)
	extension.SetFactory(filter.TypeNameFilterJWTVerification, filter.JwtVerificationFilterFactory)
	// global filters
	extension.AddGlobalFilter(filter.NewParameterParsingFilter())
}
