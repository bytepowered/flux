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
	"github.com/labstack/gommon/random"
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
	DefaultHttpVersionHeader = "X-Version"
)

const (
	ConfigHttpRootName      = "HttpServer"
	ConfigHttpVersionHeader = "version-header"
	ConfigHttpDebugEnable   = "debug"
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
)

// Context pipeline
type ContextPipelineFunc func(echo.Context, flux.Context)

// FluxServer
type FluxServer struct {
	httpServer        *echo.Echo
	httpVisits        *expvar.Int
	httpAdaptWriter   internal.FxHttpWriter
	httpNotFound      echo.HandlerFunc
	httpVersionHeader string
	dispatcher        *internal.FxDispatcher
	endpointMvMap     map[string]*internal.MultiVersionEndpoint
	contextPool       sync.Pool
	snowflakeId       *snowflake.Node
	pipelines         []ContextPipelineFunc
	globals           flux.Config
}

func LoadHttpServerConfig(globals flux.Config) flux.Config {
	return ext.ConfigFactory()("flux.http", globals.Map(ConfigHttpRootName))
}

func NewFluxServer() *FluxServer {
	id, _ := snowflake.NewNode(1)
	return &FluxServer{
		httpVisits:    expvar.NewInt("HttpVisits"),
		dispatcher:    internal.NewDispatcher(),
		endpointMvMap: make(map[string]*internal.MultiVersionEndpoint),
		contextPool:   sync.Pool{New: internal.NewFxContext},
		pipelines:     make([]ContextPipelineFunc, 0),
		snowflakeId:   id,
	}
}

// Prepare Call before init and startup
func (fs *FluxServer) Prepare(globals flux.Config, hooks ...flux.PrepareHook) error {
	for _, prepare := range append(ext.PrepareHooks(), hooks...) {
		if err := prepare(globals); nil != err {
			return err
		}
	}
	return nil
}

// Init : Call before startup
func (fs *FluxServer) Init(globals flux.Config) error {
	fs.globals = globals
	// Http server
	httpConfig := LoadHttpServerConfig(fs.globals)
	fs.httpVersionHeader = httpConfig.StringOrDefault(ConfigHttpVersionHeader, DefaultHttpVersionHeader)
	fs.httpServer = echo.New()
	fs.httpServer.HideBanner = true
	fs.httpServer.HidePort = true
	fs.httpServer.HTTPErrorHandler = fs.httpErrorAdapting
	// Http拦截器
	fs.AddHttpInterceptor(middleware.CORS())
	fs.AddHttpInterceptor(fs.prepareRequest())
	// Http debug features
	if httpConfig.BooleanOrDefault(ConfigHttpDebugEnable, false) {
		fs.debugFeatures(httpConfig)
	}
	return fs.dispatcher.Init(globals)
}

// Startup server
func (fs *FluxServer) Startup(version flux.BuildInfo) error {
	fs.checkInit()
	logger.Info(Banner)
	logger.Infof(VersionFormat, version.CommitId, version.Version, version.Date)
	if err := fs.dispatcher.Startup(); nil != err {
		return err
	}
	eventCh := make(chan flux.EndpointEvent, 2)
	defer close(eventCh)
	if err := fs.dispatcher.WatchRegistry(eventCh); nil != err {
		return fmt.Errorf("start registry watching: %w", err)
	} else {
		go fs.handleHttpRouteEvent(eventCh)
	}
	// Start http server at last
	httpConfig := LoadHttpServerConfig(fs.globals)
	address := fmt.Sprintf("%s:%d", httpConfig.String("address"), httpConfig.Int64("port"))
	certFile := httpConfig.String(ConfigHttpTlsCertFile)
	keyFile := httpConfig.String(ConfigHttpTlsKeyFile)
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

func (fs *FluxServer) handleHttpRouteEvent(events <-chan flux.EndpointEvent) {
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
			logger.Errorf("Unsupported http method: %s, ignore", eEndpoint.HttpMethod)
			continue
		}
		switch event.Type {
		case flux.EndpointEventAdded:
			logger.Infof("Endpoint new: [%s@%s] %s", eEndpoint.Version, event.HttpMethod, pattern)
			vEndpoint.Update(eEndpoint.Version, &eEndpoint)
			if isNew {
				logger.Infof("HTTP router: [%s] %s", event.HttpMethod, pattern)
				fs.httpServer.Add(event.HttpMethod, pattern, fs.generateRouter(vEndpoint))
			}
		case flux.EndpointEventUpdated:
			logger.Infof("Endpoint update: [%s@%s] %s", eEndpoint.Version, event.HttpMethod, pattern)
			vEndpoint.Update(eEndpoint.Version, &eEndpoint)
		case flux.EndpointEventRemoved:
			logger.Infof("Endpoint removed: [%s] %s", event.HttpMethod, pattern)
			vEndpoint.Delete(eEndpoint.Version)
		}
	}
}

