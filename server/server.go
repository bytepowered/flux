package server

import (
	"context"
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/internal"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/webx"
	"github.com/bytepowered/flux/webx/echo"
	"github.com/bytepowered/flux/webx/middleware"
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
	HttpServerConfigKeyFeatureCorsEnable  = "feature-cors-enable"
	HttpServerConfigKeyVersionHeader      = "version-header"
	HttpServerConfigKeyRequestIdHeaders   = "request-id-headers"
	HttpServerConfigKeyRequestLogEnable   = "request-log-enable"
	HttpServerConfigKeyAddress            = "address"
	HttpServerConfigKeyPort               = "port"
	HttpServerConfigKeyTlsCertFile        = "tls-cert-file"
	HttpServerConfigKeyTlsKeyFile         = "tls-key-file"
)

const (
	DebugPathVars      = "/debug/vars"
	DebugPathPprof     = "/debug/pprof/*"
	DebugPathEndpoints = "/debug/endpoints"
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
		HttpServerConfigKeyAddress:            "0.0.0.0",
		HttpServerConfigKeyPort:               8080,
	}
)

// HttpContextPipelineFunc
type HttpContextPipelineFunc func(webx.WebContext, flux.Context)

// Server
type HttpServer struct {
	webServer         webx.WebServer
	webWriter         webx.WebServerWriter
	httpConfig        *flux.Configuration
	httpVersionHeader string
	routerEngine      *internal.RouteEngine
	routerRegistry    flux.Registry
	mvEndpointMap     map[string]*internal.MultiVersionEndpoint
	contextWrappers   sync.Pool
	pipelines         []HttpContextPipelineFunc
	stateStarted      chan struct{}
	stateStopped      chan struct{}
}

