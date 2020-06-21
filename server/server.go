package server

import (
	"bytes"
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
	"io"
	"io/ioutil"
	https "net/http"
	_ "net/http/pprof"
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

const (
	_echoAttrRoutedContext = "$inner.flux.context"
)

var (
	errEndpointVersionNotFound = &flux.InvokeError{
		StatusCode: flux.StatusNotFound,
		Message:    "ENDPOINT_VERSION_NOT_FOUND",
	}
	httpDefaults = map[string]interface{}{
		ConfigHttpVersionHeader: DefaultHttpVersionHeader,
		ConfigHttpDebugEnable:   false,
		ConfigHttpAddress:       "0.0.0.0",
		ConfigHttpPort:          8080,
	}
)

// Context pipeline
type ContextPipelineFunc func(echo.Context, flux.Context)

// FluxServer
type FluxServer struct {
	httpServer        *echo.Echo
	httpConfig        *flux.Configuration
	httpVisits        *expvar.Int
	httpAdaptWriter   internal.HttpWriter
	httpNotFound      echo.HandlerFunc
	httpVersionHeader string
	dispatcher        *internal.FxDispatcher
	endpointMvMap     map[string]*internal.MultiVersionEndpoint
	contextWrappers   sync.Pool
	snowflakeId       *snowflake.Node
	pipelines         []ContextPipelineFunc
}

func NewFluxServer() *FluxServer {
	id, _ := snowflake.NewNode(1)
	return &FluxServer{
		httpVisits:      expvar.NewInt("visits"),
		dispatcher:      internal.NewDispatcher(),
		endpointMvMap:   make(map[string]*internal.MultiVersionEndpoint),
		contextWrappers: sync.Pool{New: internal.NewFxContext},
		pipelines:       make([]ContextPipelineFunc, 0),
		snowflakeId:     id,
	}
}

// HttpConfig return Http server configuration
func (fs *FluxServer) HttpConfig() *flux.Configuration {
	return fs.httpConfig
}

// Prepare Call before init and startup
func (fs *FluxServer) Prepare(hooks ...flux.PrepareHook) error {
	for _, prepare := range append(ext.PrepareHooks(), hooks...) {
		if err := prepare(); nil != err {
			return err
		}
	}
	return nil
}

func (fs *FluxServer) Initial() error {
	return fs.InitServer()
}

// InitServer : Call before startup
func (fs *FluxServer) InitServer() error {
	// Http server
	fs.httpConfig = flux.NewConfigurationOf(ConfigHttpRootName)
	fs.httpConfig.SetDefaults(httpDefaults)
	fs.httpVersionHeader = fs.httpConfig.GetString(ConfigHttpVersionHeader)
	fs.httpServer = echo.New()
	fs.httpServer.HideBanner = true
	fs.httpServer.HidePort = true
	fs.httpServer.HTTPErrorHandler = fs.httpErrorAdapting
	// Http拦截器
	if !fs.httpConfig.GetBool("cors-disable") {
		fs.AddHttpInterceptor(middleware.CORS())
	}
	fs.AddHttpInterceptor(fs.requestPrepare())
	// Http debug features
	if fs.httpConfig.GetBool(ConfigHttpDebugEnable) {
		fs.debugFeatures(fs.httpConfig)
	}
	return fs.dispatcher.Initial()
}

func (fs *FluxServer) Startup(version flux.BuildInfo) error {
	return fs.StartServe(version)
}

// StartServe server
func (fs *FluxServer) StartServe(version flux.BuildInfo) error {
	return fs.StartServeWith(version, fs.httpConfig)
}

func (fs *FluxServer) StartupWith(version flux.BuildInfo, httpConfig *flux.Configuration) error {
	return fs.StartServeWith(version, httpConfig)
}

// StartServeWith server
func (fs *FluxServer) StartServeWith(info flux.BuildInfo, config *flux.Configuration) error {
	logger.Info(Banner)
	logger.Infof(VersionFormat, info.CommitId, info.Version, info.Date)
	if err := fs.ensure().dispatcher.Startup(); nil != err {
		return err
	}
	eventCh := make(chan flux.EndpointEvent, 2)
	defer close(eventCh)
	if err := fs.dispatcher.WatchRegistry(eventCh); nil != err {
		return fmt.Errorf("start registry watching: %w", err)
	} else {
		go fs.handleEndpointRegisterEvent(eventCh)
	}
	address := fmt.Sprintf("%s:%d", config.GetString("address"), config.GetInt("port"))
	certFile := config.GetString(ConfigHttpTlsCertFile)
	keyFile := config.GetString(ConfigHttpTlsKeyFile)
	if certFile != "" && keyFile != "" {
		logger.Infof("HttpServer(HTTP/2 TLS) starting: %s", address)
		return fs.httpServer.StartTLS(address, certFile, keyFile)
	} else {
		logger.Infof("HttpServer starting: %s", address)
		return fs.httpServer.Start(address)
	}
}

// Shutdown to cleanup resources
func (fs *FluxServer) Shutdown(ctx context.Context) error {
	logger.Info("HttpServer shutdown...")
	// Stop http server
	if err := fs.httpServer.Shutdown(ctx); nil != err {
		return err
	}
	// Stop dispatcher
	return fs.dispatcher.Shutdown(ctx)
}

func (fs *FluxServer) handleEndpointRegisterEvent(events <-chan flux.EndpointEvent) {
	for event := range events {
		pattern := fs.toHttpServerPattern(event.HttpPattern)
		routeKey := fmt.Sprintf("%s#%s", event.HttpMethod, pattern)
		vEndpoint, isNew := fs.getVersionEndpoint(routeKey)
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
				fs.httpServer.Add(event.HttpMethod, pattern, fs.newHttpRouter(vEndpoint))
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

func (fs *FluxServer) acquire(echo echo.Context, endpoint *flux.Endpoint) *internal.ContextWrapper {
	requestId := echo.Request().Header.Get(DefaultHttpRequestIdHeader)
	if "" == requestId {
		requestId = fs.snowflakeId.Generate().Base64()
	}
	ctx := fs.contextWrappers.Get().(*internal.ContextWrapper)
	ctx.Reattach(requestId, echo, endpoint)
	return ctx
}

func (fs *FluxServer) release(context *internal.ContextWrapper) {
	context.Release()
	fs.contextWrappers.Put(context)
}

func (fs *FluxServer) newHttpRouter(mvEndpoint *internal.MultiVersionEndpoint) echo.HandlerFunc {
	return func(echo echo.Context) error {
		fs.httpVisits.Add(1)
		httpRequest := echo.Request()
		// Multi version selection
		version := httpRequest.Header.Get(fs.httpVersionHeader)
		endpoint, found := mvEndpoint.Get(version)
		ctx := fs.acquire(echo, endpoint)
		defer func(requestId string) {
			if err := recover(); err != nil {
				logger.Trace(requestId).Errorw("Http request unexpected error", "error", err)
			}
		}(ctx.RequestId())
		// Context exchange: Echo <-> Flux
		for _, pipe := range fs.pipelines {
			pipe(echo, ctx)
		}
		echo.Set(_echoAttrRoutedContext, ctx)
		defer fs.release(ctx)
		logger.Trace(ctx.RequestId()).Info("Http routing request",
			"id", ctx.RequestId(),
			"method", httpRequest.Method,
			"uri", httpRequest.RequestURI,
			"ver", version)
		if !found {
			return errEndpointVersionNotFound
		}
		if err := fs.dispatcher.Dispatch(ctx); nil != err {
			return err
		} else {
			return fs.httpAdaptWriter.WriteResponse(ctx)
		}
	}
}

// httpErrorAdapting EchoHttp状态错误处理函数。Err包含网关内部错误，以及Http框架原生错误
func (fs *FluxServer) httpErrorAdapting(inErr error, ctx echo.Context) {
	iErr, isie := inErr.(*flux.InvokeError)
	// 解析非Flux.InvokeError的消息
	if !isie {
		code := https.StatusInternalServerError
		msg := "INTERNAL:SERVER_ERROR"
		if he, ishe := inErr.(*echo.HTTPError); ishe {
			code = he.Code
			msg = he.Error()
			if mstr, isms := he.Message.(string); isms {
				msg = mstr
			}
		}
		iErr = flux.NewInvokeError(code, msg, inErr)
	}
	// 统一异常处理：flux.context仅当找到路由匹配元数据才存在
	var requestId string
	var headers https.Header
	if fc, ok := ctx.Get(_echoAttrRoutedContext).(*internal.ContextWrapper); ok {
		requestId = fc.RequestId()
		headers = fc.ResponseWriter().Headers()
	} else {
		requestId = "ID:NON_CONTEXT"
	}
	var outErr error
	if ctx.Request().Method == https.MethodHead {
		ctx.Response().Header().Set("X-Error-Message", iErr.Message)
		outErr = ctx.NoContent(iErr.StatusCode)
	} else {
		outErr = fs.httpAdaptWriter.WriteError(ctx.Response(), requestId, headers, iErr)
	}
	if nil != outErr {
		logger.Errorf("Error responding(%s): ", requestId, outErr)
	}
}

func (fs *FluxServer) toHttpServerPattern(uri string) string {
	// /api/{userId} -> /api/:userId
	replaced := strings.Replace(uri, "}", "", -1)
	if len(replaced) < len(uri) {
		return strings.Replace(replaced, "{", ":", -1)
	} else {
		return uri
	}
}

func (fs *FluxServer) getVersionEndpoint(routeKey string) (*internal.MultiVersionEndpoint, bool) {
	if mve, ok := fs.endpointMvMap[routeKey]; ok {
		return mve, false
	} else {
		mve = internal.NewMultiVersionEndpoint()
		fs.endpointMvMap[routeKey] = mve
		return mve, true
	}
}

func (*FluxServer) requestPrepare() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Body缓存，允许通过 GetBody 多次读取Body
			request := c.Request()
			data, err := ioutil.ReadAll(request.Body)
			if nil != err {
				return &flux.InvokeError{
					StatusCode: flux.StatusBadRequest,
					Message:    "REQUEST:BODY_PREPARE",
					Internal:   fmt.Errorf("read req-body, method: %s, uri:%s, err: %w", request.Method, request.RequestURI, err),
				}
			}
			request.GetBody = func() (io.ReadCloser, error) {
				return ioutil.NopCloser(bytes.NewBuffer(data)), nil
			}
			// 恢复Body，但ParseForm解析后，request.Body无法重读，需要通过GetBody
			request.Body = ioutil.NopCloser(bytes.NewBuffer(data))
			if err := request.ParseForm(); nil != err {
				return &flux.InvokeError{
					StatusCode: flux.StatusBadRequest,
					Message:    "REQUEST:FORM_PARSING",
					Internal:   fmt.Errorf("parsing req-form, method: %s, uri:%s, err: %w", request.Method, request.RequestURI, err),
				}
			} else {
				return next(c)
			}
		}
	}
}

