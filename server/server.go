package server

import (
	"context"
	"fmt"
	"github.com/bytepowered/flux"
	context2 "github.com/bytepowered/flux/context"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/webserver"
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
	defaultBanner = "Flux-GO // Fast gateway for microservice: dubbo, grpc, http"
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
	HttpWebServerConfigKeyRequestLogEnable   = "request-log-enable"
	HttpWebServerConfigKeyAddress            = "address"
	HttpWebServerConfigKeyPort               = "port"
	HttpWebServerConfigKeyTlsCertFile        = "tls-cert-file"
	HttpWebServerConfigKeyTlsKeyFile         = "tls-key-file"
)

type (
	// Option 配置HttpServeEngine函数
	Option func(engine *HttpServeEngine)
	// VersionLookupFunc Http请求版本查找函数
	VersionLookupFunc func(webc flux.WebContext) string
)

// ServeEngine
type HttpServeEngine struct {
	httpWebServer  flux.WebServer
	responseWriter flux.ServerResponseWriter
	errorsWriter   flux.ServerErrorsWriter
	ctxHooks       []flux.ServerContextHookFunc
	interceptors   []flux.WebInterceptor
	debugServer    *http.Server
	config         *flux.Configuration
	defaults       map[string]interface{}
	router         *Router
	versionLookup  VersionLookupFunc
	ctxPool        sync.Pool
	started        chan struct{}
	stopped        chan struct{}
	banner         string
}

// WithServerResponseWriter 用于配置Web服务响应数据输出函数
func WithServerResponseWriter(writer flux.ServerResponseWriter) Option {
	return func(engine *HttpServeEngine) {
		engine.responseWriter = writer
	}
}

// WithServerErrorsWriter 用于配置Web服务错误输出响应数据函数
func WithServerErrorsWriter(writer flux.ServerErrorsWriter) Option {
	return func(engine *HttpServeEngine) {
		engine.errorsWriter = writer
	}
}

// WithServerContextHooks 配置请求Hook函数列表
func WithServerContextHooks(hooks ...flux.ServerContextHookFunc) Option {
	return func(engine *HttpServeEngine) {
		engine.ctxHooks = append(engine.ctxHooks, hooks...)
	}
}

// WithServerWebInterceptors 配置Web服务拦截器列表
func WithServerWebInterceptors(wis ...flux.WebInterceptor) Option {
	return func(engine *HttpServeEngine) {
		engine.interceptors = append(engine.interceptors, wis...)
	}
}

// WithServerWebVersionLookupFunc 配置Web请求版本选择函数
func WithServerWebVersionLookupFunc(fun VersionLookupFunc) Option {
	return func(engine *HttpServeEngine) {
		engine.versionLookup = fun
	}
}

// WithServerDefaults 配置Web服务默认配置
func WithServerDefaults(defaults map[string]interface{}) Option {
	return func(engine *HttpServeEngine) {
		engine.defaults = defaults
	}
}

// WithBanner 配置服务Banner
func WithServerBanner(banner string) Option {
	return func(engine *HttpServeEngine) {
		engine.banner = banner
	}
}

// WithPrepareHooks 配置服务启动预备阶段Hook函数列表
func WithPrepareHooks(hooks ...flux.PrepareHookFunc) Option {
	return func(engine *HttpServeEngine) {
		engine.router.hooks = append(engine.router.hooks, hooks...)
	}
}

func NewHttpServeEngine() *HttpServeEngine {
	return NewHttpServeEngineOverride()
}

func NewHttpServeEngineOverride(overrides ...Option) *HttpServeEngine {
	opts := []Option{WithServerBanner(defaultBanner),
		WithServerResponseWriter(DefaultServerResponseWriter),
		WithServerErrorsWriter(DefaultServerErrorsWriter),
		WithServerWebInterceptors(
			webserver.NewCORSInterceptor(),
			webserver.NewRequestIdInterceptor(),
		),
		WithServerWebVersionLookupFunc(func(webc flux.WebContext) string {
			return webc.HeaderValue(DefaultHttpHeaderVersion)
		}),
		WithServerDefaults(map[string]interface{}{
			HttpWebServerConfigKeyFeatureDebugEnable: false,
			HttpWebServerConfigKeyFeatureDebugPort:   9527,
			HttpWebServerConfigKeyAddress:            "0.0.0.0",
			HttpWebServerConfigKeyPort:               8080,
		})}
	return NewHttpServeEngineWith(context2.DefaultContextFactory, append(opts, overrides...)...)
}

func NewHttpServeEngineWith(factory func() flux.Context, opts ...Option) *HttpServeEngine {
	hse := &HttpServeEngine{
		router:         NewRouter(),
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
		opt(hse)
	}
	return hse
}

// Prepare Call before init and startup
func (s *HttpServeEngine) Prepare() error {
	return s.router.Prepare()
}

// Initial
func (s *HttpServeEngine) Initial() error {
	// Http server
	s.config = flux.NewConfigurationOf(HttpWebServerConfigRootName)
	s.config.SetDefaults(s.defaults)
	// 创建WebServer
	s.httpWebServer = ext.GetWebServerFactory()(s.config)
	// 默认必备的WebServer功能
	s.httpWebServer.SetWebErrorHandler(s.defaultServerErrorHandler)
	s.httpWebServer.SetWebNotFoundHandler(s.defaultNotFoundErrorHandler)
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
		ns := flux.NamespaceEndpointDiscovery + "." + discovery.Id()
		if err := s.router.InitialHook(discovery, flux.NewConfigurationOf(ns)); nil != err {
			return err
		}
	}
	return s.router.Initial()
}