func NewHttpServer() *HttpServer {
	return &HttpServer{
		webWriter:       new(HttpServerWriter),
		routerEngine:    internal.NewRouteEngine(),
		mvEndpointMap:   make(map[string]*internal.MultiVersionEndpoint),
		contextWrappers: sync.Pool{New: internal.NewContextWrapper},
		pipelines:       make([]HttpContextPipelineFunc, 0),
		stateStarted:    make(chan struct{}),
		stateStopped:    make(chan struct{}),
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

func (s *HttpServer) Initial() error {
	return s.InitServer()
}

// InitServer : Call before startup
func (s *HttpServer) InitServer() error {
	// Http server
	s.httpConfig = flux.NewConfigurationOf(HttpServerConfigRootName)
	s.httpConfig.SetDefaults(HttpServerConfigDefaults)
	s.httpVersionHeader = s.httpConfig.GetString(HttpServerConfigKeyVersionHeader)
	// 创建Echo框架的WebServer
	s.webServer = echo.NewAdaptWebServer()

	// 默认必备的WebServer功能
	s.webServer.SetWebErrorHandler(s.handleServerError)
	s.webServer.SetRouteNotFoundHandler(s.handleNotFoundError)

	// - 请求CORS跨域支持：默认关闭，需要配置开启
	if s.httpConfig.GetBool(HttpServerConfigKeyFeatureCorsEnable) {
		s.AddHttpInterceptor(middleware.NewCORSMiddleware())
	}

	// - RequestId查找与生成
	if headers := s.httpConfig.GetStringSlice(HttpServerConfigKeyRequestIdHeaders); 0 < len(headers) {
		for _, header := range headers {
			middleware.AddRequestIdLookupHeader(header)
		}
	}
	s.AddHttpInterceptor(middleware.NewRequestIdMiddleware())

	// - Debug特性支持：默认关闭，需要配置开启
	if s.httpConfig.GetBool(HttpServerConfigKeyFeatureDebugEnable) {
		s.debugFeatures(s.httpConfig)
	}

	// Registry
	if registry, config, err := findRouterRegistry(); nil != err {
		return err
	} else {
		if err := s.routerEngine.InitialHook(registry, config); nil != err {
			return err
		}
		s.routerRegistry = registry
	}
	return s.routerEngine.Initial()
}

func (s *HttpServer) Startup(version flux.BuildInfo) error {
	return s.StartServe(version)
}

// StartServe server
func (s *HttpServer) StartServe(version flux.BuildInfo) error {
	return s.StartServeWith(version, s.httpConfig)
}

func (s *HttpServer) StartupWith(version flux.BuildInfo, httpConfig *flux.Configuration) error {
	return s.StartServeWith(version, httpConfig)
}

// StartServeWith server
func (s *HttpServer) StartServeWith(info flux.BuildInfo, config *flux.Configuration) error {
	if err := s.ensure().routerEngine.Startup(); nil != err {
		return err
	}
	events := make(chan flux.EndpointEvent, 2)
	defer close(events)
	if err := s.watchRouterRegistry(events); nil != err {
		return fmt.Errorf("start registry watching: %w", err)
	} else {
		go s.handleRouteRegistryEvent(events)
	}
	address := fmt.Sprintf("%s:%d", config.GetString("address"), config.GetInt("port"))
	certFile := config.GetString(HttpServerConfigKeyTlsCertFile)
	keyFile := config.GetString(HttpServerConfigKeyTlsKeyFile)
	close(s.stateStarted)
	logger.Info(Banner)
	logger.Infof(VersionFormat, info.CommitId, info.Version, info.Date)
	if certFile != "" && keyFile != "" {
		logger.Infof("HttpServer(HTTP/2 TLS) starting: %s", address)
		return s.webServer.StartTLS(address, certFile, keyFile)
	} else {
		logger.Infof("HttpServer starting: %s", address)
		return s.webServer.Start(address)
	}
}

// Shutdown to cleanup resources
func (s *HttpServer) Shutdown(ctx context.Context) error {
	logger.Info("HttpServer shutdown...")
	defer close(s.stateStopped)
	// Stop http server
	if err := s.webServer.Shutdown(ctx); nil != err {
		return err
	}
	// Stop routerEngine
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

// AddHttpInterceptor 添加Http前拦截器。将在Http被路由到对应Handler之前执行
func (s *HttpServer) AddHttpInterceptor(m webx.WebMiddleware) {
	s.ensure().webServer.AddWebInterceptor(m)
}

// AddHttpMiddleware 添加Http中间件。在Http路由到对应Handler后执行
func (s *HttpServer) AddHttpMiddleware(m webx.WebMiddleware) {
	s.ensure().webServer.AddWebMiddleware(m)
}

// AddHttpHandler 添加Http处理接口。
func (s *HttpServer) AddHttpHandler(method, pattern string, h webx.WebRouteHandler, m ...webx.WebMiddleware) {
	s.ensure().webServer.AddWebRouteHandler(method, pattern, h, m...)
}

// SetHttpNotFoundHandler 设置Http路由失败的处理接口
func (s *HttpServer) SetHttpNotFoundHandler(nfh webx.WebRouteHandler) {
	s.ensure().webServer.SetRouteNotFoundHandler(nfh)
}

// SetHttpNotFoundHandler 设置Http响应数据写入的处理接口
func (s *HttpServer) SetHttpServerWriter(writer webx.WebServerWriter) {
	s.webWriter = writer
}

// AddHttpContextPipeline 添加Http与Flux的Context桥接函数
func (s *HttpServer) AddHttpContextPipeline(f HttpContextPipelineFunc) {
	s.pipelines = append(s.pipelines, f)
}

func (s *HttpServer) watchRouterRegistry(events chan<- flux.EndpointEvent) error {
	return s.routerRegistry.WatchEvents(events)
}

func (s *HttpServer) handleRouteRegistryEvent(events <-chan flux.EndpointEvent) {
	for event := range events {
		pattern := toHttpPattern(event.HttpPattern)
		routeKey := fmt.Sprintf("%s#%s", event.HttpMethod, pattern)
		multi, isregister := s.prepareMultiVersionEndpoint(routeKey)
		// Check http method
		event.Endpoint.HttpMethod = strings.ToUpper(event.Endpoint.HttpMethod)
		if !isAllowMethod(event.Endpoint.HttpMethod) {
			continue
		}
		// Refresh endpoint
		endpoint := event.Endpoint
		switch event.EventType {
		case flux.EndpointEventAdded:
			logger.Infow("New endpoint", "version", endpoint.Version, "method", event.HttpMethod, "pattern", pattern)
			multi.Update(endpoint.Version, &endpoint)
			if isregister {
				logger.Infow("Register http router", "method", event.HttpMethod, "pattern", pattern)
				s.webServer.AddWebRouteHandler(event.HttpMethod, pattern, s.newHttpRouteHandler(multi))
			}
		case flux.EndpointEventUpdated:
			logger.Infow("Update endpoint", "version", endpoint.Version, "method", event.HttpMethod, "pattern", pattern)
			multi.Update(endpoint.Version, &endpoint)
		case flux.EndpointEventRemoved:
			logger.Infow("Delete endpoint", "method", event.HttpMethod, "pattern", pattern)
			multi.Delete(endpoint.Version)
		}
	}
}

func (s *HttpServer) acquire(id string, webc webx.WebContext, endpoint *flux.Endpoint) *internal.ContextWrapper {
	ctx := s.contextWrappers.Get().(*internal.ContextWrapper)
	ctx.Reattach(id, webc, endpoint)
	return ctx
}

func (s *HttpServer) release(context *internal.ContextWrapper) {
	context.Release()
	s.contextWrappers.Put(context)
}

func (s *HttpServer) newHttpRouteHandler(mvEndpoint *internal.MultiVersionEndpoint) webx.WebRouteHandler {
	requestLogEnable := s.httpConfig.GetBool(HttpServerConfigKeyRequestLogEnable)
	return func(webc webx.WebContext) error {
		// Multi version selection
		version := webc.RequestHeader().Get(s.httpVersionHeader)
		endpoint, found := mvEndpoint.FindByVersion(version)
		requestId := cast.ToString(webc.GetValue(webx.HeaderXRequestId))
		defer func() {
			if err := recover(); err != nil {
				tl := logger.Trace(requestId)
				tl.Errorw("Server dispatch: unexpected error", "error", err)
				tl.Error(string(debug.Stack()))
			}
		}()
		if !found {
			if requestLogEnable {
				logger.Trace(requestId).Infow("HttpServer routing: ENDPOINT_NOT_FOUND",
					"method", webc.Method(), "uri", webc.RequestURI(), "path", webc.RequestURLPath(), "version", version,
				)
			}
			return s.webWriter.WriteError(webc, requestId, http.Header{}, ErrEndpointVersionNotFound)
		}
		ctxw := s.acquire(requestId, webc, endpoint)
		defer s.release(ctxw)
		// Context exchange
		for _, pf := range s.pipelines {
			pf(webc, ctxw)
		}
		if requestLogEnable {
			logger.Trace(ctxw.RequestId()).Infow("HttpServer routing: DISPATCHING",
				"method", webc.Method(), "uri", webc.RequestURI(), "path", webc.RequestURLPath(), "version", version,
				"endpoint", endpoint.UpstreamMethod+":"+endpoint.UpstreamUri,
			)
		}
		// Route and response
		if err := s.routerEngine.Route(ctxw); nil != err {
			return s.webWriter.WriteError(webc, requestId, ctxw.Response().Headers(), err)
		} else {
			rw := ctxw.Response()
			return s.webWriter.WriteBody(webc, requestId, rw.Headers(), rw.StatusCode(), rw.Body())
		}
	}
}

func (s *HttpServer) handleNotFoundError(webc webx.WebContext) error {
	return &flux.StateError{
		StatusCode: flux.StatusNotFound,
		ErrorCode:  flux.ErrorCodeRequestNotFound,
		Message:    "ROUTE:NOT_FOUND",
	}
}

// handleServerError EchoHttp状态错误处理函数。
func (s *HttpServer) handleServerError(err error, webc webx.WebContext) {
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
	requestId := cast.ToString(webc.GetValue(webx.HeaderXRequestId))
	if err := s.webWriter.WriteError(webc, requestId, http.Header{}, serr); nil != err {
		logger.Errorw("Server http response error", "error", err)
	}
}

func (s *HttpServer) prepareMultiVersionEndpoint(routeKey string) (*internal.MultiVersionEndpoint, bool) {
	if mve, ok := s.mvEndpointMap[routeKey]; ok {
		return mve, false
	} else {
		mve = internal.NewMultiVersionEndpoint()
		s.mvEndpointMap[routeKey] = mve
		return mve, true
	}
}

func (s *HttpServer) ensure() *HttpServer {
	if s.webServer == nil {
		logger.Panicf("Call must after InitServer()")
	}
	return s
}

func findRouterRegistry() (flux.Registry, *flux.Configuration, error) {
	config := flux.NewConfigurationOf(flux.KeyConfigRootRegistry)
	config.SetDefault(flux.KeyConfigRegistryId, ext.RegistryIdDefault)
	registryId := config.GetString(flux.KeyConfigRegistryId)
	logger.Infow("Active router registry", "registry-id", registryId)
	if factory, ok := ext.GetRegistryFactory(registryId); !ok {
		return nil, config, fmt.Errorf("RegistryFactory not found, id: %s", registryId)
	} else {
		return factory(), config, nil
	}
}

func toHttpPattern(uri string) string {
	// /api/{userId} -> /api/:userId
	replaced := strings.Replace(uri, "}", "", -1)
	if len(replaced) < len(uri) {
		return strings.Replace(replaced, "{", ":", -1)
	} else {
		return uri
	}
}

func isAllowMethod(method string) bool {
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
