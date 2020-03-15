package server

import (
	stdContext "context"
	"encoding/json"
	"expvar"
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/extension"
	"github.com/bytepowered/flux/internal"
	"github.com/bytepowered/flux/logger"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/random"
	httplib "net/http"
	_ "net/http/pprof"
	"strings"
	"sync"
	"time"
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
	configHttpSectionName   = "HttpServer"
	configHttpVersionHeader = "version-header"
	configHttpDebugEnable   = "debug"
	configHttpTlsCertFile   = "tls-cert-file"
	configHttpTlsKeyFile    = "tls-key-file"
)

var (
	errEndpointVersionNotFound = &flux.InvokeError{
		StatusCode: flux.StatusNotFound,
		Message:    "ENDPOINT_VERSION_NOT_FOUND",
	}
)

// FluxServer
type FluxServer struct {
	httpServer        *echo.Echo
	httpVisits        *expvar.Int
	httpAdapter       internal.HttpAdapter
	httpInterceptors  []echo.MiddlewareFunc
	httpMiddlewares   []echo.MiddlewareFunc
	httpVersionHeader string
	dispatcher        *internal.Dispatcher
	endpointMvMap     map[string]*internal.MultiVersionEndpoint
	contextPool       sync.Pool
	globals           flux.Config
}

func NewFluxServer() *FluxServer {
	return &FluxServer{
		httpVisits:       expvar.NewInt("HttpVisits"),
		httpInterceptors: make([]echo.MiddlewareFunc, 0),
		httpMiddlewares:  make([]echo.MiddlewareFunc, 0),
		dispatcher:       internal.NewDispatcher(),
		endpointMvMap:    make(map[string]*internal.MultiVersionEndpoint),
		contextPool:      sync.Pool{New: internal.NewContext},
	}
}

func (fs *FluxServer) Init(globals flux.Config) error {
	fs.globals = globals
	// Http server
	httpConfig := fs.globals.Config(configHttpSectionName)
	fs.httpVersionHeader = httpConfig.StringOrDefault(configHttpVersionHeader, defaultHttpVersionHeader)
	fs.httpServer = echo.New()
	fs.httpServer.HideBanner = true
	fs.httpServer.HidePort = true
	// Http拦截器
	fs.AddHttpInterceptor(middleware.CORS())
	fs.httpServer.Pre(fs.httpInterceptors...)
	fs.httpServer.Use(fs.httpMiddlewares...)
	// Http debug features
	if httpConfig.BooleanOrDefault(configHttpDebugEnable, false) {
		fs.debugFeatures(httpConfig)
	}
	return fs.dispatcher.Init(globals)
}

// Start server
func (fs *FluxServer) Start(version flux.BuildInfo) error {
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
	httpConfig := fs.globals.Config(configHttpSectionName)
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
func (fs *FluxServer) Shutdown() {
	// Stop http server
	ctx, cancel := stdContext.WithTimeout(stdContext.Background(), 10*time.Second)
	defer cancel()
	_ = fs.httpServer.Shutdown(ctx)
	if ctx.Err() == stdContext.DeadlineExceeded {
		logger.Error("HTTP Server forcefully stopped after timeout")
	} else {
		logger.Info("HTTP Server gracefully stopped")
	}
	// Stop dispatcher
	if err := fs.dispatcher.Shutdown(); nil != err {
		logger.Error("Dispatcher shutdown:", err)
	}
}

// AddHttpInterceptor 添加Http前拦截器。将在Http被路由到对应Handler之前执行
func (fs *FluxServer) AddHttpInterceptor(m echo.MiddlewareFunc) {
	fs.httpInterceptors = append(fs.httpInterceptors, m)
}

// AddHttpMiddleware 添加Http中间件。在Http路由到对应Handler后执行
func (fs *FluxServer) AddHttpMiddleware(m echo.MiddlewareFunc) {
	fs.httpMiddlewares = append(fs.httpMiddlewares, m)
}

func (fs *FluxServer) handleHttpRouteEvent(events <-chan flux.EndpointEvent) {
	for event := range events {
		pattern := fs.httpAdapter.Pattern(event.HttpPattern)
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
			logger.Infof("Endpoint new: %s@%s:%s", eEndpoint.Version, event.HttpMethod, pattern)
			vEndpoint.Update(eEndpoint.Version, &eEndpoint)
			if isNew {
				logger.Infof("HTTP router: %s:%s", event.HttpMethod, pattern)
				fs.httpServer.Add(event.HttpMethod, pattern, fs.generateRouter(vEndpoint))
			}
		case flux.EndpointEventUpdated:
			logger.Infof("Endpoint update: %s@%s:%s", eEndpoint.Version, event.HttpMethod, pattern)
			vEndpoint.Update(eEndpoint.Version, &eEndpoint)
		case flux.EndpointEventRemoved:
			logger.Infof("Endpoint removed: %s:%s", event.HttpMethod, pattern)
			vEndpoint.Delete(eEndpoint.Version)
		}
	}
}

func (fs *FluxServer) acquire(c echo.Context, endpoint *flux.Endpoint) *internal.Context {
	context := fs.contextPool.Get().(*internal.Context)
	context.Reattach(c, endpoint)
	return context
}

func (fs *FluxServer) release(c *internal.Context) {
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
		version := httpRequest.Header.Get(fs.httpVersionHeader)
		endpoint, found := mvEndpoint.Get(version)
		if !found {
			version = defaultHttpVersionValue
			endpoint, found = mvEndpoint.Get(version)
		}
		context := fs.acquire(c, endpoint)
		defer fs.release(context)
		logger.Infof("Received request, id: %s, method: %s, uri: %s, version: %s",
			context.RequestId(), httpRequest.Method, httpRequest.RequestURI, version)
		if !found {
			return fs.httpAdapter.WriteError(context, errEndpointVersionNotFound)
		}
		if err := httpRequest.ParseForm(); nil != err {
			logger.Errorf("Parse http req-form, method: %s, uri:%s", httpRequest.Method, httpRequest.RequestURI)
			return fs.httpAdapter.WriteError(context, &flux.InvokeError{
				StatusCode: flux.StatusBadRequest,
				Message:    "REQUEST:FORM_PARSING",
				Internal:   err,
			})
		}
		if err := fs.dispatcher.Dispatch(context); nil != err {
			return fs.httpAdapter.WriteError(context, err)
		} else {
			return fs.httpAdapter.WriteResponse(context)
		}
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
	basicAuthC := httpConfig.Config("BasicAuth")
	username := basicAuthC.StringOrDefault("username", "flux")
	password := basicAuthC.StringOrDefault("password", random.String(8))
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

func (*FluxServer) SetExtendExchange(protoName string, exchange flux.Exchange) {
	extension.SetExchange(protoName, exchange)
}

func (*FluxServer) SetExtendFactory(typeName string, exchange flux.Exchange) {
	extension.SetExchange(typeName, exchange)
}

func (*FluxServer) SetExtendGlobalFilter(filter flux.Filter) {
	extension.AddGlobalFilter(filter)
}

func (*FluxServer) SetExtendLogger(logger flux.Logger) {
	extension.SetLogger(logger)
}

func (*FluxServer) SetExtendRegistryFactory(protoName string, factory extension.RegistryFactory) {
	extension.SetRegistryFactory(protoName, factory)
}

func (*FluxServer) SetExtendSerializer(typeName string, serializer flux.Serializer) {
	extension.SetSerializer(typeName, serializer)
}
