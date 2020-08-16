package webex

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

type (
	// MiddlewareFunc defines a function to process middleware.
	MiddlewareFunc func(HandlerFunc) HandlerFunc

	// HandlerFunc defines a function to serve HTTP requests.
	HandlerFunc func(WebContext) error

	ErrorHandlerFunc func(err error, ctx WebContext)

	WebContext interface {
		Request() *http.Request
		RequestMethod() string
		RequestURI() string
		RequestHost() string
		RequestUserAgent() string
		RequestURL() *url.URL
		RequestHeader() http.Header
		RequestBody() io.ReadCloser
		RequestRemoteAddr() string

		QueryValue(name string) string
		PathValue(name string) string
		FormValue(name string) string
		CookieValue(name string) string

		Response() http.ResponseWriter
		SetValue(name string, value interface{})
		GetValue(name string) interface{}
	}

	WebServer interface {
		SetErrorHandler(h ErrorHandlerFunc)
		SetNotFoundHandler(h HandlerFunc)
		AddInterceptor(m MiddlewareFunc)
		AddMiddleware(m MiddlewareFunc)
		AddRouteHandler(method, pattern string, h HandlerFunc, m ...MiddlewareFunc)
		Start(addr string) error
		StartTLS(addr string, certFile, keyFile string) error
		Shutdown(ctx context.Context) error
	}
)
