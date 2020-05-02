package server

import (
	"bytes"
	context "context"
	"encoding/json"
	"expvar"
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/internal"
	"github.com/bytepowered/flux/logger"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/random"
	"io"
	"io/ioutil"
	httplib "net/http"
	_ "net/http/pprof"
	"strings"
	"sync"
)

const (
	Banner        = "Flux-GO // Fast gateway for microservice: dubbo, grpc, http"
	VersionFormat = "Version // git.commit=%s, build.version=%s, build.date=%s"
)

const (
	defaultHttpVersionHeader = "X-Version"
	defaultHttpVersionValue  = "v1"
)

const (
	configHttpRootName      = "HttpServer"
	configHttpVersionHeader = "version-header"
	configHttpDebugEnable   = "debug"
	configHttpTlsCertFile   = "tls-cert-file"
	configHttpTlsKeyFile    = "tls-key-file"
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

// Bridge func
type ContextBridgeFunc func(echo.Context, flux.Context)

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
	contextBridges    []ContextBridgeFunc
	globals           flux.Config
}

func NewFluxServer() *FluxServer {
	return &FluxServer{
		httpVisits:     expvar.NewInt("HttpVisits"),
		dispatcher:     internal.NewDispatcher(),
		endpointMvMap:  make(map[string]*internal.MultiVersionEndpoint),
		contextPool:    sync.Pool{New: internal.NewFxContext},
		contextBridges: make([]ContextBridgeFunc, 0),
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
	httpConfig := ext.ConfigFactory()("flux.http", fs.globals.Map(configHttpRootName))
	fs.httpVersionHeader = httpConfig.StringOrDefault(configHttpVersionHeader, defaultHttpVersionHeader)
	fs.httpServer = echo.New()
	fs.httpServer.HideBanner = true
	fs.httpServer.HidePort = true
	fs.httpServer.HTTPErrorHandler = fs.httpErrorAdapting
	// Http拦截器
	fs.AddHttpInterceptor(middleware.CORS())
	fs.AddHttpInterceptor(fs.prepareRequest())
	// Http debug features
	if httpConfig.BooleanOrDefault(configHttpDebugEnable, false) {
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
	httpConfig := ext.ConfigFactory()("flux.http", fs.globals.Map(configHttpRootName))
	address := fmt.Sprintf("%s:%d", httpConfig.String("address"), httpConfig.Int64("port"))
	certFile := httpConfig.String(configHttpTlsCertFile)
	keyFile := httpConfig.String(configHttpTlsKeyFile)
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
		case httplib.MethodGet, httplib.MethodPost, httplib.MethodDelete, httplib.MethodPut,
			httplib.MethodHead, httplib.MethodOptions, httplib.MethodPatch, httplib.MethodTrace:
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
	ctx.Reattach(c, endpoint)
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
		// Bridge Echo <-> Flux
		for _, cbf := range fs.contextBridges {
			cbf(c, newCtx)
		}
		c.Set(_echoAttrRoutedContext, newCtx)
		defer fs.release(newCtx)
		logger.Infof("Received request, id: %s, method: %s, uri: %s, version: %s",
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

func (fs *FluxServer) httpErrorAdapting(err error, ctx echo.Context) {
	iErr, ok := err.(*flux.InvokeError)
	if !ok {
		iErr = &flux.InvokeError{
			StatusCode: flux.StatusBadRequest,
			Message:    "REQUEST:FORM_PARSING",
			Internal:   err,
		}
	}
	fc := ctx.Get(_echoAttrRoutedContext)
	// 统一异常处理：flux.context仅当找到路由匹配元数据才存在
	if fxCtx, ok := fc.(*internal.FxContext); ok {
		if err := fs.httpAdaptWriter.WriteError(ctx.Response(), fxCtx.RequestId(), fxCtx.ResponseWriter().Headers(), iErr); nil != err {
			logger.Errorf("Response errors: ", err)
		}
	} else {
		if err := fs.httpAdaptWriter.WriteError(ctx.Response(), "UNROUTED_REQUEST", httplib.Header{}, iErr); nil != err {
			logger.Errorf("Proxy errors: ", err)
		}
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
	debugHandler := echo.WrapHandler(httplib.DefaultServeMux)
	fs.httpServer.GET("/debug/vars", debugHandler, authMiddleware)
	fs.httpServer.GET("/debug/pprof/*", debugHandler, authMiddleware)
	fs.httpServer.GET("/debug/endpoints", func(c echo.Context) error {
		m := make(map[string]interface{})
		for k, v := range fs.endpointMvMap {
			m[k] = v.ToSerializableMap()
		}
		if data, err := json.Marshal(m); nil != err {
			return err
		} else {
			return c.JSONBlob(200, data)
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

// AddContextBridge 添加Http与Flux的Context桥接参数
func (fs *FluxServer) AddContextBridge(bridgeFunc ContextBridgeFunc) {
	fs.contextBridges = append(fs.contextBridges, bridgeFunc)
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
