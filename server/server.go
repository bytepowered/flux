package server

import (
	"context"
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/support"
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
	HttpServerConfigRootName              = "HttpServer"
	HttpServerConfigKeyFeatureDebugEnable = "feature-debug-enable"
	HttpServerConfigKeyFeatureDebugPort   = "feature-debug-port"
	HttpServerConfigKeyFeatureCorsEnable  = "feature-cors-enable"
	HttpServerConfigKeyVersionHeader      = "version-header"
	HttpServerConfigKeyRequestIdHeaders   = "requestReader-id-headers"
	HttpServerConfigKeyRequestLogEnable   = "requestReader-log-enable"
	HttpServerConfigKeyAddress            = "address"
	HttpServerConfigKeyPort               = "port"
	HttpServerConfigKeyTlsCertFile        = "tls-cert-file"
	HttpServerConfigKeyTlsKeyFile         = "tls-key-file"
)

var (
	ErrEndpointVersionNotFound = &flux.StateError{
		StatusCode: flux.StatusNotFound,
		ErrorCode:  flux.ErrorCodeGatewayEndpoint,
		Message:    "ENDPOINT_VERSION_NOT_FOUND",
	}
)

var (
	HttpServerConfigDefaults = map[string]interface{}{
		HttpServerConfigKeyVersionHeader:      DefaultHttpHeaderVersion,
		HttpServerConfigKeyFeatureDebugEnable: false,
		HttpServerConfigKeyFeatureDebugPort:   9527,
		HttpServerConfigKeyAddress:            "0.0.0.0",
		HttpServerConfigKeyPort:               8080,
	}
)

// Server
type HttpServer struct {
	webServer                 flux.WebServer
	serverResponseWriter      flux.ServerResponseWriter
	serverErrorsWriter        flux.ServerErrorsWriter
	serverContextExchangeFuns []flux.ServerContextExchangeFunc
	debugServer               *http.Server
	httpConfig                *flux.Configuration
	httpVersionHeader         string
	routerEngine              *RouterEngine
	endpointRegistry          flux.EndpointRegistry
	mvEndpointMap             map[string]*support.MultiVersionEndpoint
	contextWrappers           sync.Pool
	stateStarted              chan struct{}
	stateStopped              chan struct{}
}

func NewHttpServer() *HttpServer {
	return &HttpServer{
		serverResponseWriter:      DefaultServerResponseWriter,
		serverErrorsWriter:        DefaultServerErrorsWriter,
		routerEngine:              NewRouteEngine(),
		mvEndpointMap:             make(map[string]*support.MultiVersionEndpoint),
		contextWrappers:           sync.Pool{New: NewContextWrapper},
		serverContextExchangeFuns: make([]flux.ServerContextExchangeFunc, 0, 4),
		stateStarted:              make(chan struct{}),
		stateStopped:              make(chan struct{}),
	}
}

// Prepare Call before init and startup
func (s *HttpServer) Prepare(hooks ...flux.PrepareHookFunc) error {
	for _, prepare := range append(ext.GetPrepareHooks(), hooks...) {
		if err := prepare(); nil != err {
			return err
		}
	}
	return nil
}

