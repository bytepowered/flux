package server

import (
	"context"
	"fmt"
	dubgo "github.com/apache/dubbo-go/config"
	"github.com/bytepowered/flux"
	context2 "github.com/bytepowered/flux/context"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/webserver"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cast"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

const (
	defaultBanner = "Flux-GO // Fast gateway for microservice: dubbo, grpc, http"
	VersionFormat = "Version // git.commit=%s, build.version=%s, build.date=%s"
)

const (
	DefaultHttpHeaderVersion = "X-Version"
)

const (
	HttpWebServerConfigRootName              = "http_web_server"
	HttpWebServerConfigKeyFeatureDebugEnable = "feature_debug_enable"
	HttpWebServerConfigKeyFeatureDebugPort   = "feature-debug-port"
	HttpWebServerConfigKeyRequestLogEnable   = "request_log_enable"
	HttpWebServerConfigKeyAddress            = "address"
	HttpWebServerConfigKeyPort               = "port"
	HttpWebServerConfigKeyTlsCertFile        = "tls_cert_file"
	HttpWebServerConfigKeyTlsKeyFile         = "tls_key_file"
)

type (
	// Option 配置HttpServeEngine函数
	Option func(engine *AppServer)
	// VersionLookupFunc Http请求版本查找函数
	VersionLookupFunc func(webc flux.WebContext) string
)

// AppServer
type AppServer struct {
	httpWebServer  flux.ListenServer
	responseWriter flux.ServerResponseWriter
	errorsWriter   flux.ServerErrorsWriter
	ctxHooks       []flux.ServerContextHookFunc
	interceptors   []flux.WebInterceptor
	debugServer    *http.Server
	config         *flux.Configuration
	defaults       map[string]interface{}
	router         *AppRouter
	versionLookup  VersionLookupFunc
	ctxPool        sync.Pool
	started        chan struct{}
	stopped        chan struct{}
	banner         string
}

// WithServerResponseWriter 用于配置Web服务响应数据输出函数
func WithServerResponseWriter(writer flux.ServerResponseWriter) Option {
	return func(engine *AppServer) {
		engine.responseWriter = writer
	}
}

// WithServerErrorsWriter 用于配置Web服务错误输出响应数据函数
func WithServerErrorsWriter(writer flux.ServerErrorsWriter) Option {
	return func(engine *AppServer) {
		engine.errorsWriter = writer
	}
}

// WithServerContextHooks 配置请求Hook函数列表
func WithServerContextHooks(hooks ...flux.ServerContextHookFunc) Option {
	return func(engine *AppServer) {
		engine.ctxHooks = append(engine.ctxHooks, hooks...)
	}
}

// WithServerWebInterceptors 配置Web服务拦截器列表
func WithServerWebInterceptors(wis ...flux.WebInterceptor) Option {
	return func(engine *AppServer) {
		engine.interceptors = append(engine.interceptors, wis...)
	}
}

// WithServerWebVersionLookupFunc 配置Web请求版本选择函数
func WithServerWebVersionLookupFunc(fun VersionLookupFunc) Option {
	return func(engine *AppServer) {
		engine.versionLookup = fun
	}
}

// WithServerDefaults 配置Web服务默认配置
func WithServerDefaults(defaults map[string]interface{}) Option {
	return func(engine *AppServer) {
		engine.defaults = defaults
	}
}

// WithBanner 配置服务Banner
func WithServerBanner(banner string) Option {
	return func(engine *AppServer) {
		engine.banner = banner
	}
}

// WithPrepareHooks 配置服务启动预备阶段Hook函数列表
func WithPrepareHooks(hooks ...flux.PrepareHookFunc) Option {
	return func(engine *AppServer) {
		engine.router.hooks = append(engine.router.hooks, hooks...)
	}
}

func NewHttpServeEngine() *AppServer {
	return NewHttpServeEngineOverride()
}

func NewHttpServeEngineOverride(overrides ...Option) *AppServer {
	opts := []Option{WithServerBanner(defaultBanner),
		WithServerResponseWriter(DefaultServerResponseWriter),
		WithServerErrorsWriter(DefaultServerErrorsWriter),
		WithServerWebInterceptors(
			webserver.NewCORSInterceptor(),
			webserver.NewRequestIdInterceptor(),
		),
		WithServerWebVersionLookupFunc(func(webc flux.WebContext) string {
			return webc.HeaderVar(DefaultHttpHeaderVersion)
		}),
		WithServerDefaults(map[string]interface{}{
			HttpWebServerConfigKeyFeatureDebugEnable: false,
			HttpWebServerConfigKeyFeatureDebugPort:   9527,
			HttpWebServerConfigKeyAddress:            "0.0.0.0",
			HttpWebServerConfigKeyPort:               8080,
		})}
	return NewAppServerWith(context2.DefaultContextFactory, append(opts, overrides...)...)
}

func NewAppServerWith(factory func() flux.Context, opts ...Option) *AppServer {
	srv := &AppServer{
		router:         NewAppRouter(),
		responseWriter: DefaultServerResponseWriter,
		errorsWriter:   DefaultServerErrorsWriter,
		ctxPool:        sync.Pool{New: func() interface{} { return factory() }},
		ctxHooks:       make([]flux.ServerContextHookFunc, 0, 4),
		interceptors:   make([]flux.WebInterceptor, 0, 4),
		started:        make(chan struct{}),
		stopped:        make(chan struct{}),
		banner:         defaultBanner,
	}
	for _, opt := range opts {
		opt(srv)
	}
	return srv
}

// Prepare Call before init and startup
func (s *AppServer) Prepare() error {
	return s.router.Prepare()
}

// Initial
func (s *AppServer) Initial() error {
	// Http server
	s.config = flux.NewConfigurationOfNS(HttpWebServerConfigRootName)
	s.config.SetDefaults(s.defaults)
	// 创建WebServer
	s.httpWebServer = ext.GetWebServerFactory()(s.config)
	// 默认必备的WebServer功能
	s.httpWebServer.SetErrorHandler(s.defaultServerErrorHandler)
	s.httpWebServer.SetNotfoundHandler(s.defaultNotFoundErrorHandler)
	// 第一优先级的拦截器
	for _, wi := range s.interceptors {
		s.AddWebInterceptor(wi)
	}
	// Internal Web Server
	port := s.config.GetInt(HttpWebServerConfigKeyFeatureDebugPort)
	s.debugServer = &http.Server{
		Handler: http.DefaultServeMux,
		Addr:    fmt.Sprintf("0.0.0.0:%d", port),
	}
	// - Debug特性支持：默认关闭，需要配置开启
	if s.config.GetBool(HttpWebServerConfigKeyFeatureDebugEnable) {
		http.DefaultServeMux.Handle("/debug/endpoints", NewDebugQueryEndpointHandler())
		http.DefaultServeMux.Handle("/debug/services", NewDebugQueryServiceHandler())
		http.DefaultServeMux.Handle("/debug/metrics", promhttp.Handler())
	}
	// Endpoint discovery
	for _, discovery := range ext.GetEndpointDiscoveries() {
		ns := flux.NamespaceEndpointDiscoveryServices + "." + discovery.Id()
		if err := s.router.InitialHook(discovery, flux.NewConfigurationOfNS(ns)); nil != err {
			return err
		}
	}
	return s.router.Initial()
}

func (s *AppServer) Startup(build flux.Build) error {
	logger.Infof(VersionFormat, build.CommitId, build.Version, build.Date)
	if "" != s.banner {
		logger.Info(s.banner)
	}
	return s.start(s.config)
}

func (s *AppServer) start(config *flux.Configuration) error {
	s.ensure()
	if err := s.router.Startup(); nil != err {
		return err
	}
	endpoints := make(chan flux.HttpEndpointEvent, 2)
	services := make(chan flux.BackendServiceEvent, 2)
	defer func() {
		close(endpoints)
		close(services)
	}()
	if err := s.startDiscovery(endpoints, services); nil != err {
		return err
	}
	// Start Servers
	if s.debugServer != nil {
		go func() {
			logger.Infow("DebugServer starting", "address", s.debugServer.Addr)
			_ = s.debugServer.ListenAndServe()
		}()
	}
	address := fmt.Sprintf("%s:%d",
		config.GetString(HttpWebServerConfigKeyAddress), config.GetInt(HttpWebServerConfigKeyPort))
	keyFile := config.GetString(HttpWebServerConfigKeyTlsKeyFile)
	certFile := config.GetString(HttpWebServerConfigKeyTlsCertFile)
	logger.Infow("Server starting", "address", address, "cert", certFile, "key", keyFile)
	close(s.started)
	return s.httpWebServer.Listen(address, certFile, keyFile)
}

func (s *AppServer) startDiscovery(endpoints chan flux.HttpEndpointEvent, services chan flux.BackendServiceEvent) error {
	for _, discovery := range ext.GetEndpointDiscoveries() {
		if err := discovery.WatchEndpoints(endpoints); nil != err {
			return err
		}
		if err := discovery.WatchServices(services); nil != err {
			return err
		}
	}
	go func() {
		logger.Info("Discovery event loop: START")
	Loop:
		for {
			select {
			case epEvt, ok := <-endpoints:
				if !ok {
					break Loop
				}
				s.HandleHttpEndpointEvent(epEvt)

			case esEvt, ok := <-services:
				if !ok {
					break Loop
				}
				s.HandleBackendServiceEvent(esEvt)
			}
		}
		logger.Info("Discovery event loop: STOP")
	}()
	return nil
}

func (s *AppServer) HandleEndpointRequest(webc flux.WebContext, endpoints *MultiEndpoint, tracing bool) error {
	version := s.versionLookup(webc)
	endpoint, found := endpoints.LookupByVersion(version)
	requestId := cast.ToString(webc.GetValue(flux.HeaderXRequestId))
	defer func() {
		if r := recover(); r != nil {
			trace := logger.With(requestId)
			if err, ok := r.(error); ok {
				trace.Errorw("CRITICAL:SERVER_PANIC", "error", err)
			} else {
				trace.Errorw("CRITICAL:SERVER_PANIC", "recover", r)
			}
			trace.Error(string(debug.Stack()))
		}
	}()
	if !found {
		if tracing {
			logger.With(requestId).Infow("Server route not-found",
				"http-pattern", []string{webc.Method(), webc.URI(), webc.URL().Path},
			)
		}
		return flux.ErrRouteNotFound
	}
	ctxw := s.acquireContext(requestId, webc, endpoint)
	defer s.releaseContext(ctxw)
	// Route call
	logger.WithContext(ctxw).Infow("Server route start")
	endcall := func(code int, start time.Time) {
		logger.WithContext(ctxw).Infow("Server route end",
			"metric", ctxw.Metrics(),
			"elapses", time.Since(start).String(), "response.code", code)
	}
	start := time.Now()
	// Context hook
	for _, hook := range s.ctxHooks {
		hook(webc, ctxw)
	}
	// Route and response
	if err := s.router.Route(ctxw); nil != err {
		defer endcall(err.StatusCode, start)
		logger.WithContext(ctxw).Errorw("Server route error", "error", err)
		err.MergeHeader(ctxw.Response().HeaderVars())
		return err
	} else {
		response := ctxw.Response()
		defer endcall(response.StatusCode(), start)
		return s.responseWriter(webc, requestId, response.HeaderVars(), response.StatusCode(), response.Payload())
	}
}

func (s *AppServer) HandleBackendServiceEvent(event flux.BackendServiceEvent) {
	service := event.Service
	initArguments(service.Arguments)
	switch event.EventType {
	case flux.EventTypeAdded:
		logger.Infow("Add service",
			"service-id", service.ServiceId, "alias-id", service.AliasId)
		ext.SetBackendService(service)
		if "" != service.AliasId {
			ext.SetBackendServiceById(service.AliasId, service)
		}
	case flux.EventTypeUpdated:
		logger.Infow("Update service",
			"service-id", service.ServiceId, "alias-id", service.AliasId)
		ext.SetBackendService(service)
		if "" != service.AliasId {
			ext.SetBackendServiceById(service.AliasId, service)
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

func (s *AppServer) HandleHttpEndpointEvent(event flux.HttpEndpointEvent) {
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
	initArguments(endpoint.Service.Arguments)
	initArguments(endpoint.Permission.Arguments)
	bind, isreg := s.selectMultiEndpoint(routeKey, &endpoint)
	switch event.EventType {
	case flux.EventTypeAdded:
		logger.Infow("Add endpoint", "version", endpoint.Version, "method", method, "pattern", pattern)
		bind.Update(endpoint.Version, &endpoint)
		if isreg {
			logger.Infow("Register http handler", "method", method, "pattern", pattern)
			s.httpWebServer.AddHandler(method, pattern, s.newWrappedEndpointHandler(bind))
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
func (s *AppServer) Shutdown(ctx context.Context) error {
	logger.Info("Server shutdown...")
	defer close(s.stopped)
	if s.debugServer != nil {
		_ = s.debugServer.Close()
	}
	if err := s.httpWebServer.Close(ctx); nil != err {
		logger.Warnw("Server shutdown http server", "error", err)
	}
	return s.router.Shutdown(ctx)
}

// GracefulShutdown
func (s *AppServer) OnSignalShutdown(quit chan os.Signal, to time.Duration) {
	// 接收停止信号
	signal.Notify(quit, dubgo.ShutdownSignals...)
	<-quit
	logger.Infof("Server received shutdown signal, shutdown...")
	ctx, cancel := context.WithTimeout(context.Background(), to)
	defer cancel()
	if err := s.Shutdown(ctx); nil != err {
		logger.Error("Server shutdown, error: ", err)
	}
}

// StateStarted 返回一个Channel。当服务启动完成时，此Channel将被关闭。
func (s *AppServer) StateStarted() <-chan struct{} {
	return s.started
}

// StateStopped 返回一个Channel。当服务停止后完成时，此Channel将被关闭。
func (s *AppServer) StateStopped() <-chan struct{} {
	return s.stopped
}

// HttpConfig return Http server configuration
func (s *AppServer) HttpConfig() *flux.Configuration {
	return s.config
}

// AddWebInterceptor 添加Http前拦截器。将在Http被路由到对应Handler之前执行
func (s *AppServer) AddWebInterceptor(m flux.WebInterceptor) {
	s.ensure().httpWebServer.AddInterceptor(m)
}

// AddWebHandler 添加Http处理接口。
func (s *AppServer) AddWebHandler(method, pattern string, h flux.WebHandler, m ...flux.WebInterceptor) {
	s.ensure().httpWebServer.AddHandler(method, pattern, h, m...)
}

// AddWebHttpHandler 添加Http处理接口。
func (s *AppServer) AddWebHttpHandler(method, pattern string, h http.Handler, m ...func(http.Handler) http.Handler) {
	s.ensure().httpWebServer.AddHttpHandler(method, pattern, h, m...)
}

// SetWebNotfoundHandler 设置Http路由失败的处理接口
func (s *AppServer) SetWebNotfoundHandler(nfh flux.WebHandler) {
	s.ensure().httpWebServer.SetNotfoundHandler(nfh)
}

// WebServer 返回WebServer实例
func (s *AppServer) WebServer() flux.ListenServer {
	return s.ensure().httpWebServer
}

// DebugServer 返回DebugServer实例，以及实体是否有效
func (s *AppServer) DebugServer() (*http.Server, bool) {
	return s.debugServer, nil != s.debugServer
}

// AddServerContextHookFunc 添加Http与Flux的Context桥接函数
func (s *AppServer) AddServerContextHookFunc(f flux.ServerContextHookFunc) {
	s.ctxHooks = append(s.ctxHooks, f)
}

func (s *AppServer) newWrappedEndpointHandler(endpoint *MultiEndpoint) flux.WebHandler {
	enabled := s.config.GetBool(HttpWebServerConfigKeyRequestLogEnable)
	return func(webc flux.WebContext) error {
		return s.HandleEndpointRequest(webc, endpoint, enabled)
	}
}

func (s *AppServer) selectMultiEndpoint(routeKey string, endpoint *flux.Endpoint) (*MultiEndpoint, bool) {
	if mve, ok := SelectMultiEndpoint(routeKey); ok {
		return mve, false
	} else {
		return RegisterMultiEndpoint(routeKey, endpoint), true
	}
}

func (s *AppServer) acquireContext(id string, webc flux.WebContext, endpoint *flux.Endpoint) *context2.DefaultContext {
	ctx := s.ctxPool.Get().(*context2.DefaultContext)
	ctx.Reattach(id, webc, endpoint)
	return ctx
}

func (s *AppServer) releaseContext(context *context2.DefaultContext) {
	context.Release()
	s.ctxPool.Put(context)
}

func (s *AppServer) ensure() *AppServer {
	if s.httpWebServer == nil {
		logger.Panicf("Call must after InitialServer()")
	}
	return s
}

func (s *AppServer) defaultNotFoundErrorHandler(webc flux.WebContext) error {
	return &flux.ServeError{
		StatusCode: flux.StatusNotFound,
		ErrorCode:  flux.ErrorCodeRequestNotFound,
		Message:    flux.ErrorMessageWebServerRequestNotFound,
	}
}

func (s *AppServer) defaultServerErrorHandler(err error, webc flux.WebContext) {
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
	if err := s.errorsWriter(webc, requestId, serve.Header, serve); nil != err {
		logger.With(requestId).Errorw("Server http response error", "error", err)
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

func initArguments(args []flux.Argument) {
	for i := range args {
		args[i].ValueResolver = ext.GetMTValueResolver(args[i].Class)
		args[i].LookupFunc = ext.GetArgumentLookupFunc()
		initArguments(args[i].Fields)
	}
}
