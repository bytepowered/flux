package server

import (
	"context"
	"fmt"
	"github.com/bwmarrin/snowflake"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/internal"
	"github.com/bytepowered/flux/logger"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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
	DefaultHttpHeaderVersion   = "X-Version"
	DefaultHttpHeaderRequestId = "X-Request-Id"
)

const (
	HttpServerConfigRootName            = "HttpServer"
	HttpServerConfigKeyVersionHeader    = "version-header"
	HttpServerConfigKeyDebugEnable      = "debug-enable"
	HttpServerConfigKeyRequestLogEnable = "request-log-enable"
	HttpServerConfigKeyCorsDisable      = "cors-disable"
	HttpServerConfigKeyAddress          = "address"
	HttpServerConfigKeyPort             = "port"
	HttpServerConfigKeyTlsCertFile      = "tls-cert-file"
	HttpServerConfigKeyTlsKeyFile       = "tls-key-file"
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
		HttpServerConfigKeyVersionHeader: DefaultHttpHeaderVersion,
		HttpServerConfigKeyDebugEnable:   false,
		HttpServerConfigKeyAddress:       "0.0.0.0",
		HttpServerConfigKeyPort:          8080,
	}
)

// HttpContextPipelineFunc
type HttpContextPipelineFunc func(echo.Context, flux.Context)

// Server
type HttpServer struct {
	server            *echo.Echo
	httpConfig        *flux.Configuration
	httpWriter        flux.HttpResponseWriter
	httpNotFound      echo.HandlerFunc
	httpVersionHeader string
	routerEngine      *internal.RouteEngine
	routerRegistry    flux.Registry
	mvEndpointMap     map[string]*internal.MultiVersionEndpoint
	contextWrappers   sync.Pool
	snowflakeId       *snowflake.Node
	pipelines         []HttpContextPipelineFunc
	stateStarted      chan struct{}
	stateStopped      chan struct{}
}

