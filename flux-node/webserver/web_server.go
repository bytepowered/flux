package webserver

import (
	"context"
	"errors"
	"fmt"
	"github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-node/ext"
	"github.com/bytepowered/flux/flux-node/logger"
	"github.com/bytepowered/flux/flux-pkg"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/random"
	"net/http"
	"net/url"
	"strings"
)

const (
	ConfigKeyAddress     = "address"
	ConfigKeyBindPort    = "bind_port"
	ConfigKeyTLSCertFile = "tls_cert_file"
	ConfigKeyTLSKeyFile  = "tls_key_file"
	ConfigKeyBodyLimit   = "body_limit"
	ConfigKeyCORSEnable  = "cors_enable"
	ConfigKeyCSRFEnable  = "csrf_enable"
	ConfigKeyFeatures    = "features"
)

var _ flux.WebListener = new(AdaptWebListener)

func init() {
	ext.SetWebListenerFactory(NewWebListener)
}

func NewWebListener(listenerId string, config *flux.Configuration) flux.WebListener {
	return NewWebListenerWith(listenerId, config, DefaultIdLookup, nil)
}

func NewWebListenerWith(listenerId string, options *flux.Configuration, LookupIdFunc flux.WebLookupIdFunc, mws *AdaptMiddleware) flux.WebListener {
	fluxpkg.Assert("" != listenerId, "empty <listener-id> in web listener configuration")
	server := echo.New()
	server.Pre(RepeatableBodyReader)
	server.HideBanner = true
	server.HidePort = true
	aws := &AdaptWebListener{
		id:              listenerId,
		server:          server,
		requestResolver: DefaultRequestResolver,
	}
	// Init context
	server.Pre(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(echoc echo.Context) error {
			id := LookupIdFunc(echoc)
			fluxpkg.Assert("" != id, "<request-id> is empty, return by id lookup func")
			webex := NewAdaptWebExchange(id, echoc, aws, aws.requestResolver)
			echoc.Set(ContextKeyWebContext, webex)
			return next(echoc)
		}
	})
	// Before feature
	if mws != nil && len(mws.BeforeFeature) > 0 {
		server.Pre(mws.BeforeFeature...)
	}

	// Feature
	features := options.Sub(ConfigKeyFeatures)
	// 是否设置BodyLimit
	if limit := features.GetString(ConfigKeyBodyLimit); "" != limit {
		logger.Infof("WebListener(id:%s), feature BODY-LIMIT: enabled, size= %s", aws.id, limit)
		server.Pre(middleware.BodyLimit(limit))
	}
	// CORS
	if enabled := features.GetBool(ConfigKeyCORSEnable); enabled {
		logger.Infof("WebListener(id:%s), feature CORS: enabled", aws.id)
		server.Pre(middleware.CORS())
	}
	// CSRF
	if enabled := features.GetBool(ConfigKeyCSRFEnable); enabled {
		logger.Infof("WebListener(id:%s), feature CSRF: enabled", aws.id)
		server.Pre(middleware.CSRF())
	}
	// After features
	if mws != nil && len(mws.AfterFeature) > 0 {
		server.Pre(mws.AfterFeature...)
	}
	return aws
}

// AdaptWebListener 默认实现的基于echo框架的WebServer
// 注意：保持AdaptWebServer的公共访问性
type AdaptWebListener struct {
	id              string
	server          *echo.Echo
	writer          flux.WebResponseWriter
	requestResolver flux.WebRequestResolver
	tlsCertFile     string
	tlsKeyFile      string
	address         string
}

func (s *AdaptWebListener) ListenerId() string {
	return s.id
}

func (s *AdaptWebListener) Init(opts *flux.Configuration) error {
	s.tlsCertFile = opts.GetString(ConfigKeyTLSCertFile)
	s.tlsKeyFile = opts.GetString(ConfigKeyTLSKeyFile)
	addr, port := opts.GetString(ConfigKeyAddress), opts.GetString(ConfigKeyBindPort)
	if strings.Contains(addr, ":") {
		s.address = addr
	} else {
		s.address = addr + ":" + port
	}
	if s.address == ":" {
		return errors.New("web server config.address is required, was empty, server-id: " + s.id)
	}
	return nil
}

