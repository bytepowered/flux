package server

import (
	"context"
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/webmidware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cast"
	"net/http"
	_ "net/http/pprof"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

const (
	Banner        = "Flux-GO // Fast gateway for microservice: dubbo, grpc, http"
	VersionFormat = "Version // git.commit=%s, build.version=%s, build.date=%s"
)

const (
	DefaultHttpHeaderVersion = "X-Version"
)

const (
	HttpWebServerConfigRootName              = "HttpWebServer"
	HttpWebServerConfigKeyFeatureEchoEnable  = "feature-echo-enable"
	HttpWebServerConfigKeyFeatureDebugEnable = "feature-debug-enable"
	HttpWebServerConfigKeyFeatureDebugPort   = "feature-debug-port"
	HttpWebServerConfigKeyFeatureCorsEnable  = "feature-cors-enable"
	HttpWebServerConfigKeyVersionHeader      = "version-header"
	HttpWebServerConfigKeyRequestIdHeaders   = "request-id-headers"
	HttpWebServerConfigKeyRequestLogEnable   = "request-log-enable"
	HttpWebServerConfigKeyAddress            = "address"
	HttpWebServerConfigKeyPort               = "port"
	HttpWebServerConfigKeyTlsCertFile        = "tls-cert-file"
	HttpWebServerConfigKeyTlsKeyFile         = "tls-key-file"
)

var (
	HttpWebServerConfigDefaults = map[string]interface{}{
		HttpWebServerConfigKeyVersionHeader:      DefaultHttpHeaderVersion,
		HttpWebServerConfigKeyFeatureDebugEnable: false,
		HttpWebServerConfigKeyFeatureDebugPort:   9527,
		HttpWebServerConfigKeyAddress:            "0.0.0.0",
		HttpWebServerConfigKeyPort:               8080,
	}
)

// ServeEngine
type HttpServeEngine struct {
	httpWebServer        flux.WebServer
	serverResponseWriter flux.ServerResponseWriter
	serverErrorsWriter   flux.ServerErrorsWriter
	serverContextHooks   []flux.ServerContextHookFunc
	debugServer          *http.Server
	httpConfig           *flux.Configuration
	httpVersionHeader    string
	router               *Router
	endpointRegistry     flux.EndpointRegistry
	contextWrappers      sync.Pool
	stateStarted         chan struct{}
	stateStopped         chan struct{}
}

func NewHttpServeEngine() *HttpServeEngine {
	return NewHttpServeEngineWith(DefaultServerResponseWriter, DefaultServerErrorsWriter)
}

func NewHttpServeEngineWith(responseWriter flux.ServerResponseWriter, errorWriter flux.ServerErrorsWriter) *HttpServeEngine {
	return &HttpServeEngine{
		router:               NewRouter(),
		serverResponseWriter: responseWriter,
		serverErrorsWriter:   errorWriter,
		contextWrappers:      sync.Pool{New: NewContextWrapper},
		serverContextHooks:   make([]flux.ServerContextHookFunc, 0, 4),
		stateStarted:         make(chan struct{}),
		stateStopped:         make(chan struct{}),
	}
}

// Prepare Call before init and startup
func (s *HttpServeEngine) Prepare(hooks ...flux.PrepareHookFunc) error {
	for _, prepare := range append(ext.LoadPrepareHooks(), hooks...) {
		if err := prepare(); nil != err {
			return err
		}
	}
	return nil
}

// Initial
func (s *HttpServeEngine) Initial() error {
	// Http server
	s.httpConfig = flux.NewConfigurationOf(HttpWebServerConfigRootName)
	s.httpConfig.SetDefaults(HttpWebServerConfigDefaults)
	s.httpVersionHeader = s.httpConfig.GetString(HttpWebServerConfigKeyVersionHeader)
	// 创建WebServer
	s.httpWebServer = ext.LoadWebServerFactory()(s.httpConfig)
	// 默认必备的WebServer功能
	s.httpWebServer.SetWebErrorHandler(s.defaultServerErrorHandler)
	s.httpWebServer.SetWebNotFoundHandler(s.defaultNotFoundErrorHandler)

	// - 请求CORS跨域支持：默认关闭，需要配置开启
	if s.httpConfig.GetBool(HttpWebServerConfigKeyFeatureCorsEnable) {
		s.AddWebInterceptor(webmidware.NewCORSMiddleware())
	}

	// - RequestId是重要的参数，不可关闭；
	headers := s.httpConfig.GetStringSlice(HttpWebServerConfigKeyRequestIdHeaders)
	s.AddWebInterceptor(webmidware.NewRequestIdMiddlewareWithinHeader(headers...))

	// Internal Web Server
	port := s.httpConfig.GetInt(HttpWebServerConfigKeyFeatureDebugPort)
	s.debugServer = &http.Server{
		Handler: http.DefaultServeMux,
		Addr:    fmt.Sprintf("0.0.0.0:%d", port),
	}
	// Endpoint registry
	if registry, config, err := activeEndpointRegistry(); nil != err {
		return err
	} else {
		if err := s.router.InitialHook(registry, config); nil != err {
			return err
		}
		s.endpointRegistry = registry
	}
	// - Debug特性支持：默认关闭，需要配置开启
	if s.httpConfig.GetBool(HttpWebServerConfigKeyFeatureDebugEnable) {
		http.DefaultServeMux.Handle("/debug/endpoints", NewDebugQueryEndpointHandler())
		http.DefaultServeMux.Handle("/debug/services", NewDebugQueryServiceHandler())
		http.DefaultServeMux.Handle("/debug/metrics", promhttp.Handler())
	}
	// Echo feature
	if s.httpConfig.GetBool(HttpWebServerConfigKeyFeatureEchoEnable) {
		logger.Info("EchoEndpoint register")
		for _, evt := range NewEchoEndpoints() {
			s.HandleHttpEndpointEvent(evt)
		}
	}
	return s.router.Initial()
}

func (s *HttpServeEngine) Startup(version flux.BuildInfo) error {
	return s.StartServe(version, s.httpConfig)
}

// StartServe server
func (s *HttpServeEngine) StartServe(info flux.BuildInfo, config *flux.Configuration) error {
	if err := s.ensure().router.Startup(); nil != err {
		return err
	}
	// Http endpoints
	if events, err := s.endpointRegistry.WatchHttpEndpoints(); nil != err {
		return fmt.Errorf("start registry watching: %w", err)
	} else {
		go func() {
			logger.Info("HttpEndpoint event loop: starting")
			for event := range events {
				s.HandleHttpEndpointEvent(event)
			}
			logger.Info("HttpEndpoint event loop: Stopped")
		}()
	}
	// Backend services
	if events, err := s.endpointRegistry.WatchBackendServices(); nil != err {
		return fmt.Errorf("start registry watching: %w", err)
	} else {
		go func() {
			logger.Info("BackendService event loop: starting")
			for event := range events {
				s.HandleBackendServiceEvent(event)
			}
			logger.Info("BackendService event loop: Stopped")
		}()
	}
	close(s.stateStarted)
	logger.Info(Banner)
	logger.Infof(VersionFormat, info.CommitId, info.Version, info.Date)
	// Start Servers
	if s.debugServer != nil {
		go func() {
			logger.Infow("DebugServer starting", "address", s.debugServer.Addr)
			_ = s.debugServer.ListenAndServe()
		}()
	}
	address := fmt.Sprintf("%s:%d", config.GetString("address"), config.GetInt("port"))
	keyFile := config.GetString(HttpWebServerConfigKeyTlsKeyFile)
	certFile := config.GetString(HttpWebServerConfigKeyTlsCertFile)
	logger.Infow("HttpServeEngine starting", "address", address, "cert", certFile, "key", keyFile)
	return s.httpWebServer.StartTLS(address, certFile, keyFile)
}

func (s *HttpServeEngine) HandleEndpointRequest(webc flux.WebContext, endpoints *MultiEndpoint, tracing bool) error {
	version := webc.HeaderValue(s.httpVersionHeader)
	endpoint, found := endpoints.FindByVersion(version)
	requestId := cast.ToString(webc.GetValue(flux.HeaderXRequestId))
	defer func() {
		if r := recover(); r != nil {
			trace := logger.Trace(requestId)
			if err, ok := r.(error); ok {
				trace.Errorw("HttpServeEngine panics", "error", err)
			} else {
				trace.Errorw("HttpServeEngine panics", "recover", r)
			}
			trace.Error(string(debug.Stack()))
		}
	}()
	if !found {
		if tracing {
			url, _ := webc.RequestURL()
			logger.Trace(requestId).Infow("HttpServeEngine route not-found",
				"http-pattern", []string{webc.Method(), webc.RequestURI(), url.Path},
			)
		}
		return flux.ErrRouteNotFound
	}
	ctxw := s.acquireContext(requestId, webc, endpoint)
	defer s.releaseContext(ctxw)
	// Route call
	logger.TraceContext(ctxw).Infow("HttpServeEngine route start")
	endcall := func(code int, start time.Time) {
		logger.TraceContext(ctxw).Infow("HttpServeEngine route end",
			"metric", ctxw.LoadMetrics(),
			"elapses", time.Since(start).String(), "response.code", code)
	}
	start := time.Now()
	// Context hook
	for _, ctxhook := range s.serverContextHooks {
		ctxhook(webc, ctxw)
	}
	// Route and response
	response := ctxw.Response()
	if err := s.router.Route(ctxw); nil != err {
		defer endcall(err.StatusCode, start)
		logger.TraceContext(ctxw).Errorw("HttpServeEngine route error", "error", err)
		err.MergeHeader(response.HeaderValues())
		return err
	} else {
		defer endcall(response.StatusCode(), start)
		return s.serverResponseWriter(webc, requestId, response.HeaderValues(), response.StatusCode(), response.Body())
	}
}

func (s *HttpServeEngine) HandleBackendServiceEvent(event flux.BackendServiceEvent) {
	service := event.Service
	switch event.EventType {
	case flux.EventTypeAdded:
		logger.Infow("New service",
			"service-id", service.ServiceId, "alias-id", service.AliasId)
		ext.StoreBackendService(service)
		if "" != service.AliasId {
			ext.StoreBackendServiceById(service.AliasId, service)
		}
	case flux.EventTypeUpdated:
		logger.Infow("Update service",
			"service-id", service.ServiceId, "alias-id", service.AliasId)
		ext.StoreBackendService(service)
		if "" != service.AliasId {
			ext.StoreBackendServiceById(service.AliasId, service)
		}
	case flux.EventTypeRemoved:
		logger.Infow("Delete service",
			"service-id", service.ServiceId, "alias-id", service.AliasId)
		ext.RemoveBackendService(service.ServiceId)
		if "" != service.AliasId {
			ext.RemoveBackendService(service.AliasId)
		}
	}
}

func (s *HttpServeEngine) HandleHttpEndpointEvent(event flux.HttpEndpointEvent) {
	method := strings.ToUpper(event.Endpoint.HttpMethod)
	// Check http method
	if !isAllowedHttpMethod(method) {
		logger.Warnw("Unsupported http method", "method", method, "pattern", event.Endpoint.HttpPattern)
		return
	}
	pattern := event.Endpoint.HttpPattern
	routeKey := fmt.Sprintf("%s#%s", method, pattern)
	// Refresh endpoint
	endpoint := event.Endpoint
	bind, isreg := s.selectMultiEndpoint(routeKey, &endpoint)
	switch event.EventType {
	case flux.EventTypeAdded:
		logger.Infow("New endpoint", "version", endpoint.Version, "method", method, "pattern", pattern)
		bind.Update(endpoint.Version, &endpoint)
		if isreg {
			logger.Infow("Register http handler", "method", method, "pattern", pattern)
			s.httpWebServer.AddWebHandler(method, pattern, s.newWrappedEndpointHandler(bind))
		}
	case flux.EventTypeUpdated:
		logger.Infow("Update endpoint", "version", endpoint.Version, "method", method, "pattern", pattern)
		bind.Update(endpoint.Version, &endpoint)
	case flux.EventTypeRemoved:
		logger.Infow("Delete endpoint", "method", method, "pattern", pattern)
		bind.Delete(endpoint.Version)
	}
}

// Shutdown to cleanup resources
func (s *HttpServeEngine) Shutdown(ctx context.Context) error {
	logger.Info("HttpServeEngine shutdown...")
	defer close(s.stateStopped)
	if s.debugServer != nil {
		_ = s.debugServer.Close()
	}
	if err := s.httpWebServer.Shutdown(ctx); nil != err {
		return err
	}
	return s.router.Shutdown(ctx)
}

// StateStarted 返回一个Channel。当服务启动完成时，此Channel将被关闭。
func (s *HttpServeEngine) StateStarted() <-chan struct{} {
	return s.stateStarted
}

// StateStopped 返回一个Channel。当服务停止后完成时，此Channel将被关闭。
func (s *HttpServeEngine) StateStopped() <-chan struct{} {
	return s.stateStopped
}

// HttpConfig return Http server configuration
func (s *HttpServeEngine) HttpConfig() *flux.Configuration {
	return s.httpConfig
}

// AddWebInterceptor 添加Http前拦截器。将在Http被路由到对应Handler之前执行
func (s *HttpServeEngine) AddWebInterceptor(m flux.WebInterceptor) {
	s.ensure().httpWebServer.AddWebInterceptor(m)
}

// AddWebHandler 添加Http处理接口。
func (s *HttpServeEngine) AddWebHandler(method, pattern string, h flux.WebHandler, m ...flux.WebInterceptor) {
	s.ensure().httpWebServer.AddWebHandler(method, pattern, h, m...)
}

// AddWebHttpHandler 添加Http处理接口。
func (s *HttpServeEngine) AddWebHttpHandler(method, pattern string, h http.Handler, m ...func(http.Handler) http.Handler) {
	s.ensure().httpWebServer.AddWebHttpHandler(method, pattern, h, m...)
}

// SetWebNotFoundHandler 设置Http路由失败的处理接口
func (s *HttpServeEngine) SetWebNotFoundHandler(nfh flux.WebHandler) {
	s.ensure().httpWebServer.SetWebNotFoundHandler(nfh)
}

// WebServer 返回WebServer实例
func (s *HttpServeEngine) WebServer() flux.WebServer {
	return s.ensure().httpWebServer
}

// DebugServer 返回DebugServer实例，以及实体是否有效
func (s *HttpServeEngine) DebugServer() (*http.Server, bool) {
	return s.debugServer, nil != s.debugServer
}

// AddServerContextExchangeHook 添加Http与Flux的Context桥接函数
func (s *HttpServeEngine) AddServerContextExchangeHook(f flux.ServerContextHookFunc) {
	s.serverContextHooks = append(s.serverContextHooks, f)
}

func (s *HttpServeEngine) newWrappedEndpointHandler(endpoint *MultiEndpoint) flux.WebHandler {
	enabled := s.httpConfig.GetBool(HttpWebServerConfigKeyRequestLogEnable)
	return func(webc flux.WebContext) error {
		return s.HandleEndpointRequest(webc, endpoint, enabled)
	}
}

func (s *HttpServeEngine) selectMultiEndpoint(routeKey string, endpoint *flux.Endpoint) (*MultiEndpoint, bool) {
	if mve, ok := SelectMultiEndpoint(routeKey); ok {
		return mve, false
	} else {
		return RegisterMultiEndpoint(routeKey, endpoint), true
	}
}

func (s *HttpServeEngine) acquireContext(id string, webc flux.WebContext, endpoint *flux.Endpoint) *WrappedContext {
	ctx := s.contextWrappers.Get().(*WrappedContext)
	ctx.Reattach(id, webc, endpoint)
	return ctx
}

func (s *HttpServeEngine) releaseContext(context *WrappedContext) {
	context.Release()
	s.contextWrappers.Put(context)
}

func (s *HttpServeEngine) ensure() *HttpServeEngine {
	if s.httpWebServer == nil {
		logger.Panicf("Call must after InitialServer()")
	}
	return s
}

func (s *HttpServeEngine) defaultNotFoundErrorHandler(webc flux.WebContext) error {
	return &flux.ServeError{
		StatusCode: flux.StatusNotFound,
		ErrorCode:  flux.ErrorCodeRequestNotFound,
		Message:    flux.ErrorMessageWebServerRequestNotFound,
	}
}

func (s *HttpServeEngine) defaultServerErrorHandler(err error, webc flux.WebContext) {
	if err == nil {
		return
	}
	// Http中间件等返回InvokeError错误
	serve, ok := err.(*flux.ServeError)
	if !ok {
		serve = &flux.ServeError{
			StatusCode: flux.StatusServerError,
			ErrorCode:  flux.ErrorCodeGatewayInternal,
			Message:    err.Error(),
			Header:     http.Header{},
			Internal:   err,
		}
	}
	requestId := cast.ToString(webc.GetValue(flux.HeaderXRequestId))
	if err := s.serverErrorsWriter(webc, requestId, serve.Header, serve); nil != err {
		logger.Trace(requestId).Errorw("HttpServeEngine http response error", "error", err)
	}
}

func activeEndpointRegistry() (flux.EndpointRegistry, *flux.Configuration, error) {
	config := flux.NewConfigurationOf(flux.KeyConfigRootEndpointRegistry)
	config.SetDefault(flux.KeyConfigEndpointRegistryId, ext.EndpointRegistryIdDefault)
	registryId := config.GetString(flux.KeyConfigEndpointRegistryId)
	logger.Infow("Active endpoint registry", "registry-id", registryId)
	if factory, ok := ext.LoadEndpointRegistryFactory(registryId); !ok {
		return nil, config, fmt.Errorf("EndpointRegistryFactory not found, id: %s", registryId)
	} else {
		return factory(), config, nil
	}
}

func isAllowedHttpMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodPut,
		http.MethodHead, http.MethodOptions, http.MethodPatch, http.MethodTrace:
		// Allowed
		return true
	default:
		// http.MethodConnect, and Others
		logger.Errorw("Ignore unsupported http method:", "method", method)
		return false
	}
}
