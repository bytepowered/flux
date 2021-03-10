package webserver

import (
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
	ConfigKeyServerId          = "server_id"
	ConfigKeyAddress           = "address"
	ConfigKeyBindPort          = "bind_port"
	ConfigKeyTLSCertFile       = "tls_cert_file"
	ConfigKeyTLSKeyFile        = "tls_key_file"
	ConfigKeyBodyLimit         = "body_limit"
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
		serverId:        options.GetString(ConfigKeyServerId),
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
		logger.Infof("WebServer(echo/id:%s), feature BODY-LIMIT: enabled, size= %s", aws.serverId, limit)
		server.Pre(middleware.BodyLimit(limit))
	}
	// CORS
	if enabled := features.GetBool(ConfigKeyCORSEnable); enabled {
		logger.Infof("WebServer(echo/id:%s), feature CORS: enabled", aws.serverId)
		server.Pre(middleware.CORS())
	}
	// CSRF
	if enabled := features.GetBool(ConfigKeyCSRFEnable); enabled {
		logger.Infof("WebServer(echo/id:%s), feature CSRF: enabled", aws.serverId)
		server.Pre(middleware.CSRF())
	}
	// RequestId；默认开启
	if disabled := features.GetBool(ConfigKeyRequestIdDisabled); !disabled {
		logger.Infof("WebServer(echo/id:%s), feature RequestID: enabled", aws.serverId)
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
	serverId        string
	server          *echo.Echo
	writer          flux.WebResponseWriter
	requestResolver flux.WebRequestResolver
	tlsCertFile     string
	tlsKeyFile      string
	address         string
}

func (s *AdaptWebServer) ServerId() string {
	return s.serverId
}

func (s *AdaptWebServer) Init(opts *flux.Configuration) error {
	s.tlsCertFile = opts.GetString(ConfigKeyTLSCertFile)
	s.tlsKeyFile = opts.GetString(ConfigKeyTLSKeyFile)
	addr, port := opts.GetString(ConfigKeyAddress), opts.GetString(ConfigKeyBindPort)
	if strings.Contains(addr, ":") {
		s.address = addr
	} else {
		s.address = addr + ":" + port
	}
	if s.address == ":" {
		return errors.New("web server config.address is required, was empty, server-id: " + s.serverId)
	}
	return nil
}

func (s *AdaptWebServer) Listen() error {
	logger.Infof("WebServer(echo/%s) start listen: %s", s.serverId, s.address)
	if "" != s.tlsCertFile && "" != s.tlsKeyFile {
		return s.server.StartTLS(s.address, s.tlsCertFile, s.tlsKeyFile)
	} else {
		return s.server.Start(s.address)
	}
}

func (s *AdaptWebServer) Write(webc flux.WebContext, header http.Header, status int, data interface{}) error {
	return s.writer(webc, header, status, data, nil)
}

func (s *AdaptWebServer) WriteError(webc flux.WebContext, err *flux.ServeError) {
	if err := s.writer(webc, err.Header, err.StatusCode, nil, err); nil != err {
		logger.Errorw("WebServer write error failed", "error", err, "server-id", s.serverId)
	}
}

func (s *AdaptWebServer) WriteNotfound(webc flux.WebContext) error {
	return echo.NotFoundHandler(webc.WebContext().(echo.Context))
}

func (s *AdaptWebServer) SetResponseWriter(f flux.WebResponseWriter) {
	s.writer = pkg.RequireNotNil(f, "WebResponseWriter is nil, server-id: "+s.serverId).(flux.WebResponseWriter)
}

func (s *AdaptWebServer) SetRequestResolver(resolver flux.WebRequestResolver) {
	s.requestResolver = resolver
}

func (s *AdaptWebServer) SetNotfoundHandler(fun flux.WebHandler) {
	echo.NotFoundHandler = AdaptWebRouteHandler(fun).AdaptFunc
}

func (s *AdaptWebServer) SetServerErrorHandler(handler flux.WebServerErrorHandler) {
	// Route请求返回的Error，全部经由此函数处理
	s.server.HTTPErrorHandler = func(err error, c echo.Context) {
		webc, ok := c.Get(ContextKeyWebContext).(*AdaptWebContext)
		pkg.Assert(ok, "<web-context> is invalid in http-error-handler")
		handler(webc, err)
	}
}

func (s *AdaptWebServer) AddInterceptor(m flux.WebInterceptor) {
	s.server.Pre(AdaptWebInterceptor(m).AdaptFunc)
}

func (s *AdaptWebServer) AddWebMiddleware(m flux.WebInterceptor) {
	s.server.Use(AdaptWebInterceptor(m).AdaptFunc)
}

func (s *AdaptWebServer) AddHandler(method, pattern string, h flux.WebHandler, m ...flux.WebInterceptor) {
	wms := make([]echo.MiddlewareFunc, len(m))
	for i, mi := range m {
		wms[i] = AdaptWebInterceptor(mi).AdaptFunc
	}
	s.server.Add(method, toRoutePattern(pattern), AdaptWebRouteHandler(h).AdaptFunc, wms...)
}

func (s *AdaptWebServer) AddHttpHandler(method, pattern string, h http.Handler, m ...func(http.Handler) http.Handler) {
	wms := make([]echo.MiddlewareFunc, len(m))
	for i, mf := range m {
		wms[i] = echo.WrapMiddleware(mf)
	}
	s.server.Add(method, toRoutePattern(pattern), echo.WrapHandler(h), wms...)
}

func (s *AdaptWebServer) Router() interface{} {
	return s.server
}

func (s *AdaptWebServer) Server() interface{} {
	return s.server
}

func (s *AdaptWebServer) Close(ctx context.Context) error {
	return s.server.Shutdown(ctx)
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
