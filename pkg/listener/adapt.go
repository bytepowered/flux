package listener

import (
	"github.com/labstack/echo/v4"
)
import (
	"github.com/bytepowered/fluxgo/pkg/flux"
	"github.com/bytepowered/fluxgo/pkg/internal"
)

// AdaptWebHandler 实现flux.WebRouteHandler与Echo框架的echo.HandlerFunc函数适配
type AdaptWebHandler flux.WebHandlerFunc

func (call AdaptWebHandler) AdaptFunc(ctx echo.Context) error {
	return call(toServerWebContext(ctx))
}

// AdaptWebInterceptor 实现flux.WebInterceptor与Echo框架的echo.MiddlewareFunc函数适配
type AdaptWebInterceptor flux.WebFilter

func (call AdaptWebInterceptor) AdaptFunc(next echo.HandlerFunc) echo.HandlerFunc {
	return func(echoc echo.Context) error {
		return call(func(webex flux.WebContext) error {
			return next(echoc)
		})(toServerWebContext(echoc))
	}
}

func toServerWebContext(echoc echo.Context) flux.WebContext {
	webex, ok := echoc.Get(string(internal.CtxkeyWebContext)).(flux.WebContext)
	flux.Assert(ok == true && webex != nil, "<web-context> not found in echo.context")
	return webex
}
