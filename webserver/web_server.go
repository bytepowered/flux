package webserver

import (
	"compress/flate"
	"context"
	"errors"
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/pkg"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
	"net/url"
	"strings"
)

const (
	ConfigKeyServerName = "name"
	ConfigKeyBodyLimit  = "body-limit"
	ConfigKeyGzipLevel  = "gzip-level"
	ConfigKeyCORSEnable = "cors-enable"
	ConfigKeyCSRFEnable = "csrf-enable"
)

var _ flux.ListenServer = new(AdaptWebServer)

func init() {
	ext.SetWebServerFactory(NewAdaptWebServer)
}

func NewAdaptWebServer(options *flux.Configuration) flux.ListenServer {
	server := echo.New()
	server.HideBanner = true
	server.HidePort = true
	aws := &AdaptWebServer{
		name:            options.GetString(ConfigKeyServerName),
		server:          server,
		requestResolver: DefaultRequestResolver,
	}
	features := options.Sub("features")
	// 是否设置BodyLimit
	if limit := features.GetString(ConfigKeyBodyLimit); "" != limit {
		logger.Infof("WebServer(echo/%s), enabled body limit: %s", aws.name, limit)
		server.Pre(middleware.BodyLimit(limit))
	}
	// 是否设置压缩
	if level := features.GetString(ConfigKeyGzipLevel); "" != level {
		levels := map[string]int{
			"nocompression":      flate.NoCompression,
			"bestspeed":          flate.BestSpeed,
			"bestcompression":    flate.BestCompression,
			"defaultcompression": flate.DefaultCompression,
			"huffmanonly":        flate.HuffmanOnly,
		}
		logger.Infof("WebServer(echo/%s), enabled gzip level: %s", aws.name, level)
		server.Pre(middleware.GzipWithConfig(middleware.GzipConfig{
			Level: levels[strings.ToLower(level)],
		}))
	}
	// 是否开启CORS
	if enabled := features.GetBool(ConfigKeyCORSEnable); enabled {
		logger.Infof("WebServer(echo/%s), enabled CORS feature", aws.name)
		server.Pre(middleware.CORS())
	}
	// 是否开启CSRF
	if enabled := features.GetBool(ConfigKeyCSRFEnable); enabled {
		logger.Infof("WebServer(echo/%s), enabled CSRF feature", aws.name)
		server.Pre(middleware.CSRF())
	}
	// 注入EchoContext
	server.Pre(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(keyWebBodyDecoder, aws.requestResolver)
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
	name            string
	server          *echo.Echo
	writer          flux.WebResponseWriter
	requestResolver flux.WebRequestResolver
	tlsCert         string
	tlsKey          string
	address         string
}

func (w *AdaptWebServer) Init(opts *flux.Configuration) error {
	w.tlsCert = opts.GetString("tlsCertFile")
	w.tlsKey = opts.GetString("tlsKeyFile")
	w.address = opts.GetString("address")
	if w.address == "" {
		return errors.New("web server config.address is required, was empty, server: " + w.name)
	}
	return nil
}

func (w *AdaptWebServer) Listen() error {
	logger.Infof("WebServer(echo/%s) start listen: %s", w.name, w.address)
	if "" != w.tlsCert && "" != w.tlsKey {
		return w.server.StartTLS(w.address, w.tlsCert, w.tlsKey)
	} else {
		return w.server.Start(w.address)
	}
}

func (w *AdaptWebServer) Write(webc flux.WebContext, header http.Header, status int, data interface{}) error {
	return w.writer(webc, header, status, data, nil)
}

func (w *AdaptWebServer) WriteError(webc flux.WebContext, err *flux.ServeError) {
	if err := w.writer(webc, err.Header, err.StatusCode, nil, err); nil != err {
		logger.Errorw("WebServer write error failed", "error", err, "server-name", w.name)
	}
}

func (w *AdaptWebServer) WriteNotfound(webc flux.WebContext) error {
	return echo.NotFoundHandler(webc.WebContext().(echo.Context))
}

func (w *AdaptWebServer) SetResponseWriter(f flux.WebResponseWriter) {
	w.writer = pkg.RequireNotNil(f, "WebResponseWriter is nil, server: "+w.name).(flux.WebResponseWriter)
}

func (w *AdaptWebServer) SetRequestResolver(resolver flux.WebRequestResolver) {
	w.requestResolver = resolver
}

func (w *AdaptWebServer) SetNotfoundHandler(fun flux.WebHandler) {
	echo.NotFoundHandler = AdaptWebRouteHandler(fun).AdaptFunc
}

func (w *AdaptWebServer) SetServerErrorHandler(handler flux.WebServerErrorHandler) {
	w.server.HTTPErrorHandler = func(err error, c echo.Context) {
		handler(toAdaptWebContext(c), err)
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
func DefaultRequestResolver(webc flux.WebContext) url.Values {
	form, err := webc.(*AdaptWebContext).echoc.FormParams()
	if nil != err {
		panic(fmt.Errorf("parse form params failed, err: %w", err))
	}
	return form
}
