package listener

import (
	"github.com/bytepowered/flux"
	fluxpkg "github.com/bytepowered/flux/fluxkit"
	"github.com/bytepowered/flux/internal"
	"github.com/labstack/echo/v4"
)

// AdaptWebHandler 实现flux.WebRouteHandler与Echo框架的echo.HandlerFunc函数适配
type AdaptWebHandler flux.WebHandler

func (call AdaptWebHandler) AdaptFunc(ctx echo.Context) error {
	return call(toServerWebContext(ctx))
}

// AdaptWebInterceptor 实现flux.WebInterceptor与Echo框架的echo.MiddlewareFunc函数适配
type AdaptWebInterceptor flux.WebInterceptor

func (call AdaptWebInterceptor) AdaptFunc(next echo.HandlerFunc) echo.HandlerFunc {
	return func(echoc echo.Context) error {
		return call(func(webex flux.ServerWebContext) error {
			return next(echoc)
		})(toServerWebContext(echoc))
	}
}

func toServerWebContext(echoc echo.Context) flux.ServerWebContext {
	webex, ok := echoc.Get(string(internal.CtxkeyWebContext)).(flux.ServerWebContext)
	fluxpkg.Assert(ok == true && webex != nil, "<server-web-context> not found in echo.context")
	return webex
}