func NewHttpServer() *HttpServer {
	id, _ := snowflake.NewNode(1)
	return &HttpServer{
		httpWriter:      new(HttpServerResponseWriter),
		routerEngine:    internal.NewRouteEngine(),
		mvEndpointMap:   make(map[string]*internal.MultiVersionEndpoint),
		contextWrappers: sync.Pool{New: internal.NewContextWrapper},
		pipelines:       make([]HttpContextPipelineFunc, 0),
		snowflakeId:     id,
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
	s.server = echo.New()
	s.server.HideBanner = true
	s.server.HidePort = true
	s.server.HTTPErrorHandler = s.handleServerError
	// Http拦截器
	if !s.httpConfig.GetBool(HttpServerConfigKeyCorsDisable) {
		s.AddHttpInterceptor(middleware.CORS())
	}
	s.AddHttpInterceptor(RepeatableHttpBody)
	// Http debug features
	if s.httpConfig.GetBool(HttpServerConfigKeyDebugEnable) {
		s.debugFeatures(s.httpConfig)
	}
	// Http Registry
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
	logger.Info(Banner)
	logger.Infof(VersionFormat, info.CommitId, info.Version, info.Date)
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
	if certFile != "" && keyFile != "" {
		logger.Infof("HttpServer(HTTP/2 TLS) starting: %s", address)
		return s.server.StartTLS(address, certFile, keyFile)
	} else {
		logger.Infof("HttpServer starting: %s", address)
		return s.server.Start(address)
	}
}

// Shutdown to cleanup resources
func (s *HttpServer) Shutdown(ctx context.Context) error {
	logger.Info("HttpServer shutdown...")
	defer close(s.stateStopped)
	// Stop http server
	if err := s.server.Shutdown(ctx); nil != err {
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

// HttpServer 返回Http服务器实例
func (s *HttpServer) HttpServer() *echo.Echo {
	return s.ensure().server
}

// AddHttpInterceptor 添加Http前拦截器。将在Http被路由到对应Handler之前执行
func (s *HttpServer) AddHttpInterceptor(m echo.MiddlewareFunc) {
	s.ensure().server.Pre(m)
}

// AddHttpMiddleware 添加Http中间件。在Http路由到对应Handler后执行
func (s *HttpServer) AddHttpMiddleware(m echo.MiddlewareFunc) {
	s.ensure().server.Use(m)
}

// AddHttpHandler 添加Http处理接口。
func (s *HttpServer) AddHttpHandler(method, pattern string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) {
	s.ensure().server.Add(method, pattern, h, m...)
}

// SetHttpNotFoundHandler 设置Http路由失败的处理接口
func (s *HttpServer) SetHttpNotFoundHandler(nfh echo.HandlerFunc) {
	echo.NotFoundHandler = nfh
}

// SetHttpNotFoundHandler 设置Http响应数据写入的处理接口
func (s *HttpServer) SetHttpResponseWriter(writer flux.HttpResponseWriter) {
	s.httpWriter = writer
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
				s.server.Add(event.HttpMethod, pattern, s.newHttpRouteHandler(multi))
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

func (s *HttpServer) lookupId(echo echo.Context) string {
	id := echo.Request().Header.Get(DefaultHttpHeaderRequestId)
	if "" == id {
		id = s.snowflakeId.Generate().Base64()
	}
	echo.Request().Header.Set(DefaultHttpHeaderRequestId, id)
	echo.Response().Header().Set(DefaultHttpHeaderRequestId, id)
	return id
}

func (s *HttpServer) acquire(id string, echo echo.Context, endpoint *flux.Endpoint) *internal.ContextWrapper {
	ctx := s.contextWrappers.Get().(*internal.ContextWrapper)
	ctx.Reattach(id, echo, endpoint)
	return ctx
}

func (s *HttpServer) release(context *internal.ContextWrapper) {
	context.Release()
	s.contextWrappers.Put(context)
}

func (s *HttpServer) newHttpRouteHandler(mvEndpoint *internal.MultiVersionEndpoint) echo.HandlerFunc {
	requestLogEnable := s.httpConfig.GetBool(HttpServerConfigKeyRequestLogEnable)
	return func(httpctx echo.Context) error {
		request := httpctx.Request()
		// Multi version selection
		version := request.Header.Get(s.httpVersionHeader)
		endpoint, found := mvEndpoint.FindByVersion(version)
		requestId := s.lookupId(httpctx)
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
					"method", request.Method, "uri", request.RequestURI, "path", request.URL.Path, "version", version,
				)
			}
			return s.httpWriter.WriteError(httpctx, requestId, http.Header{}, ErrEndpointVersionNotFound)
		}
		ctxw := s.acquire(requestId, httpctx, endpoint)
		defer s.release(ctxw)
		// Context exchange
		for _, pf := range s.pipelines {
			pf(httpctx, ctxw)
		}
		if requestLogEnable {
			logger.Trace(ctxw.RequestId()).Infow("HttpServer routing: DISPATCHING",
				"method", request.Method, "uri", request.RequestURI, "path", request.URL.Path, "version", version,
				"endpoint", endpoint.UpstreamMethod+":"+endpoint.UpstreamUri,
			)
		}
		// Route and response
		if err := s.routerEngine.Route(ctxw); nil != err {
			return s.httpWriter.WriteError(httpctx, requestId, ctxw.ResponseWriter().Headers(), err)
		} else {
			rw := ctxw.ResponseWriter()
			return s.httpWriter.WriteBody(httpctx, requestId, rw.Headers(), rw.StatusCode(), rw.Body())
		}
	}
}

// handleServerError EchoHttp状态错误处理函数。
func (s *HttpServer) handleServerError(err error, ctx echo.Context) {
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
	id := ctx.Response().Header().Get(flux.XRequestId)
	if err := s.httpWriter.WriteError(ctx, id, http.Header{}, serr); nil != err {
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
	if s.server == nil {
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
