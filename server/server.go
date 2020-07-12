package server

import (
	context "context"
	"expvar"
	"fmt"
	"github.com/bwmarrin/snowflake"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/internal"
	"github.com/bytepowered/flux/logger"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	https "net/http"
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
	DefaultHttpVersionHeader   = "X-Version"
	DefaultHttpRequestIdHeader = "X-Request-Id"
)

const (
	ConfigHttpRootName      = "HttpServer"
	ConfigHttpVersionHeader = "version-header"
	ConfigHttpDebugEnable   = "debug-enable"
	ConfigHttpAddress       = "address"
	ConfigHttpPort          = "port"
	ConfigHttpTlsCertFile   = "tls-cert-file"
	ConfigHttpTlsKeyFile    = "tls-key-file"
)

const (
	DebugPathVars      = "/debug/vars"
	DebugPathPprof     = "/debug/pprof/*"
	DebugPathEndpoints = "/debug/endpoints"
)

var (
	ErrEndpointVersionNotFound = &flux.InvokeError{
		StatusCode: flux.StatusNotFound,
		ErrorCode:  flux.ErrorCodeGatewayEndpoint,
		Message:    "ENDPOINT_VERSION_NOT_FOUND",
	}
)

var (
	HttpServerConfigDefaults = map[string]interface{}{
		ConfigHttpVersionHeader: DefaultHttpVersionHeader,
		ConfigHttpDebugEnable:   false,
		ConfigHttpAddress:       "0.0.0.0",
		ConfigHttpPort:          8080,
	}
)

// HttpContextPipelineFunc
type HttpContextPipelineFunc func(echo.Context, flux.Context)

// Server
type HttpServer struct {
	server            *echo.Echo
	httpConfig        *flux.Configuration
	httpVisits        *expvar.Int
	httpWriter        flux.HttpResponseWriter
	httpNotFound      echo.HandlerFunc
	httpVersionHeader string
	dispatcher        *internal.ServerDispatcher
	mvEndpointMap     map[string]*internal.MultiVersionEndpoint
	contextWrappers   sync.Pool
	snowflakeId       *snowflake.Node
	pipelines         []HttpContextPipelineFunc
}

func NewHttpServer() *HttpServer {
	id, _ := snowflake.NewNode(1)
	return &HttpServer{
		httpVisits:      expvar.NewInt("visits"),
		httpWriter:      new(HttpServerResponseWriter),
		dispatcher:      internal.NewDispatcher(),
		mvEndpointMap:   make(map[string]*internal.MultiVersionEndpoint),
		contextWrappers: sync.Pool{New: internal.NewContextWrapper},
		pipelines:       make([]HttpContextPipelineFunc, 0),
		snowflakeId:     id,
	}
}

// HttpConfig return Http server configuration
func (s *HttpServer) HttpConfig() *flux.Configuration {
	return s.httpConfig
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
	s.httpConfig = flux.NewConfigurationOf(ConfigHttpRootName)
	s.httpConfig.SetDefaults(HttpServerConfigDefaults)
	s.httpVersionHeader = s.httpConfig.GetString(ConfigHttpVersionHeader)
	s.server = echo.New()
	s.server.HideBanner = true
	s.server.HidePort = true
	s.server.HTTPErrorHandler = s.handleServerError
	// Http拦截器
	if !s.httpConfig.GetBool("cors-disable") {
		s.AddHttpInterceptor(middleware.CORS())
	}
	s.AddHttpInterceptor(RepeatableBody)
	// Http debug features
	if s.httpConfig.GetBool(ConfigHttpDebugEnable) {
		s.debugFeatures(s.httpConfig)
	}
	return s.dispatcher.Initial()
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
	if err := s.ensure().dispatcher.Startup(); nil != err {
		return err
	}
	eventCh := make(chan flux.EndpointEvent, 2)
	defer close(eventCh)
	if err := s.dispatcher.WatchRegistry(eventCh); nil != err {
		return fmt.Errorf("start registry watching: %w", err)
	} else {
		go s.handleEndpointRegisterEvent(eventCh)
	}
	address := fmt.Sprintf("%s:%d", config.GetString("address"), config.GetInt("port"))
	certFile := config.GetString(ConfigHttpTlsCertFile)
	keyFile := config.GetString(ConfigHttpTlsKeyFile)
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
	// Stop http server
	if err := s.server.Shutdown(ctx); nil != err {
		return err
	}
	// Stop dispatcher
	return s.dispatcher.Shutdown(ctx)
}

