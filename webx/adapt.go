package webx

import (
	"net/http"
)

// Adapt HttpHandler

func AdaptHttpHandler(h http.Handler) WebRouteHandler {
	return func(webc WebContext) error {
		h.ServeHTTP(webc.Response(), webc.Request())
		return nil
	}
}