// Initial
func (s *HttpServer) Initial() error {
	// Http server
	s.httpConfig = flux.NewConfigurationOf(HttpServerConfigRootName)
	s.httpConfig.SetDefaults(HttpServerConfigDefaults)
	s.httpVersionHeader = s.httpConfig.GetString(HttpServerConfigKeyVersionHeader)
	// 创建WebServer
	s.webServer = ext.GetWebServerFactory()()
	// 默认必备的WebServer功能
	s.webServer.SetWebErrorHandler(s.handleServerError)
	s.webServer.SetWebNotFoundHandler(s.handleNotFoundError)

	// - 请求CORS跨域支持：默认关闭，需要配置开启
	if s.httpConfig.GetBool(HttpServerConfigKeyFeatureCorsEnable) {
		s.AddWebInterceptor(webmidware.NewCORSMiddleware())
	}

	// - RequestId是重要的参数，不可关闭；
	headers := s.httpConfig.GetStringSlice(HttpServerConfigKeyRequestIdHeaders)
	s.AddWebInterceptor(webmidware.NewRequestIdMiddlewareWithinHeader(headers...))

	// - Debug特性支持：默认关闭，需要配置开启
	if s.httpConfig.GetBool(HttpServerConfigKeyFeatureDebugEnable) {
		servemux := http.DefaultServeMux
		s.debugServer = &http.Server{
			Handler: servemux,
			Addr:    fmt.Sprintf("0.0.0.0:%d", s.httpConfig.GetInt(HttpServerConfigKeyFeatureDebugPort)),
		}
		servemux.Handle("/debug/endpoints", DebugQueryEndpoint(s.mvEndpointMap))
		servemux.Handle("/debug/metrics", promhttp.Handler())
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

func (s *HttpServer) Startup(version flux.BuildInfo) error {
	return s.StartServe(version, s.httpConfig)
}

// StartServe server
func (s *HttpServer) StartServe(info flux.BuildInfo, config *flux.Configuration) error {
	if err := s.ensure().routerEngine.Startup(); nil != err {
		return err
	}
	// Watch endpoint register
	if events, err := s.endpointRegistry.WatchEvents(); nil != err {
		return fmt.Errorf("start registry watching: %w", err)
	} else {
		go func() {
			logger.Info("Endpoint event loop: starting")
			for event := range events {
				s.HandleEndpointEvent(event)
			}
			logger.Info("Endpoint event loop: Stopped")
		}()
	}
	address := fmt.Sprintf("%s:%d", config.GetString("address"), config.GetInt("port"))
	certFile := config.GetString(HttpServerConfigKeyTlsCertFile)
	keyFile := config.GetString(HttpServerConfigKeyTlsKeyFile)
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
	logger.Infow("HttpServer starting", "address", address, "cert", certFile, "key", keyFile)
	return s.webServer.StartTLS(address, certFile, keyFile)
}

// Shutdown to cleanup resources
func (s *HttpServer) Shutdown(ctx context.Context) error {
	logger.Info("HttpServer shutdown...")
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
func (s *HttpServer) StateStarted() <-chan struct{} {
	return s.stateStarted
}

// StateStopped 返回一个Channel。当服务停止后完成时，此Channel将被关闭。
func (s *HttpServer) StateStopped() <-chan struct{} {
	return s.stateStopped
}

// HttpConfig return Http server configuration
func (s *HttpServer) HttpConfig() *flux.Configuration {
	return s.httpConfig
}

// AddWebInterceptor 添加Http前拦截器。将在Http被路由到对应Handler之前执行
func (s *HttpServer) AddWebInterceptor(m flux.WebInterceptor) {
	s.ensure().webServer.AddWebInterceptor(m)
}

// AddWebHandler 添加Http处理接口。
func (s *HttpServer) AddWebHandler(method, pattern string, h flux.WebHandler, m ...flux.WebInterceptor) {
	s.ensure().webServer.AddWebHandler(method, pattern, h, m...)
}

// AddWebHttpHandler 添加Http处理接口。
func (s *HttpServer) AddWebHttpHandler(method, pattern string, h http.Handler, m ...func(http.Handler) http.Handler) {
	s.ensure().webServer.AddWebHttpHandler(method, pattern, h, m...)
}

// SetWebNotFoundHandler 设置Http路由失败的处理接口
func (s *HttpServer) SetWebNotFoundHandler(nfh flux.WebHandler) {
	s.ensure().webServer.SetWebNotFoundHandler(nfh)
}

// RawWebServer 返回WebServer实例
func (s *HttpServer) WebServer() flux.WebServer {
	return s.ensure().webServer
}

// DebugServer 返回DebugServer实例，以及实体是否有效
func (s *HttpServer) DebugServer() (*http.Server, bool) {
	return s.debugServer, nil != s.debugServer
}

// SetServerResponseWriter 设置Http响应数据写入的处理接口
func (s *HttpServer) SetServerResponseWriter(writer flux.ServerResponseWriter) {
	s.serverResponseWriter = writer
}

// SetServerErrorsWriter 设置Http响应异常消息写入的处理接口
func (s *HttpServer) SetServerErrorsWriter(writer flux.ServerErrorsWriter) {
	s.serverErrorsWriter = writer
}

// AddServerContextExchangeFunc 添加Http与Flux的Context桥接函数
func (s *HttpServer) AddServerContextExchangeFunc(f flux.ServerContextExchangeFunc) {
	s.serverContextExchangeFuns = append(s.serverContextExchangeFuns, f)
}

func (s *HttpServer) HandleEndpointRequest(webc flux.WebContext, mvendpoint *support.MultiVersionEndpoint, tracing bool) error {
	version := webc.HeaderValue(s.httpVersionHeader)
	endpoint, found := mvendpoint.FindByVersion(version)
	requestId := cast.ToString(webc.GetValue(flux.HeaderXRequestId))
	defer func() {
		if err := recover(); err != nil {
			tl := logger.Trace(requestId)
			tl.Errorw("Server dispatch: unexpected error", "error", err)
			tl.Error(string(debug.Stack()))
		}
	}()
	if !found {
		if tracing {
			requrl, _ := webc.RequestURL()
			logger.Trace(requestId).Infow("HttpServer routing: ENDPOINT_NOT_FOUND",
				"method", webc.Method(), "uri", webc.RequestURI(), "path", requrl.Path, "version", version,
			)
		}
		return s.serverErrorsWriter(webc, requestId, http.Header{}, ErrEndpointVersionNotFound)
	}
	ctxw := s.acquireContext(requestId, webc, endpoint)
	defer s.releaseContext(ctxw)
	// Context hook
	for _, ctxex := range s.serverContextExchangeFuns {
		ctxex(webc, ctxw)
	}
	if tracing {
		requrl, _ := webc.RequestURL()
		logger.TraceContext(ctxw).Infow("HttpServer routing: DISPATCHING",
			"method", webc.Method(), "uri", webc.RequestURI(), "path", requrl.Path, "version", version,
			"endpoint", endpoint.UpstreamMethod+":"+endpoint.UpstreamUri,
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

func (s *HttpServer) HandleEndpointEvent(event flux.EndpointEvent) {
	routeMethod := strings.ToUpper(event.Endpoint.HttpMethod)
	// Check http method
	if !_isExpectedMethod(routeMethod) {
		logger.Warnw("Unsupported http method", "method", routeMethod, "pattern", event.HttpPattern)
		return
	}
	pattern := event.HttpPattern
	routeKey := fmt.Sprintf("%s#%s", routeMethod, pattern)
	// Refresh endpoint
	endpoint := event.Endpoint
	_initializeArguments(endpoint.Arguments)
	if endpoint.Permission.IsValid() {
		_initializeArguments(endpoint.Permission.Arguments)
	}
	multi, isRegister := s.selectMultiEndpoint(routeKey, &endpoint)
	switch event.EventType {
	case flux.EndpointEventAdded:
		logger.Infow("New endpoint", "version", endpoint.Version, "method", routeMethod, "pattern", pattern)
		multi.Update(endpoint.Version, &endpoint)
		if isRegister {
			logger.Infow("Register http router", "method", routeMethod, "pattern", pattern)
			s.webServer.AddWebHandler(routeMethod, pattern, s.newWrappedEndpointHandler(multi))
		}
	case flux.EndpointEventUpdated:
		logger.Infow("Update endpoint", "version", endpoint.Version, "method", routeMethod, "pattern", pattern)
		multi.Update(endpoint.Version, &endpoint)
	case flux.EndpointEventRemoved:
		logger.Infow("Delete endpoint", "method", routeMethod, "pattern", pattern)
		multi.Delete(endpoint.Version)
	}
}

func (s *HttpServer) newWrappedEndpointHandler(endpoint *support.MultiVersionEndpoint) flux.WebHandler {
	enabled := s.httpConfig.GetBool(HttpServerConfigKeyRequestLogEnable)
	return func(webc flux.WebContext) error {
		return s.HandleEndpointRequest(webc, endpoint, enabled)
	}
}

func (s *HttpServer) selectMultiEndpoint(routeKey string, endpoint *flux.Endpoint) (*support.MultiVersionEndpoint, bool) {
	if mve, ok := s.mvEndpointMap[routeKey]; ok {
		return mve, false
	} else {
		mve = support.NewMultiVersionEndpoint(endpoint)
		s.mvEndpointMap[routeKey] = mve
		return mve, true
	}
}

func (s *HttpServer) acquireContext(id string, webc flux.WebContext, endpoint *flux.Endpoint) *WrappedContext {
	ctx := s.contextWrappers.Get().(*WrappedContext)
	ctx.Reattach(id, webc, endpoint)
	return ctx
}

func (s *HttpServer) releaseContext(context *WrappedContext) {
	context.Release()
	s.contextWrappers.Put(context)
}

func (s *HttpServer) ensure() *HttpServer {
	if s.webServer == nil {
		logger.Panicf("Call must after InitialServer()")
	}
	return s
}

func (s *HttpServer) handleNotFoundError(webc flux.WebContext) error {
	return &flux.StateError{
		StatusCode: flux.StatusNotFound,
		ErrorCode:  flux.ErrorCodeRequestNotFound,
		Message:    "ROUTE:NOT_FOUND",
	}
}

func (s *HttpServer) handleServerError(err error, webc flux.WebContext) {
	// Http中间件等返回InvokeError错误
	serr, ok := err.(*flux.StateError)
	if !ok {
		serr = &flux.StateError{
			StatusCode: flux.StatusServerError,
			ErrorCode:  flux.ErrorCodeGatewayInternal,
			Message:    err.Error(),
			Internal:   err,
		}
	}
	requestId := cast.ToString(webc.GetValue(flux.HeaderXRequestId))
	if err := s.serverErrorsWriter(webc, requestId, http.Header{}, serr); nil != err {
		logger.Trace(requestId).Errorw("Server http responseWriter error", "error", err)
	}
}

func _initializeArguments(args []flux.Argument) {
	for i := range args {
		_initializeArgument(&args[i])
	}
}

func _initializeArgument(arg *flux.Argument) {
	arg.HttpValue = flux.NewWrapValue(nil)
	for i := range arg.Fields {
		_initializeArgument(&arg.Fields[i])
	}
}

func _activeEndpointRegistry() (flux.EndpointRegistry, *flux.Configuration, error) {
	config := flux.NewConfigurationOf(flux.KeyConfigRootEndpointRegistry)
	config.SetDefault(flux.KeyConfigEndpointRegistryId, ext.EndpointRegistryIdDefault)
	registryId := config.GetString(flux.KeyConfigEndpointRegistryId)
	logger.Infow("Active router registry", "registry-id", registryId)
	if factory, ok := ext.GetEndpointRegistryFactory(registryId); !ok {
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
