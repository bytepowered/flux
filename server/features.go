package server

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/server/debug"
	"net/http"
)

func enableDebugFeature(s *HttpServer, config *flux.Configuration) {
	auth := debug.BasicAuthMiddleware(config)
	// Debug查询接口
	debugHandler := flux.WrapHttpHandler(http.DefaultServeMux)
	s.AddWebRouteHandler("GET", DebugPathVars, debugHandler, auth)
	s.AddWebRouteHandler("GET", DebugPathPprof, debugHandler, auth)
	// Endpoint查询
	s.AddWebRouteHandler("GET", DebugPathEndpoints, debug.QueryEndpointFeature(s.mvEndpointMap), auth)
}
