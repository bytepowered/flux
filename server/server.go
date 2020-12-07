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
	ErrEndpointVersionNotFound = &flux.StateError{
		StatusCode: flux.StatusNotFound,
		ErrorCode:  flux.ErrorCodeGatewayEndpoint,
		Message:    "ENDPOINT_VERSION_NOT_FOUND",
	}
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

// Server
type HttpWebServer struct {
	webServer                  flux.WebServer
	serverResponseWriter       flux.ServerResponseWriter
	serverErrorsWriter         flux.ServerErrorsWriter
	serverContextExchangeHooks []flux.ServerContextExchangeHook
	debugServer                *http.Server
	httpConfig                 *flux.Configuration
	httpVersionHeader          string
	routerEngine               *RouterEngine
	endpointRegistry           flux.EndpointRegistry
	bindEndpoint               map[string]*BindEndpoint
	contextWrappers            sync.Pool
	stateStarted               chan struct{}
	stateStopped               chan struct{}
}

func NewHttpServer() *HttpWebServer {
	return &HttpWebServer{
		serverResponseWriter:       DefaultServerResponseWriter,
		serverErrorsWriter:         DefaultServerErrorsWriter,
		routerEngine:               NewRouteEngine(),
		bindEndpoint:               make(map[string]*BindEndpoint),
		contextWrappers:            sync.Pool{New: NewContextWrapper},
		serverContextExchangeHooks: make([]flux.ServerContextExchangeHook, 0, 4),
		stateStarted:               make(chan struct{}),
		stateStopped:               make(chan struct{}),
	}
}

// Prepare Call before init and startup
func (s *HttpWebServer) Prepare(hooks ...flux.PrepareHookFunc) error {
	for _, prepare := range append(ext.LoadPrepareHooks(), hooks...) {
		if err := prepare(); nil != err {
			return err
		}
	}
	return nil
}

// Initial
func (s *HttpWebServer) Initial() error {
	// Http server
	s.httpConfig = flux.NewConfigurationOf(HttpWebServerConfigRootName)
	s.httpConfig.SetDefaults(HttpWebServerConfigDefaults)
	s.httpVersionHeader = s.httpConfig.GetString(HttpWebServerConfigKeyVersionHeader)
	// 创建WebServer
	s.webServer = ext.LoadWebServerFactory()()
	// 默认必备的WebServer功能
	s.webServer.SetWebErrorHandler(s.handleServerError)
	s.webServer.SetWebNotFoundHandler(s.handleNotFoundError)

	// - 请求CORS跨域支持：默认关闭，需要配置开启
	if s.httpConfig.GetBool(HttpWebServerConfigKeyFeatureCorsEnable) {
		s.AddWebInterceptor(webmidware.NewCORSMiddleware())
	}

	// - RequestId是重要的参数，不可关闭；
	headers := s.httpConfig.GetStringSlice(HttpWebServerConfigKeyRequestIdHeaders)
	s.AddWebInterceptor(webmidware.NewRequestIdMiddlewareWithinHeader(headers...))

	// Internal Web Server
	internalPort := s.httpConfig.GetInt(HttpWebServerConfigKeyFeatureDebugPort)
	s.debugServer = &http.Server{
		Handler: http.DefaultServeMux,
		Addr:    fmt.Sprintf("0.0.0.0:%d", internalPort),
	}
	// - Debug特性支持：默认关闭，需要配置开启
	if s.httpConfig.GetBool(HttpWebServerConfigKeyFeatureDebugEnable) {
		http.DefaultServeMux.Handle("/debug/endpoints", NewDebugQueryEndpointHandler(s.bindEndpoint))
		http.DefaultServeMux.Handle("/debug/services", NewDebugQueryServiceHandler())
		http.DefaultServeMux.Handle("/debug/metrics", promhttp.Handler())
	}

	// Endpoint registry
	if registry, config, err := _activeEndpointRegistry(); nil != err {
		return err
	} else {
		if err := s.routerEngine.InitialHook(registry, config); nil != err {
			return err
		}
		s.endpointRegistry = registry
	}
	return s.routerEngine.Initial()
}

func (s *HttpWebServer) Startup(version flux.BuildInfo) error {
	return s.StartServe(version, s.httpConfig)
}

