package echoserver

import (
	"bytes"
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
	"io"
	"io/ioutil"
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

var _ flux.WebListener = new(EchoWebListener)

func init() {
	ext.SetWebListenerFactory(NewEchoWebListener)
}

func NewEchoWebListener(listenerId string, config *flux.Configuration) flux.WebListener {
	return NewEchoWebListenerWith(listenerId, config, DefaultIdentifier, nil)
}

func NewEchoWebListenerWith(listenerId string, options *flux.Configuration, identifier flux.WebRequestIdentifier, mws *AdaptMiddleware) flux.WebListener {
	fluxpkg.Assert("" != listenerId, "empty <listener-id> in web listener configuration")
	server := echo.New()
	server.Pre(RepeatableReader)
	server.HideBanner = true
	server.HidePort = true
	aws := &EchoWebListener{
		id:              listenerId,
		server:          server,
		requestResolver: DefaultRequestBodyResolver,
	}
	// Init context
	server.Pre(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(echoc echo.Context) error {
			id := identifier(echoc)
			fluxpkg.Assert("" != id, "<request-id> is empty, return by id lookup func")
			webex := NewAdaptWebExchange(id, echoc, aws, aws.requestResolver)
			echoc.Set(__interContextKeyWebContext, webex)
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

// EchoWebListener 默认实现的基于echo框架的WebServer
// 注意：保持AdaptWebServer的公共访问性
type EchoWebListener struct {
	id              string
	server          *echo.Echo
	responseWriter  flux.WebResponseWriter
	requestResolver flux.WebRequestBodyResolver
	tlsCertFile     string
	tlsKeyFile      string
	address         string
	started         bool
}

func (s *EchoWebListener) ListenerId() string {
	return s.id
}

func (s *EchoWebListener) Init(opts *flux.Configuration) error {
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
	fluxpkg.AssertNotNil(s.requestResolver, "<request-resolver> is required, server-id: "+s.id)
	fluxpkg.AssertNotNil(s.responseWriter, "<response-writer> is required, server-id: "+s.id)
	return nil
}

func (s *EchoWebListener) Listen() error {
	logger.Infof("WebListener(id:%s) start listen: %s", s.id, s.address)
	s.started = true
	if "" != s.tlsCertFile && "" != s.tlsKeyFile {
		return s.server.StartTLS(s.address, s.tlsCertFile, s.tlsKeyFile)
	} else {
		return s.server.Start(s.address)
	}
}

func (s *EchoWebListener) Write(webex flux.WebExchange, header http.Header, status int, data interface{}) error {
	return s.responseWriter.Write(webex, header, status, data)
}

func (s *EchoWebListener) WriteError(webex flux.WebExchange, err *flux.ServeError) {
	if err := s.responseWriter.Write(webex, err.Header, err.StatusCode, err); nil != err {
		logger.Errorw("WebListener write error failed", "error", err, "server-id", s.id)
	}
}

func (s *EchoWebListener) WriteNotfound(webex flux.WebExchange) error {
	// Call not found handler
	return echo.NotFoundHandler(webex.ShadowContext().(echo.Context))
}

func (s *EchoWebListener) SetResponseWriter(f flux.WebResponseWriter) {
	fluxpkg.AssertNotNil(f, "WebResponseWriter must not nil, server-id: "+s.id)
	s.state()
	s.responseWriter = f
}

func (s *EchoWebListener) SetRequestBodyResolver(r flux.WebRequestBodyResolver) {
	fluxpkg.AssertNotNil(r, "WebRequestBodyResolver must not nil, server-id: "+s.id)
	s.state()
	s.requestResolver = r
}

func (s *EchoWebListener) SetNotfoundHandler(f flux.WebHandler) {
	fluxpkg.AssertNotNil(f, "NotfoundHandler must not nil, server-id: "+s.id)
	s.state()
	echo.NotFoundHandler = EchoWebHandler(f).AdaptFunc
}

func (s *EchoWebListener) SetErrorHandler(handler flux.WebErrorHandler) {
	// Route请求返回的Error，全部经由此函数处理
	fluxpkg.AssertNotNil(handler, "ErrorHandler must not nil, server-id: "+s.id)
	s.state()
	s.server.HTTPErrorHandler = func(err error, c echo.Context) {
		webex, ok := c.Get(__interContextKeyWebContext).(*EchoWebExchange)
		fluxpkg.Assert(ok, "<web-context> is invalid in http-error-handler")
		handler(webex, err)
	}
}

func (s *EchoWebListener) AddInterceptor(i flux.WebInterceptor) {
	fluxpkg.AssertNotNil(i, "Interceptor must not nil, server-id: "+s.id)
	s.server.Pre(EchoWebInterceptor(i).AdaptFunc)
}

func (s *EchoWebListener) AddMiddleware(m flux.WebInterceptor) {
	fluxpkg.AssertNotNil(m, "Middleware must not nil, server-id: "+s.id)
	s.server.Use(EchoWebInterceptor(m).AdaptFunc)
}

func (s *EchoWebListener) AddHandler(method, pattern string, h flux.WebHandler, is ...flux.WebInterceptor) {
	fluxpkg.AssertNotNil(h, "Handler must not nil, server-id: "+s.id)
	fluxpkg.Assert("" != method, "Method must not empty")
	fluxpkg.Assert("" != pattern, "Pattern must not empty")
	wms := make([]echo.MiddlewareFunc, len(is))
	for i, mi := range is {
		wms[i] = EchoWebInterceptor(mi).AdaptFunc
	}
	s.server.Add(method, toRoutePattern(pattern), EchoWebHandler(h).AdaptFunc, wms...)
}

func (s *EchoWebListener) AddHttpHandler(method, pattern string, h http.Handler, m ...func(http.Handler) http.Handler) {
	fluxpkg.AssertNotNil(h, "Handler must not nil, server-id: "+s.id)
	fluxpkg.Assert("" != method, "Method must not empty")
	fluxpkg.Assert("" != pattern, "Pattern must not empty")
	wms := make([]echo.MiddlewareFunc, len(m))
	for i, mf := range m {
		wms[i] = echo.WrapMiddleware(mf)
	}
	s.server.Add(method, toRoutePattern(pattern), echo.WrapHandler(h), wms...)
}

func (s *EchoWebListener) ShadowRouter() interface{} {
	return s.server
}

func (s *EchoWebListener) ShadowServer() interface{} {
	return s.server
}

func (s *EchoWebListener) Close(ctx context.Context) error {
	s.started = false
	return s.server.Shutdown(ctx)
}

func (s *EchoWebListener) state() {
	fluxpkg.Assert(!s.started, "illegal state: web listener is started")
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
func DefaultRequestBodyResolver(webex flux.WebExchange) url.Values {
	form, err := webex.(*EchoWebExchange).echoc.FormParams()
	if nil != err {
		panic(fmt.Errorf("parse form params failed, err: %w", err))
	}
	return form
}

func DefaultIdentifier(ctx interface{}) string {
	echoc, ok := ctx.(echo.Context)
	fluxpkg.Assert(ok, "<context> must be echo.context")
	id := echoc.Request().Header.Get(flux.XRequestId)
	if "" != id {
		return id
	}
	echoc.Request().Header.Set("X-RequestId-By", "flux")
	return "fxid_" + random.String(32)
}

// Body缓存，允许通过 GetBody 多次读取Body
func RepeatableReader(next echo.HandlerFunc) echo.HandlerFunc {
	// 包装Http处理错误，统一由HttpErrorHandler处理
	return func(echo echo.Context) error {
		request := echo.Request()
		data, err := ioutil.ReadAll(request.Body)
		if nil != err {
			return &flux.ServeError{
				StatusCode: flux.StatusBadRequest,
				ErrorCode:  flux.ErrorCodeGatewayInternal,
				Message:    flux.ErrorMessageRequestPrepare,
				CauseError: fmt.Errorf("read request body, method: %s, uri:%s, err: %w", request.Method, request.RequestURI, err),
			}
		}
		request.GetBody = func() (io.ReadCloser, error) {
			return ioutil.NopCloser(bytes.NewBuffer(data)), nil
		}
		// 恢复Body，但ParseForm解析后，request.Body无法重读，需要通过GetBody
		request.Body = ioutil.NopCloser(bytes.NewBuffer(data))
		return next(echo)
	}
}

type AdaptMiddleware struct {
	BeforeFeature []echo.MiddlewareFunc
	AfterFeature  []echo.MiddlewareFunc
}
