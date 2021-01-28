package webserver

import (
	"context"
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/labstack/echo/v4"
	"net/http"
	"net/url"
	"strings"
)

var _ flux.ListenServer = new(AdaptWebServer)

func init() {
	ext.SetWebServerFactory(NewAdaptWebServer)
}

func NewAdaptWebServer(config *flux.Configuration) flux.ListenServer {
	server := echo.New()
	server.HideBanner = true
	server.HidePort = true
	aws := &AdaptWebServer{
		server:      server,
		bodyDecoder: DefaultRequestBodyDecoder,
	}
	// 注入EchoContext
	server.Pre(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(keyWebBodyDecoder, aws.bodyDecoder)
			return next(c)
		}
	})
	// 注入对Body的可重读逻辑
	server.Pre(RepeatableBodyReader)
	return aws
}

// AdaptWebServer 默认实现的基于echo框架的WebServer
// 注意：保持AdaptWebServer的公共访问性
type AdaptWebServer struct {
	server      *echo.Echo
	bodyDecoder flux.WebRequestBodyDecoder
}

func (w *AdaptWebServer) SetRequestBodyDecoder(decoder flux.WebRequestBodyDecoder) {
	w.bodyDecoder = decoder
}

func (w *AdaptWebServer) SetNotfoundHandler(fun flux.WebHandler) {
	echo.NotFoundHandler = AdaptWebRouteHandler(fun).AdaptFunc
}

func (w *AdaptWebServer) HandleNotfound(webc flux.WebContext) error {
	return echo.NotFoundHandler(webc.WebContext().(echo.Context))
}

func (w *AdaptWebServer) SetErrorHandler(fun flux.WebErrorHandler) {
	w.server.HTTPErrorHandler = func(err error, c echo.Context) {
		fun(err, toAdaptWebContext(c))
	}
}

func (w *AdaptWebServer) AddInterceptor(m flux.WebInterceptor) {
	w.server.Pre(AdaptWebInterceptor(m).AdaptFunc)
}

func (w *AdaptWebServer) AddWebMiddleware(m flux.WebInterceptor) {
	w.server.Use(AdaptWebInterceptor(m).AdaptFunc)
}

func (w *AdaptWebServer) AddHandler(method, pattern string, h flux.WebHandler, m ...flux.WebInterceptor) {
	wms := make([]echo.MiddlewareFunc, len(m))
	for i, mi := range m {
		wms[i] = AdaptWebInterceptor(mi).AdaptFunc
	}
	w.server.Add(method, toRoutePattern(pattern), AdaptWebRouteHandler(h).AdaptFunc, wms...)
}

func (w *AdaptWebServer) AddHttpHandler(method, pattern string, h http.Handler, m ...func(http.Handler) http.Handler) {
	wms := make([]echo.MiddlewareFunc, len(m))
	for i, mf := range m {
		wms[i] = echo.WrapMiddleware(mf)
	}
	w.server.Add(method, toRoutePattern(pattern), echo.WrapHandler(h), wms...)
}

func (w *AdaptWebServer) Router() interface{} {
	return w.server
}

func (w *AdaptWebServer) Server() interface{} {
	return w.server
}

func (w *AdaptWebServer) Listen(addr string, certFile, keyFile string) error {
	if "" == certFile || "" == keyFile {
		return w.server.Start(addr)
	} else {
		return w.server.StartTLS(addr, certFile, keyFile)
	}
}

func (w *AdaptWebServer) Close(ctx context.Context) error {
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

// 默认对RequestBody的表单数据进行解析
func DefaultRequestBodyDecoder(webc flux.WebContext) url.Values {
	form, err := webc.(*AdaptWebContext).echoc.FormParams()
	if nil != err {
		panic(fmt.Errorf("parse form params failed, err: %w", err))
	}
	return form
}
