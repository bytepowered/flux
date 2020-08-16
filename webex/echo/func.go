package echo

import (
	"github.com/bytepowered/flux/webex"
	"github.com/labstack/echo/v4"
)

func ToAdaptHandlerFunc(fun webex.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		return fun(NewAdaptEchoContext(ctx))
	}
}

func ToAdaptMiddlewareFunc(m webex.MiddlewareFunc) echo.MiddlewareFunc {
	//return func(next echo.HandlerFunc) echo.HandlerFunc {
	//	nextf := HandlerFunc(func(c Context) error {
	//		next(NewAdaptEchoContext(c))
	//	})
	//}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return next
	}
}