func (s *AdaptWebListener) Listen() error {
	logger.Infof("WebListener(id:%s) start listen: %s", s.id, s.address)
	if "" != s.tlsCertFile && "" != s.tlsKeyFile {
		return s.server.StartTLS(s.address, s.tlsCertFile, s.tlsKeyFile)
	} else {
		return s.server.Start(s.address)
	}
}

func (s *AdaptWebListener) Write(webex flux.WebExchange, header http.Header, status int, data interface{}) error {
	return s.writer(webex, header, status, data, nil)
}

func (s *AdaptWebListener) WriteError(webex flux.WebExchange, err *flux.ServeError) {
	if err := s.writer(webex, err.Header, err.StatusCode, nil, err); nil != err {
		logger.Errorw("WebListener write error failed", "error", err, "server-id", s.id)
	}
}

func (s *AdaptWebListener) WriteNotfound(webex flux.WebExchange) error {
	return echo.NotFoundHandler(webex.ShadowContext().(echo.Context))
}

func (s *AdaptWebListener) SetResponseWriter(f flux.WebResponseWriter) {
	s.writer = fluxpkg.MustNotNil(f, "WebResponseWriter is nil, server-id: "+s.id).(flux.WebResponseWriter)
}

func (s *AdaptWebListener) SetRequestResolver(resolver flux.WebRequestResolver) {
	s.requestResolver = resolver
}

func (s *AdaptWebListener) SetNotfoundHandler(fun flux.WebHandler) {
	echo.NotFoundHandler = AdaptWebHandler(fun).AdaptFunc
}

func (s *AdaptWebListener) SetErrorHandler(handler flux.WebErrorHandler) {
	// Route请求返回的Error，全部经由此函数处理
	s.server.HTTPErrorHandler = func(err error, c echo.Context) {
		webex, ok := c.Get(ContextKeyWebContext).(*AdaptWebExchange)
		fluxpkg.Assert(ok, "<web-context> is invalid in http-error-handler")
		handler(webex, err)
	}
}

func (s *AdaptWebListener) AddInterceptor(m flux.WebInterceptor) {
	s.server.Pre(AdaptWebInterceptor(m).AdaptFunc)
}

func (s *AdaptWebListener) AddWebMiddleware(m flux.WebInterceptor) {
	s.server.Use(AdaptWebInterceptor(m).AdaptFunc)
}

func (s *AdaptWebListener) AddHandler(method, pattern string, h flux.WebHandler, m ...flux.WebInterceptor) {
	wms := make([]echo.MiddlewareFunc, len(m))
	for i, mi := range m {
		wms[i] = AdaptWebInterceptor(mi).AdaptFunc
	}
	s.server.Add(method, toRoutePattern(pattern), AdaptWebHandler(h).AdaptFunc, wms...)
}

func (s *AdaptWebListener) AddHttpHandler(method, pattern string, h http.Handler, m ...func(http.Handler) http.Handler) {
	wms := make([]echo.MiddlewareFunc, len(m))
	for i, mf := range m {
		wms[i] = echo.WrapMiddleware(mf)
	}
	s.server.Add(method, toRoutePattern(pattern), echo.WrapHandler(h), wms...)
}

func (s *AdaptWebListener) ShadowRouter() interface{} {
	return s.server
}

func (s *AdaptWebListener) ShadowServer() interface{} {
	return s.server
}

func (s *AdaptWebListener) Close(ctx context.Context) error {
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
func DefaultRequestResolver(webex flux.WebExchange) url.Values {
	form, err := webex.(*AdaptWebExchange).echoc.FormParams()
	if nil != err {
		panic(fmt.Errorf("parse form params failed, err: %w", err))
	}
	return form
}

func DefaultIdLookup(ctx interface{}) string {
	echoc, ok := ctx.(echo.Context)
	fluxpkg.Assert(ok, "<context> must be echo.context")
	id := echoc.Request().Header.Get(flux.XRequestId)
	if "" != id {
		return id
	}
	echoc.Request().Header.Set("X-RequestId-By", "flux")
	return "fxid_" + random.String(32)
}

type AdaptMiddleware struct {
	BeforeFeature []echo.MiddlewareFunc
	AfterFeature  []echo.MiddlewareFunc
}