func (s *HttpServeEngine) Startup(info flux.BuildInfo) error {
	logger.Infof(VersionFormat, info.CommitId, info.Version, info.Date)
	if "" != s.banner {
		logger.Info(s.banner)
	}
	return s.start(s.config)
}

func (s *HttpServeEngine) start(config *flux.Configuration) error {
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
	logger.Infow("HttpServeEngine starting", "address", address, "cert", certFile, "key", keyFile)
	close(s.started)
	return s.httpWebServer.StartTLS(address, certFile, keyFile)
}

func (s *HttpServeEngine) startDiscovery(endpoints chan flux.HttpEndpointEvent, services chan flux.BackendServiceEvent) error {
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

func (s *HttpServeEngine) HandleEndpointRequest(webc flux.WebContext, endpoints *MultiEndpoint, tracing bool) error {
	version := s.versionLookup(webc)
	endpoint, found := endpoints.LookupByVersion(version)
	requestId := cast.ToString(webc.GetValue(flux.HeaderXRequestId))
	defer func() {
		if r := recover(); r != nil {
			trace := logger.With(requestId)
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
			logger.With(requestId).Infow("HttpServeEngine route not-found",
				"http-pattern", []string{webc.Method(), webc.RequestURI(), url.Path},
			)
		}
		return flux.ErrRouteNotFound
	}
	ctxw := s.acquireContext(requestId, webc, endpoint)
	defer s.releaseContext(ctxw)
	// Route call
	logger.WithContext(ctxw).Infow("HttpServeEngine route start")
	endcall := func(code int, start time.Time) {
		logger.WithContext(ctxw).Infow("HttpServeEngine route end",
			"metric", ctxw.LoadMetrics(),
			"elapses", time.Since(start).String(), "response.code", code)
	}
	start := time.Now()
	// Context hook
	for _, hook := range s.ctxHooks {
		hook(webc, ctxw)
	}
	// Route and response
	response := ctxw.Response()
	if err := s.router.Route(ctxw); nil != err {
		defer endcall(err.StatusCode, start)
		logger.WithContext(ctxw).Errorw("HttpServeEngine route error", "error", err)
		err.MergeHeader(response.HeaderValues())
		return err
	} else {
		defer endcall(response.StatusCode(), start)
		return s.responseWriter(webc, requestId, response.HeaderValues(), response.StatusCode(), response.Body())
	}
}

func (s *HttpServeEngine) HandleBackendServiceEvent(event flux.BackendServiceEvent) {
	service := event.Service
	initArguments(service.Arguments)
	switch event.EventType {
	case flux.EventTypeAdded:
		logger.Infow("With service",
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
	initArguments(endpoint.Service.Arguments)
	initArguments(endpoint.Permission.Arguments)
	bind, isreg := s.selectMultiEndpoint(routeKey, &endpoint)
	switch event.EventType {
	case flux.EventTypeAdded:
		logger.Infow("With endpoint", "version", endpoint.Version, "method", method, "pattern", pattern)
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
	defer close(s.stopped)
	if s.debugServer != nil {
		_ = s.debugServer.Close()
	}
	if err := s.httpWebServer.Shutdown(ctx); nil != err {
		logger.Warnw("HttpServeEngine shutdown http server", "error", err)
	}
	return s.router.Shutdown(ctx)
}

// StateStarted 返回一个Channel。当服务启动完成时，此Channel将被关闭。
func (s *HttpServeEngine) StateStarted() <-chan struct{} {
	return s.started
}

// StateStopped 返回一个Channel。当服务停止后完成时，此Channel将被关闭。
func (s *HttpServeEngine) StateStopped() <-chan struct{} {
	return s.stopped
}

// HttpConfig return Http server configuration
func (s *HttpServeEngine) HttpConfig() *flux.Configuration {
	return s.config
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

// AddServerContextHookFunc 添加Http与Flux的Context桥接函数
func (s *HttpServeEngine) AddServerContextHookFunc(f flux.ServerContextHookFunc) {
	s.ctxHooks = append(s.ctxHooks, f)
}

func (s *HttpServeEngine) newWrappedEndpointHandler(endpoint *MultiEndpoint) flux.WebHandler {
	enabled := s.config.GetBool(HttpWebServerConfigKeyRequestLogEnable)
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

func (s *HttpServeEngine) acquireContext(id string, webc flux.WebContext, endpoint *flux.Endpoint) *context2.DefaultContext {
	ctx := s.ctxPool.Get().(*context2.DefaultContext)
	ctx.Reattach(id, webc, endpoint)
	return ctx
}

func (s *HttpServeEngine) releaseContext(context *context2.DefaultContext) {
	context.Release()
	s.ctxPool.Put(context)
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
	if err := s.errorsWriter(webc, requestId, serve.Header, serve); nil != err {
		logger.With(requestId).Errorw("HttpServeEngine http response error", "error", err)
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
