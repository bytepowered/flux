package listener

import (
	"context"
	"errors"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/internal"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/toolkit"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
	"runtime/debug"
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
	ext.SetWebListenerFactory(NewAdaptWebListener)
}

func NewAdaptWebListener(listenerId string, config *flux.Configuration) flux.WebListener {
	return NewAdaptWebListenerWith(listenerId, config, DefaultIdentifier, nil)
}

func NewAdaptWebListenerWith(listenerId string, options *flux.Configuration, identifier flux.WebRequestIdentifier, mws *AdaptMiddleware) flux.WebListener {
	toolkit.Assert("" != listenerId, "empty <listener-id> in web listener configuration")
	server := echo.New()
	server.Pre(RepeatableReader)
	server.HideBanner = true
	server.HidePort = true
	webListener := &AdaptWebListener{
		id:           listenerId,
		server:       server,
		bodyResolver: DefaultRequestBodyResolver,
	}
	// Init context
	server.Pre(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(echoc echo.Context) error {
			id := identifier(echoc)
			toolkit.Assert("" != id, "<request-id> is empty, return by id lookup func")
			swc := NewServeWebContext(echoc, id, webListener)
			// Note: 不要修改echo.context.request对象引用，echo路由绑定了函数入口的request对象，从而导致
			// 后续基于request修改路由时，会导致Http路由失败。
			toolkit.AssertNil(echoc.Get(string(internal.CtxkeyWebContext)), "<web-context> must be nil")
			echoc.Set(string(internal.CtxkeyWebContext), swc)
			defer func() {
				if rvr := recover(); rvr != nil && rvr != http.ErrAbortHandler {
					logger.Trace(id).Errorw("SERVER:CRITICAL:PANIC", "error", rvr, "error.trace", string(debug.Stack()))
					_ = echoc.JSON(http.StatusInternalServerError, map[string]interface{}{
						"server.traceid": id,
						"server.status":  "error",
						"error.level":    "critical",
						"error.message":  "unexpected fault of the server",
						"error.cause":    "internal error",
					})
				}
			}()
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
		logger.Infof("WebListener(id:%s), feature BODY-LIMIT: enabled, size= %s", webListener.id, limit)
		server.Pre(middleware.BodyLimit(limit))
	}
	// CORS
	if enabled := features.GetBool(ConfigKeyCORSEnable); enabled {
		logger.Infof("WebListener(id:%s), feature CORS: enabled", webListener.id)
		server.Pre(middleware.CORS())
	}
	// CSRF
	if enabled := features.GetBool(ConfigKeyCSRFEnable); enabled {
		logger.Infof("WebListener(id:%s), feature CSRF: enabled", webListener.id)
		server.Pre(middleware.CSRF())
	}
	// After features
	if mws != nil && len(mws.AfterFeature) > 0 {
		server.Pre(mws.AfterFeature...)
	}
	return webListener
}

// AdaptWebListener 默认实现的基于echo框架的WebServer
// 注意：保持AdaptWebServer的公共访问性
type AdaptWebListener struct {
	id           string
	server       *echo.Echo
	bodyResolver flux.WebBodyResolver
	tlsCertFile  string
	tlsKeyFile   string
	address      string
	started      bool
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
		return errors.New("web server config.address is required, was empty, listener-id: " + s.id)
	}
	toolkit.AssertNotNil(s.bodyResolver, "<body-resolver> is required, listener-id: "+s.id)
	return nil
}

func (s *AdaptWebListener) Listen() error {
	logger.Infof("WebListener(id:%s) start listen: %s", s.id, s.address)
	s.started = true
	if "" != s.tlsCertFile && "" != s.tlsKeyFile {
		return s.server.StartTLS(s.address, s.tlsCertFile, s.tlsKeyFile)
	} else {
		return s.server.Start(s.address)
	}
}

func (s *AdaptWebListener) SetBodyResolver(r flux.WebBodyResolver) {
	toolkit.AssertNotNil(r, "WebBodyResolver must not nil, listener-id: "+s.id)
	s.mustNotStarted().bodyResolver = r
}