func (fs *FluxServer) ensure() *FluxServer {
	if fs.httpServer == nil {
		logger.Panicf("Call must after InitServer()")
	}
	return fs
}

// HttpServer 返回Http服务器实例
func (fs *FluxServer) HttpServer() *echo.Echo {
	return fs.ensure().httpServer
}

// AddHttpInterceptor 添加Http前拦截器。将在Http被路由到对应Handler之前执行
func (fs *FluxServer) AddHttpInterceptor(m echo.MiddlewareFunc) {
	fs.ensure().httpServer.Pre(m)
}

// AddHttpMiddleware 添加Http中间件。在Http路由到对应Handler后执行
func (fs *FluxServer) AddHttpMiddleware(m echo.MiddlewareFunc) {
	fs.ensure().httpServer.Use(m)
}

// AddHttpHandler 添加Http处理接口。
func (fs *FluxServer) AddHttpHandler(method, pattern string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) {
	fs.ensure().httpServer.Add(method, pattern, h, m...)
}

// SetHttpNotFoundHandler 设置Http路由失败的处理接口
func (fs *FluxServer) SetHttpNotFoundHandler(nfh echo.HandlerFunc) {
	echo.NotFoundHandler = nfh
}

// AddLifecycleHook 添加生命周期Hook接口：Startuper/Shutdowner接口
func (fs *FluxServer) AddLifecycleHook(lifecycleHook interface{}) {
	fs.dispatcher.AddLifecycleHook(lifecycleHook)
}

// AddContextPipeline 添加Http与Flux的Context桥接函数
func (fs *FluxServer) AddContextPipeline(bridgeFunc ContextPipelineFunc) {
	fs.pipelines = append(fs.pipelines, bridgeFunc)
}

func (*FluxServer) AddPrepareHook(ph flux.PrepareHook) {
	ext.AddPrepareHook(ph)
}

func (*FluxServer) SetExchange(protoName string, exchange flux.Exchange) {
	ext.SetExchange(protoName, exchange)
}

func (*FluxServer) SetFactory(typeName string, f flux.Factory) {
	ext.SetFactory(typeName, f)
}

func (*FluxServer) AddGlobalFilter(filter flux.Filter) {
	ext.AddGlobalFilter(filter)
}

func (*FluxServer) AddSelectiveFilter(filter flux.Filter) {
	ext.AddSelectiveFilter(filter)
}

func (*FluxServer) SetRegistryFactory(protoName string, factory ext.RegistryFactory) {
	ext.SetRegistryFactory(protoName, factory)
}

func (*FluxServer) SetSerializer(typeName string, serializer flux.Serializer) {
	ext.SetSerializer(typeName, serializer)
}
