package middleware

import (
	"github.com/bytepowered/flux/webx"
)

func NewCORSMiddleware() webx.WebMiddleware {
	return func(next webx.WebRouteHandler) webx.WebRouteHandler {
		return func(webc webx.WebContext) error {
			// TODO
			return next(webc)
		}
	}
}