func (s *HttpServer) handleEndpointRegisterEvent(events <-chan flux.EndpointEvent) {
	for event := range events {
		pattern := s.toHttpServerPattern(event.HttpPattern)
		routeKey := fmt.Sprintf("%s#%s", event.HttpMethod, pattern)
		vEndpoint, isNew := s.getVersionEndpoint(routeKey)
		// Check http method
		event.Endpoint.HttpMethod = strings.ToUpper(event.Endpoint.HttpMethod)
		eEndpoint := event.Endpoint
		switch eEndpoint.HttpMethod {
		case https.MethodGet, https.MethodPost, https.MethodDelete, https.MethodPut,
			https.MethodHead, https.MethodOptions, https.MethodPatch, https.MethodTrace:
			// Allowed
		default:
			// http.MethodConnect, and Others
			logger.Errorw("Ignore unsupported http method:", "method", eEndpoint.HttpMethod)
			continue
		}
		switch event.EventType {
		case flux.EndpointEventAdded:
			logger.Infow("New endpoint", "version", eEndpoint.Version, "method", event.HttpMethod, "pattern", pattern)
			vEndpoint.Update(eEndpoint.Version, &eEndpoint)
			if isNew {
				logger.Infow("New http routing", "method", event.HttpMethod, "pattern", pattern)
				s.server.Add(event.HttpMethod, pattern, s.newRequestRouter(vEndpoint))
			}
		case flux.EndpointEventUpdated:
			logger.Infow("Update endpoint", "version", eEndpoint.Version, "method", event.HttpMethod, "pattern", pattern)
			vEndpoint.Update(eEndpoint.Version, &eEndpoint)
		case flux.EndpointEventRemoved:
			logger.Infow("Delete endpoint", "method", event.HttpMethod, "pattern", pattern)
			vEndpoint.Delete(eEndpoint.Version)
		}
	}
}

func (s *HttpServer) acquire(echo echo.Context, endpoint *flux.Endpoint) *internal.ContextWrapper {
	requestId := echo.Request().Header.Get(DefaultHttpRequestIdHeader)
	if "" == requestId {
		requestId = s.snowflakeId.Generate().Base64()
	}
	ctx := s.contextWrappers.Get().(*internal.ContextWrapper)
	ctx.Reattach(requestId, echo, endpoint)
	return ctx
}

func (s *HttpServer) release(context *internal.ContextWrapper) {
	context.Release()
	s.contextWrappers.Put(context)
}

func (s *HttpServer) newRequestRouter(mvEndpoint *internal.MultiVersionEndpoint) echo.HandlerFunc {
	return func(echo echo.Context) error {
		s.httpVisits.Add(1)
		request := echo.Request()
		// Multi version selection
		version := request.Header.Get(s.httpVersionHeader)
		endpoint, found := mvEndpoint.Get(version)
		ctx := s.acquire(echo, endpoint)
		defer func(requestId string) {
			if err := recover(); err != nil {
				tl := logger.Trace(requestId)
				tl.Errorw("Server dispatch: unexpected error", "error", err)
				tl.Error(string(debug.Stack()))
			}
		}(ctx.RequestId())
		echo.Response().Header().Set(flux.XRequestId, ctx.RequestId())
		defer s.release(ctx)
		trace := logger.Trace(ctx.RequestId())
		if !found {
			trace.Infow("Server dispatch: <ENDPOINT_NOT_FOUND>",
				"method", request.Method, "uri", request.RequestURI, "path", request.URL.Path, "version", version,
			)
			return s.httpWriter.WriteError(echo, ctx.RequestId(), ctx.ResponseWriter().Headers(), ErrEndpointVersionNotFound)
		}
		// Context exchange: Echo <-> Flux
		for _, pf := range s.pipelines {
			pf(echo, ctx)
		}
		trace.Infow("Server dispatch: routing",
			"method", request.Method, "uri", request.RequestURI, "path", request.URL.Path, "version", version,
			"endpoint", endpoint.UpstreamMethod+":"+endpoint.UpstreamUri,
		)
		rw := ctx.ResponseWriter()
		if err := s.dispatcher.Dispatch(ctx); nil != err {
			return s.httpWriter.WriteError(echo, ctx.RequestId(), rw.Headers(), err)
		} else {
			return s.httpWriter.WriteBody(echo, ctx.RequestId(), rw.Headers(), rw.StatusCode(), rw.Body())
		}
	}
}

// handleServerError EchoHttp状态错误处理函数。
func (s *HttpServer) handleServerError(err error, ctx echo.Context) {
	// Http中间件等返回InvokeError错误
	inverr, ok := err.(*flux.InvokeError)
	if !ok {
		inverr = &flux.InvokeError{
			StatusCode: flux.StatusServerError,
			ErrorCode:  flux.ErrorCodeGatewayInternal,
			Message:    err.Error(),
			Internal:   err,
		}
	}
	id := ctx.Response().Header().Get(flux.XRequestId)
	if err := s.httpWriter.WriteError(ctx, id, https.Header{}, inverr); nil != err {
		logger.Errorw("Server http response error", "error", err)
	}
}

func (s *HttpServer) toHttpServerPattern(uri string) string {
	// /api/{userId} -> /api/:userId
	replaced := strings.Replace(uri, "}", "", -1)
	if len(replaced) < len(uri) {
		return strings.Replace(replaced, "{", ":", -1)
	} else {
		return uri
	}
}

func (s *HttpServer) getVersionEndpoint(routeKey string) (*internal.MultiVersionEndpoint, bool) {
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
