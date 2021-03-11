package webserver

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/random"
)

type (
	// LookupRequestIdFunc defines a function to generate an ID.
	// Optional. Default value random.String(32).
	LookupRequestIdFunc func(echoc echo.Context) string
)

// RequestID returns a X-Request-ID middleware.
func RequestID() echo.MiddlewareFunc {
	return RequestIDWith(generator)
}

// RequestIDWith returns a X-Request-ID middleware with config.
func RequestIDWith(fun LookupRequestIdFunc) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()
			rid := req.Header.Get(echo.HeaderXRequestID)
			if rid == "" {
				rid = fun(c)
			}
			req.Header.Set(echo.HeaderXRequestID, rid)
			res.Header().Set(echo.HeaderXRequestID, rid)
			return next(c)
		}
	}
}

func generator(_ echo.Context) string {
	return random.String(32)
}
