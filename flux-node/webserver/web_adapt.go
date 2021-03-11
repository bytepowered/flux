package webserver

import (
	flux2 "github.com/bytepowered/flux/flux-node"
	"github.com/labstack/echo/v4"
)

// RouteHandler 实现flux.WebRouteHandler与Echo框架的echo.HandlerFunc函数适配
type AdaptWebHandler flux2.WebHandler

func (call AdaptWebHandler) AdaptFunc(ctx echo.Context) error {
	return call(toAdaptWebExchange(ctx))
}

// AdaptWebInterceptor 实现flux.WebInterceptor与Echo框架的echo.MiddlewareFunc函数适配
type AdaptWebInterceptor flux2.WebInterceptor

func (awi AdaptWebInterceptor) AdaptFunc(next echo.HandlerFunc) echo.HandlerFunc {
	call := awi(func(webex flux2.WebExchange) error {
		return next(webex.(*AdaptWebExchange).echoc)
	})
	return func(echoc echo.Context) error {
		return call(toAdaptWebExchange(echoc))
	}
}
