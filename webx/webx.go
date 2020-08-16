package webx

import (
	"context"
	"github.com/bytepowered/flux"
	"io"
	"net/http"
	"net/url"
)

const (
	charsetUTF8 = "charset=UTF-8"
)

// MIME types
const (
	MIMEApplicationJSON            = "application/json"
	MIMEApplicationJSONCharsetUTF8 = MIMEApplicationJSON + "; " + charsetUTF8
	MIMEApplicationForm            = "application/x-www-form-urlencoded"
)

const (
	HeaderContentType = "Content-Type"
	HeaderServer      = "Server"
	HeaderXRequestId  = flux.XRequestId
)

// Web interfaces
type (
	// WebMiddleware defines a function to process middleware.
	WebMiddleware func(WebRouteHandler) WebRouteHandler

	// WebRouteHandler defines a function to serve HTTP requests.
	WebRouteHandler func(WebContext) error

	// WebRouteHandler defines a function to handle errors.
	WebErrorHandler func(err error, ctx WebContext)
)

// WebContext defines a context for http server handlers/middleware
type WebContext interface {
	Method() string
	Host() string
	UserAgent() string
	Request() *http.Request
	RequestURI() string
	RequestURLPath() string
	RequestHeader() http.Header
	RequestBody() (io.ReadCloser, error)

	QueryValues() url.Values
	PathValues() url.Values
	FormValues() (url.Values, error)
	CookieValues() []*http.Cookie

	QueryValue(name string) string
	PathValue(name string) string
	FormValue(name string) string
	CookieValue(name string) (*http.Cookie, bool)

	Response() http.ResponseWriter
	ResponseHeader() http.Header
	ResponseWrite(statusCode int, bytes []byte) error

	SetValue(name string, value interface{})
	GetValue(name string) interface{}
}

// WebServerWriter 实现将错误消息和响应数据写入Response实例
type WebServerWriter interface {
	WriteError(webc WebContext, requestId string, header http.Header, error *flux.StateError) error
	WriteBody(webc WebContext, requestId string, header http.Header, status int, body interface{}) error
}

// WebServer
type WebServer interface {
	SetWebErrorHandler(h WebErrorHandler)
	SetRouteNotFoundHandler(h WebRouteHandler)
	AddWebInterceptor(m WebMiddleware)
	AddWebMiddleware(m WebMiddleware)
	AddWebRouteHandler(method, pattern string, h WebRouteHandler, m ...WebMiddleware)
	Start(addr string) error
	StartTLS(addr string, certFile, keyFile string) error
	Shutdown(ctx context.Context) error
}