func (s *AdaptWebListener) SetNotfoundHandler(f flux.WebHandler) {
	toolkit.AssertNotNil(f, "NotfoundHandler must not nil, listener-id: "+s.id)
	s.mustNotStarted()
	echo.NotFoundHandler = AdaptWebHandler(f).AdaptFunc
}

func (s *AdaptWebListener) HandleNotfound(webex flux.ServerWebContext) error {
	return echo.NotFoundHandler(webex.(*AdaptWebContext).ShadowContext())
}

func (s *AdaptWebListener) SetErrorHandler(handler flux.WebErrorHandler) {
	// Route请求返回的Error，全部经由此函数处理
	toolkit.AssertNotNil(handler, "ErrorHandler must not nil, listener-id: "+s.id)
	s.mustNotStarted().server.HTTPErrorHandler = func(err error, c echo.Context) {
		// 修正Error未判定为nil的问题问题
		if toolkit.IsNil(err) {
			return
		}
		webex, ok := c.Get(string(internal.CtxkeyWebContext)).(flux.ServerWebContext)
		toolkit.Assert(ok, "<web-context> is invalid in http-error-handler")
		handler(webex, err)
	}
}

func (s *AdaptWebListener) HandleError(webex flux.ServerWebContext, err error) {
	s.server.HTTPErrorHandler(err, webex.(*AdaptWebContext).ShadowContext())
}

func (s *AdaptWebListener) AddInterceptor(i flux.WebInterceptor) {
	toolkit.AssertNotNil(i, "Interceptor must not nil, listener-id: "+s.id)
	s.server.Pre(AdaptWebInterceptor(i).AdaptFunc)
}

func (s *AdaptWebListener) AddMiddleware(m flux.WebInterceptor) {
	toolkit.AssertNotNil(m, "Middleware must not nil, listener-id: "+s.id)
	s.server.Use(AdaptWebInterceptor(m).AdaptFunc)
}

func (s *AdaptWebListener) AddHandler(method, pattern string, h flux.WebHandler, is ...flux.WebInterceptor) {
	toolkit.AssertNotNil(h, "Handler must not nil, listener-id: "+s.id)
	toolkit.Assert(method != "", "Method must not empty")
	toolkit.Assert(pattern != "", "Pattern must not empty")
	wms := make([]echo.MiddlewareFunc, len(is))
	for i, mi := range is {
		wms[i] = AdaptWebInterceptor(mi).AdaptFunc
	}
	s.server.Add(strings.ToUpper(method), s.resolve(pattern), AdaptWebHandler(h).AdaptFunc, wms...)
}

func (s *AdaptWebListener) AddHttpHandler(method, pattern string, h http.Handler, m ...func(http.Handler) http.Handler) {
	toolkit.AssertNotNil(h, "Handler must not nil, listener-id: "+s.id)
	toolkit.Assert("" != method, "Method must not empty")
	toolkit.Assert("" != pattern, "Pattern must not empty")
	wms := make([]echo.MiddlewareFunc, len(m))
	for i, mf := range m {
		wms[i] = echo.WrapMiddleware(mf)
	}
	s.server.Add(strings.ToUpper(method), s.resolve(pattern), echo.WrapHandler(h), wms...)
}

func (s *AdaptWebListener) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.server.ServeHTTP(w, r)
}

func (s *AdaptWebListener) ShadowRouter() interface{} {
	return s.server
}

func (s *AdaptWebListener) ShadowServer() interface{} {
	return s.server
}

func (s *AdaptWebListener) Close(ctx context.Context) error {
	s.started = false
	return s.server.Shutdown(ctx)
}

func (s *AdaptWebListener) mustNotStarted() *AdaptWebListener {
	toolkit.Assert(!s.started, "illegal state: web listener is started")
	return s
}

func (s *AdaptWebListener) resolve(pattern string) string {
	// /api/{userId} -> /api/:userId
	replaced := strings.Replace(pattern, "}", "", -1)
	if len(replaced) < len(pattern) {
		return strings.Replace(replaced, "{", ":", -1)
	} else {
		return pattern
	}
}

type AdaptMiddleware struct {
	BeforeFeature []echo.MiddlewareFunc
	AfterFeature  []echo.MiddlewareFunc
}
