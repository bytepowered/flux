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

func (f AdaptWebHandler) AdaptFunc(ctx echo.Context) error {
	return f(toWebContext(ctx))
}

// AdaptWebFilter 实现 flux.WebFilter 与Echo框架的echo.MiddlewareFunc函数适配
type AdaptWebFilter flux.WebFilter

func (f AdaptWebFilter) AdaptFunc(next echo.HandlerFunc) echo.HandlerFunc {
	return func(echoc echo.Context) error {
		return f(func(webex flux.WebContext) error {
			return next(echoc)
		})(toWebContext(echoc))
	}
}

func toWebContext(echoc echo.Context) flux.WebContext {
	webex, ok := echoc.Get(string(internal.CtxkeyWebContext)).(flux.WebContext)
	flux.Assert(ok == true && webex != nil, "<web-context> not found in echo.context")
	return webex
}
