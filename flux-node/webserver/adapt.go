package webserver

import (
	"github.com/bytepowered/flux/flux-node"
	fluxpkg "github.com/bytepowered/flux/flux-pkg"
	"github.com/labstack/echo/v4"
)

// RouteHandler 实现flux.WebRouteHandler与Echo框架的echo.HandlerFunc函数适配
type AdaptWebHandler flux.WebHandler

func (call AdaptWebHandler) AdaptFunc(ctx echo.Context) error {
	return call(toWebExchange(ctx))
}

// AdaptWebInterceptor 实现flux.WebInterceptor与Echo框架的echo.MiddlewareFunc函数适配
type AdaptWebInterceptor flux.WebInterceptor

func (call AdaptWebInterceptor) AdaptFunc(next echo.HandlerFunc) echo.HandlerFunc {
	return func(echoc echo.Context) error {
		return call(func(webex flux.WebExchange) error {
			return next(webex.(*AdaptWebExchange).echoc)
		})(toWebExchange(echoc))
	}
}

func toWebExchange(echoc echo.Context) flux.WebExchange {
	webex, ok := echoc.Get(ContextKeyWebContext).(*AdaptWebExchange)
	fluxpkg.Assert(ok == true, "<web-context> not found in echo.context")
	return webex
}
