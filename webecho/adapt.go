package webecho

import (
	"github.com/bytepowered/flux"
	"github.com/labstack/echo/v4"
)

// RouteHandler 实现flux.WebRouteHandler与Echo框架的echo.HandlerFunc函数适配
type AdaptWebRouteHandler flux.WebRouteHandler

func (f AdaptWebRouteHandler) AdaptFunc(ctx echo.Context) error {
	return f(toAdaptWebContext(ctx))
}

// Middleware 实现flux.WebMiddleware与Echo框架的echo.MiddlewareFunc函数适配
type AdaptWebMiddleware flux.WebMiddleware

func (m AdaptWebMiddleware) AdaptFunc(next echo.HandlerFunc) echo.HandlerFunc {
	handler := m(func(webc flux.WebContext) error {
		return next(webc.(*AdaptWebContext).echoc)
	})
	return func(echoc echo.Context) error {
		return handler(toAdaptWebContext(echoc))
	}
}
