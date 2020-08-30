package webecho

import (
	"context"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
)

var _ flux.WebServer = new(AdaptWebServer)

func init() {
	ext.SetWebServerFactory(NewAdaptWebServer)
}

func NewAdaptWebServer() flux.WebServer {
	server := echo.New()
	server.HideBanner = true
	server.HidePort = true
	// 可重读Body
	server.Pre(RepeatableBodyReader)
	return &AdaptWebServer{server}
}

type AdaptWebServer struct {
	server *echo.Echo
}

func (w *AdaptWebServer) SetRouteNotFoundHandler(fun flux.WebRouteHandler) {
	echo.NotFoundHandler = AdaptWebRouteHandler(fun).AdaptFunc
}

func (w *AdaptWebServer) SetWebErrorHandler(fun flux.WebErrorHandler) {
	w.server.HTTPErrorHandler = func(err error, c echo.Context) {
		fun(err, toAdaptWebContext(c))
	}
}

func (w *AdaptWebServer) AddWebInterceptor(m flux.WebMiddleware) {
	w.server.Pre(AdaptWebMiddleware(m).AdaptFunc)
}

func (w *AdaptWebServer) AddWebMiddleware(m flux.WebMiddleware) {
	w.server.Use(AdaptWebMiddleware(m).AdaptFunc)
}

func (w *AdaptWebServer) AddWebRouteHandler(method, pattern string, h flux.WebRouteHandler, m ...flux.WebMiddleware) {
	wms := make([]echo.MiddlewareFunc, len(m))
	for i, mi := range m {
		wms[i] = AdaptWebMiddleware(mi).AdaptFunc
	}
	w.server.Add(method, toRoutePattern(pattern), AdaptWebRouteHandler(h).AdaptFunc, wms...)
}

func (w *AdaptWebServer) AddStdHttpHandler(method, pattern string, h http.Handler, m ...func(http.Handler) http.Handler) {
	wms := make([]echo.MiddlewareFunc, len(m))
	for i, mf := range m {
		wms[i] = echo.WrapMiddleware(mf)
	}
	w.server.Add(method, toRoutePattern(pattern), echo.WrapHandler(h), wms...)
}

func (w *AdaptWebServer) WebRouter() interface{} {
	return w.server
}

func (w *AdaptWebServer) WebServer() interface{} {
	return w.server
}

func (w *AdaptWebServer) Start(addr string) error {
	return w.server.Start(addr)
}

func (w *AdaptWebServer) StartTLS(addr string, certFile, keyFile string) error {
	return w.server.StartTLS(addr, certFile, keyFile)
}

func (w *AdaptWebServer) Shutdown(ctx context.Context) error {
	return w.server.Shutdown(ctx)
}

func toRoutePattern(uri string) string {
	// /api/{userId} -> /api/:userId
	replaced := strings.Replace(uri, "}", "", -1)
	if len(replaced) < len(uri) {
		return strings.Replace(replaced, "{", ":", -1)
	} else {
		return uri
	}
}