func (fs *FluxServer) acquire(c echo.Context, endpoint *flux.Endpoint) *internal.FxContext {
	ctx := fs.contextPool.Get().(*internal.FxContext)
	reqId := fs.snowflakeId.Generate().Base64()
	ctx.Reattach(reqId, c, endpoint)
	return ctx
}

func (fs *FluxServer) release(c *internal.FxContext) {
	c.Release()
	fs.contextPool.Put(c)
}

func (fs *FluxServer) generateRouter(mvEndpoint *internal.MultiVersionEndpoint) echo.HandlerFunc {
	return func(c echo.Context) error {
		defer func() {
			if err := recover(); err != nil {
				logger.Error(err)
			}
		}()
		fs.httpVisits.Add(1)
		httpRequest := c.Request()
		// Multi version selection
		version := httpRequest.Header.Get(fs.httpVersionHeader)
		endpoint, found := mvEndpoint.Get(version)
		newCtx := fs.acquire(c, endpoint)
		// Context exchange: Echo <-> Flux
		for _, pipe := range fs.pipelines {
			pipe(c, newCtx)
		}
		c.Set(_echoAttrRoutedContext, newCtx)
		defer fs.release(newCtx)
		logger.Infof("Received request, id: %s, method: %s, uri: %s, endpoint.ver(%s)",
			newCtx.RequestId(), httpRequest.Method, httpRequest.RequestURI, version)
		if !found {
			return errEndpointVersionNotFound
		}
		if err := fs.dispatcher.Dispatch(newCtx); nil != err {
			return err
		} else {
			return fs.httpAdaptWriter.WriteResponse(newCtx)
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
		iErr = &flux.InvokeError{
			StatusCode: code,
			Message:    msg,
			Internal:   inErr,
		}
	}
	// 统一异常处理：flux.context仅当找到路由匹配元数据才存在
	var requestId string
	var headers https.Header
	if fc, ok := ctx.Get(_echoAttrRoutedContext).(*internal.FxContext); ok {
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

func (fs *FluxServer) debugFeatures(httpConfig flux.Config) {
	baFactory := ext.ConfigFactory()("flux.http.basic-auth", httpConfig.Map("BasicAuth"))
	username := baFactory.StringOrDefault("username", "fluxgo")
	password := baFactory.StringOrDefault("password", random.String(8))
	logger.Infof("Http debug feature: <Enabled>, basic-auth: username=%s, password=%s", username, password)
	authMiddleware := middleware.BasicAuth(func(u string, p string, c echo.Context) (bool, error) {
		return u == username && p == password, nil
	})
	debugHandler := echo.WrapHandler(https.DefaultServeMux)
	fs.httpServer.GET(DebugPathVars, debugHandler, authMiddleware)
	fs.httpServer.GET(DebugPathPprof, debugHandler, authMiddleware)
	fs.httpServer.GET(DebugPathEndpoints, func(c echo.Context) error {
		decoder := ext.GetSerializer(ext.TypeNameSerializerJson)
		if data, err := decoder.Marshal(queryEndpoints(fs.endpointMvMap, c)); nil != err {
			return err
		} else {
			return c.JSONBlob(flux.StatusOK, data)
		}
	}, authMiddleware)
}

func (*FluxServer) prepareRequest() echo.MiddlewareFunc {
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

func (fs *FluxServer) checkInit() {
	if fs.httpServer == nil {
		logger.Panicf("Call must after init()")
	}
}

// HttpServer 返回Http服务器实例
func (fs *FluxServer) HttpServer() *echo.Echo {
	fs.checkInit()
	return fs.httpServer
}

// AddHttpInterceptor 添加Http前拦截器。将在Http被路由到对应Handler之前执行
func (fs *FluxServer) AddHttpInterceptor(m echo.MiddlewareFunc) {
	fs.checkInit()
	fs.httpServer.Pre(m)
}

// AddHttpMiddleware 添加Http中间件。在Http路由到对应Handler后执行
func (fs *FluxServer) AddHttpMiddleware(m echo.MiddlewareFunc) {
	fs.checkInit()
	fs.httpServer.Use(m)
}

// AddHttpHandler 添加Http处理接口。
func (fs *FluxServer) AddHttpHandler(method, pattern string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) {
	fs.checkInit()
	fs.httpServer.Add(method, pattern, h, m...)
}

// SetHttpNotFoundHandler 设置Http路由失败的处理接口
func (fs *FluxServer) SetHttpNotFoundHandler(nfh echo.HandlerFunc) {
	fs.checkInit()
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

func (*FluxServer) AddFilter(filter flux.Filter) {
	ext.AddFilter(filter)
}

func (*FluxServer) SetLogger(logger flux.Logger) {
	ext.SetLogger(logger)
}

func (*FluxServer) SetRegistryFactory(protoName string, factory ext.RegistryFactory) {
	ext.SetRegistryFactory(protoName, factory)
}

func (*FluxServer) SetSerializer(typeName string, serializer flux.Serializer) {
	ext.SetSerializer(typeName, serializer)
}
