package webserver

import (
	"github.com/bytepowered/flux"
	"github.com/labstack/echo/v4"
)

// RouteHandler 实现flux.WebRouteHandler与Echo框架的echo.HandlerFunc函数适配
type AdaptWebHandler flux.WebHandler

func (call AdaptWebHandler) AdaptFunc(ctx echo.Context) error {
	return call(toAdaptWebExchange(ctx))
}

// AdaptWebInterceptor 实现flux.WebInterceptor与Echo框架的echo.MiddlewareFunc函数适配
type AdaptWebInterceptor flux.WebInterceptor

func (awi AdaptWebInterceptor) AdaptFunc(next echo.HandlerFunc) echo.HandlerFunc {
	call := awi(func(webex flux.WebExchange) error {
		return next(webex.(*AdaptWebExchange).echoc)
	})
	return func(echoc echo.Context) error {
		return call(toAdaptWebExchange(echoc))
	}
}
