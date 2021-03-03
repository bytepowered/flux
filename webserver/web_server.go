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
	ConfigKeyServerName        = "name"
	ConfigKeyAddress           = "address"
	ConfigKeyBindPort          = "bind_port"
	ConfigKeyTLSCertFile       = "tls_cert_file"
	ConfigKeyTLSKeyFile        = "tls_key_file"
	ConfigKeyBodyLimit         = "body_limit"
	ConfigKeyGzipLevel         = "gzip_level"
	ConfigKeyCORSEnable        = "cors_enable"
	ConfigKeyCSRFEnable        = "csrf_enable"
	ConfigKeyFeatures          = "features"
	ConfigKeyRequestIdDisabled = "request_id_disabled"
)

var _ flux.ListenServer = new(AdaptWebServer)

func init() {
	ext.SetWebServerFactory(NewAdaptWebServer)
}

func NewAdaptWebServer(config *flux.Configuration) flux.ListenServer {
	return NewAdaptWebServerWith(config, nil)
}

func NewAdaptWebServerWith(options *flux.Configuration, mws *AdaptMiddleware) flux.ListenServer {
	server := echo.New()
	server.HideBanner = true
	server.HidePort = true
	aws := &AdaptWebServer{
		name:            options.GetString(ConfigKeyServerName),
		server:          server,
		requestResolver: DefaultRequestResolver,
	}
	// Init context
	server.Pre(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(echoc echo.Context) error {
			echoc.Set(ContextKeyWebResolver, aws.requestResolver)
			echoc.Set(ContextKeyWebBindServer, aws)
			return next(echoc)
		}
	})
	// Before feature
	server.Pre(RepeatableBodyReader)
	if mws != nil && len(mws.BeforeFeature) > 0 {
		server.Pre(mws.BeforeFeature...)
	}

	// Feature
	features := options.Sub(ConfigKeyFeatures)
	// 是否设置BodyLimit
	if limit := features.GetString(ConfigKeyBodyLimit); "" != limit {
		logger.Infof("WebServer(echo/%s), feature BODY-LIMIT: enabled, size= %s", aws.name, limit)
		server.Pre(middleware.BodyLimit(limit))
	}
	// 请求压缩
	if level := features.GetString(ConfigKeyGzipLevel); "" != level {
		levels := map[string]int{
			"nocompression":      flate.NoCompression,
			"bestspeed":          flate.BestSpeed,
			"bestcompression":    flate.BestCompression,
			"defaultcompression": flate.DefaultCompression,
			"huffmanonly":        flate.HuffmanOnly,
		}
		logger.Infof("WebServer(echo/%s), feature GZIP: enabled, level=%s", aws.name, level)
		server.Pre(middleware.GzipWithConfig(middleware.GzipConfig{
			Level: levels[strings.ToLower(level)],
		}))
	}
	// CORS
	if enabled := features.GetBool(ConfigKeyCORSEnable); enabled {
		logger.Infof("WebServer(echo/%s), feature CORS: enabled", aws.name)
		server.Pre(middleware.CORS())
	}
	// CSRF
	if enabled := features.GetBool(ConfigKeyCSRFEnable); enabled {
		logger.Infof("WebServer(echo/%s), feature CSRF: enabled", aws.name)
		server.Pre(middleware.CSRF())
	}
	// RequestId；默认开启
	if disabled := features.GetBool(ConfigKeyRequestIdDisabled); !disabled {
		logger.Infof("WebServer(echo/%s), feature RequestID: enabled", aws.name)
		server.Pre(RequestID())
	}
	// After features
	if mws != nil && len(mws.AfterFeature) > 0 {
		server.Pre(mws.AfterFeature...)
	}
	return aws
}

// AdaptWebServer 默认实现的基于echo框架的WebServer
// 注意：保持AdaptWebServer的公共访问性
type AdaptWebServer struct {
	name            string
	server          *echo.Echo
	writer          flux.WebResponseWriter
	requestResolver flux.WebRequestResolver
	tlsCertFile     string
	tlsKeyFile      string
	address         string
}

func (w *AdaptWebServer) Init(opts *flux.Configuration) error {
	w.tlsCertFile = opts.GetString(ConfigKeyTLSCertFile)
	w.tlsKeyFile = opts.GetString(ConfigKeyTLSKeyFile)
	addr, port := opts.GetString(ConfigKeyAddress), opts.GetString(ConfigKeyBindPort)
	if strings.Contains(addr, ":") {
		w.address = addr
	} else {
		w.address = addr + ":" + port
	}
	if w.address == ":" {
		return errors.New("web server config.address is required, was empty, server: " + w.name)
	}
	return nil
}

func (w *AdaptWebServer) Listen() error {
	logger.Infof("WebServer(echo/%s) start listen: %s", w.name, w.address)
	if "" != w.tlsCertFile && "" != w.tlsKeyFile {
		return w.server.StartTLS(w.address, w.tlsCertFile, w.tlsKeyFile)
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
		if nil == err {
			return
		}
		handler(wrapToAdaptWebContext(c), err)
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

type AdaptMiddleware struct {
	BeforeFeature []echo.MiddlewareFunc
	AfterFeature  []echo.MiddlewareFunc
}
