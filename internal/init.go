package internal

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/exchange/dubbo"
	"github.com/bytepowered/flux/exchange/echoex"
	"github.com/bytepowered/flux/exchange/http"
	"github.com/bytepowered/flux/extension"
	"github.com/bytepowered/flux/filter"
	"github.com/bytepowered/flux/registry"
	"github.com/bytepowered/flux/serializer"
	"github.com/labstack/echo/v4/middleware"
)

func init() {
	// Http Middleware
	extension.AddHttpMiddleware(middleware.CORS())
	// serializer
	defaultSerializer := serializer.NewJsonSerializer()
	extension.SetSerializer(extension.TypeNameSerializerDefault, defaultSerializer)
	extension.SetSerializer(extension.TypeNameSerializerJson, defaultSerializer)
	// Registry
	extension.SetRegistryFactory(extension.TypeNameRegistryActive, registry.ZookeeperRegistryFactory)
	extension.SetRegistryFactory(extension.TypeNameRegistryZookeeper, registry.ZookeeperRegistryFactory)
	extension.SetRegistryFactory(extension.TypeNameRegistryEcho, registry.EchoRegistryFactory)
	// exchanges
	extension.SetExchange(flux.ProtocolEcho, echoex.NewEchoExchange())
	extension.SetExchange(flux.ProtocolHttp, http.NewHttpExchange())
	extension.SetExchange(flux.ProtocolDubbo, dubbo.NewDubboExchange())
	// filters
	extension.SetFactory(filter.TypeNameFilterJWTVerification, filter.JwtVerificationFilterFactory)
	extension.SetFactory(filter.TypeNameFilterPermissionVerification, filter.PermissionVerificationFactory)
	extension.SetFactory(filter.TypeNameFilterJWTVerification, filter.JwtVerificationFilterFactory)
	// global filters
	extension.SetGlobalFilter(filter.NewParameterParsingFilter())
}
