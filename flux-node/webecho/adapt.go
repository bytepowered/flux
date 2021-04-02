package webecho

import (
	"github.com/bytepowered/flux/flux-node"
	fluxpkg "github.com/bytepowered/flux/flux-pkg"
	"github.com/labstack/echo/v4"
)

// RouteHandler 实现flux.WebRouteHandler与Echo框架的echo.HandlerFunc函数适配
type EchoWebHandler flux.WebHandler

func (call EchoWebHandler) AdaptFunc(ctx echo.Context) error {
	return call(toServerWebContext(ctx))
}

// EchoWebInterceptor 实现flux.WebInterceptor与Echo框架的echo.MiddlewareFunc函数适配
type EchoWebInterceptor flux.WebInterceptor

func (call EchoWebInterceptor) AdaptFunc(next echo.HandlerFunc) echo.HandlerFunc {
	return func(echoc echo.Context) error {
		return call(func(webex flux.ServerWebContext) error {
			return next(echoc)
		})(toServerWebContext(echoc))
	}
}

func toServerWebContext(echoc echo.Context) flux.ServerWebContext {
	webex, ok := echoc.Get(__interContextKeyWebContext).(flux.ServerWebContext)
	fluxpkg.Assert(ok == true && webex != nil, "<server-web-context> not found in echo.context")
	return webex
}