// StartServe server
func (s *HttpWebServer) StartServe(info flux.BuildInfo, config *flux.Configuration) error {
	if err := s.ensure().routerEngine.Startup(); nil != err {
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
	address := fmt.Sprintf("%s:%d", config.GetString("address"), config.GetInt("port"))
	certFile := config.GetString(HttpWebServerConfigKeyTlsCertFile)
	keyFile := config.GetString(HttpWebServerConfigKeyTlsKeyFile)
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
	logger.Infow("HttpWebServer starting", "address", address, "cert", certFile, "key", keyFile)
	return s.webServer.StartTLS(address, certFile, keyFile)
}

func (s *HttpWebServer) HandleEndpointRequest(webc flux.WebContext, mvendpoint *BindEndpoint, tracing bool) error {
	version := webc.HeaderValue(s.httpVersionHeader)
	endpoint, found := mvendpoint.FindByVersion(version)
	requestId := cast.ToString(webc.GetValue(flux.HeaderXRequestId))
	defer func() {
		if r := recover(); r != nil {
			trace := logger.Trace(requestId)
			if err, ok := r.(error); ok {
				trace.Errorw("HttpWebServer panics", "error", err)
			} else {
				trace.Errorw("HttpWebServer panics", "recover", r)
			}
			trace.Error(string(debug.Stack()))
		}
	}()
	if !found {
		if tracing {
			url, _ := webc.RequestURL()
			logger.Trace(requestId).Infow("HttpWebServer route not found",
				"http-pattern", []string{webc.Method(), webc.RequestURI(), url.Path}, "endpoint-version", version,
				"endpoint-service", endpoint.Service.Method+":"+endpoint.Service.Interface,
			)
		}
		return s.serverErrorsWriter(webc, requestId, http.Header{}, ErrEndpointVersionNotFound)
	}
	ctxw := s.acquireContext(requestId, webc, endpoint)
	defer s.releaseContext(ctxw)
	// Context hook
	for _, ctxhook := range s.serverContextExchangeHooks {
		ctxhook(webc, ctxw)
	}
	if tracing {
		url, _ := webc.RequestURL()
		logger.TraceContext(ctxw).Infow("HttpWebServer routing",
			"http-pattern", []string{webc.Method(), webc.RequestURI(), url.Path}, "endpoint-version", version,
			"endpoint-service", endpoint.Service.Method+":"+endpoint.Service.Interface,
		)
	}
	// Route and response
	if err := s.routerEngine.Route(ctxw); nil != err {
		return s.serverErrorsWriter(webc, requestId, ctxw.Response().HeaderValues(), err)
	} else {
		rw := ctxw.Response()
		return s.serverResponseWriter(webc, requestId, rw.HeaderValues(), rw.StatusCode(), rw.Body())
	}
}

func (s *HttpWebServer) HandleBackendServiceEvent(event flux.BackendServiceEvent) {
	service := event.Service
	serviceTag := service.Interface + ":" + service.Method
	switch event.EventType {
	case flux.EventTypeAdded:
		logger.Infow("New service",
			"service-id", service.ServiceId, "service", serviceTag)
		ext.StoreBackendService(service)
	case flux.EventTypeUpdated:
		logger.Infow("Update service",
			"service-id", service.ServiceId, "service", serviceTag)
		ext.StoreBackendService(service)
	case flux.EventTypeRemoved:
		logger.Infow("Delete service",
			"service-id", service.ServiceId, "service", serviceTag)
		ext.RemoveBackendService(service.ServiceId)
	}
}

func (s *HttpWebServer) HandleHttpEndpointEvent(event flux.HttpEndpointEvent) {
	method := strings.ToUpper(event.Endpoint.HttpMethod)
	// Check http method
	if !_isExpectedMethod(method) {
		logger.Warnw("Unsupported http method", "method", method, "pattern", event.Endpoint.HttpPattern)
		return
	}
	pattern := event.Endpoint.HttpPattern
	routeKey := fmt.Sprintf("%s#%s", method, pattern)
	// Refresh endpoint
	endpoint := event.Endpoint
	bind, isreg := s.selectBindEndpoint(routeKey, &endpoint)
	switch event.EventType {
	case flux.EventTypeAdded:
		logger.Infow("New endpoint", "version", endpoint.Version, "method", method, "pattern", pattern)
		bind.Update(endpoint.Version, &endpoint)
		if isreg {
			logger.Infow("Register http handler", "method", method, "pattern", pattern)
			s.webServer.AddWebHandler(method, pattern, s.newWrappedEndpointHandler(bind))
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
func (s *HttpWebServer) Shutdown(ctx context.Context) error {
	logger.Info("HttpWebServer shutdown...")
	defer close(s.stateStopped)
	if s.debugServer != nil {
		_ = s.debugServer.Close()
	}
	if err := s.webServer.Shutdown(ctx); nil != err {
		return err
	}
	return s.routerEngine.Shutdown(ctx)
}

// StateStarted 返回一个Channel。当服务启动完成时，此Channel将被关闭。
func (s *HttpWebServer) StateStarted() <-chan struct{} {
	return s.stateStarted
}

// StateStopped 返回一个Channel。当服务停止后完成时，此Channel将被关闭。
func (s *HttpWebServer) StateStopped() <-chan struct{} {
	return s.stateStopped
}

// HttpConfig return Http server configuration
func (s *HttpWebServer) HttpConfig() *flux.Configuration {
	return s.httpConfig
}

// AddWebInterceptor 添加Http前拦截器。将在Http被路由到对应Handler之前执行
func (s *HttpWebServer) AddWebInterceptor(m flux.WebInterceptor) {
	s.ensure().webServer.AddWebInterceptor(m)
}

// AddWebHandler 添加Http处理接口。
func (s *HttpWebServer) AddWebHandler(method, pattern string, h flux.WebHandler, m ...flux.WebInterceptor) {
	s.ensure().webServer.AddWebHandler(method, pattern, h, m...)
}

// AddWebHttpHandler 添加Http处理接口。
func (s *HttpWebServer) AddWebHttpHandler(method, pattern string, h http.Handler, m ...func(http.Handler) http.Handler) {
	s.ensure().webServer.AddWebHttpHandler(method, pattern, h, m...)
}

// SetWebNotFoundHandler 设置Http路由失败的处理接口
func (s *HttpWebServer) SetWebNotFoundHandler(nfh flux.WebHandler) {
	s.ensure().webServer.SetWebNotFoundHandler(nfh)
}

// WebServer 返回WebServer实例
func (s *HttpWebServer) WebServer() flux.WebServer {
	return s.ensure().webServer
}

// DebugServer 返回DebugServer实例，以及实体是否有效
func (s *HttpWebServer) DebugServer() (*http.Server, bool) {
	return s.debugServer, nil != s.debugServer
}

// SetServerResponseWriter 设置Http响应数据写入的处理接口
func (s *HttpWebServer) SetServerResponseWriter(writer flux.ServerResponseWriter) {
	s.serverResponseWriter = writer
}

// SetServerErrorsWriter 设置Http响应异常消息写入的处理接口
func (s *HttpWebServer) SetServerErrorsWriter(writer flux.ServerErrorsWriter) {
	s.serverErrorsWriter = writer
}

// AddServerContextExchangeHook 添加Http与Flux的Context桥接函数
func (s *HttpWebServer) AddServerContextExchangeHook(f flux.ServerContextExchangeHook) {
	s.serverContextExchangeHooks = append(s.serverContextExchangeHooks, f)
}

func (s *HttpWebServer) newWrappedEndpointHandler(endpoint *BindEndpoint) flux.WebHandler {
	enabled := s.httpConfig.GetBool(HttpWebServerConfigKeyRequestLogEnable)
	return func(webc flux.WebContext) error {
		return s.HandleEndpointRequest(webc, endpoint, enabled)
	}
}

func (s *HttpWebServer) selectBindEndpoint(routeKey string, endpoint *flux.Endpoint) (*BindEndpoint, bool) {
	if be, ok := s.bindEndpoint[routeKey]; ok {
		return be, false
	} else {
		be = NewBindEndpoint(endpoint)
		s.bindEndpoint[routeKey] = be
		return be, true
	}
}

func (s *HttpWebServer) acquireContext(id string, webc flux.WebContext, endpoint *flux.Endpoint) *WrappedContext {
	ctx := s.contextWrappers.Get().(*WrappedContext)
	ctx.Reattach(id, webc, endpoint)
	return ctx
}

func (s *HttpWebServer) releaseContext(context *WrappedContext) {
	context.Release()
	s.contextWrappers.Put(context)
}

func (s *HttpWebServer) ensure() *HttpWebServer {
	if s.webServer == nil {
		logger.Panicf("Call must after InitialServer()")
	}
	return s
}

func (s *HttpWebServer) handleNotFoundError(webc flux.WebContext) error {
	return &flux.StateError{
		StatusCode: flux.StatusNotFound,
		ErrorCode:  flux.ErrorCodeRequestNotFound,
		Message:    "ROUTE:NOT_FOUND",
	}
}

func (s *HttpWebServer) handleServerError(err error, webc flux.WebContext) {
	// Http中间件等返回InvokeError错误
	stateError, ok := err.(*flux.StateError)
	if !ok {
		stateError = &flux.StateError{
			StatusCode: flux.StatusServerError,
			ErrorCode:  flux.ErrorCodeGatewayInternal,
			Message:    err.Error(),
			Internal:   err,
		}
	}
	requestId := cast.ToString(webc.GetValue(flux.HeaderXRequestId))
	if err := s.serverErrorsWriter(webc, requestId, http.Header{}, stateError); nil != err {
		logger.Trace(requestId).Errorw("Server http response error", "error", err)
	}
}

func _activeEndpointRegistry() (flux.EndpointRegistry, *flux.Configuration, error) {
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

func _isExpectedMethod(method string) bool {
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
