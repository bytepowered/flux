package internal

import (
	stdContext "context"
	"expvar"
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/exchange/dubbo"
	"github.com/bytepowered/flux/exchange/echoex"
	"github.com/bytepowered/flux/exchange/http"
	"github.com/bytepowered/flux/extension"
	"github.com/bytepowered/flux/filter"
	_ "github.com/bytepowered/flux/filter"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/registry"
	"github.com/bytepowered/flux/serializer"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	httplib "net/http"
	_ "net/http/pprof"
	"strings"
	"sync"
	"time"
)

func init() {
	// serializer
	defaultSerializer := serializer.NewJsonSerializer()
	extension.SetSerializer(extension.TypeNameSerializerDefault, defaultSerializer)
	extension.SetSerializer(extension.TypeNameSerializerJson, defaultSerializer)
	// Registry
	extension.SetRegistryFactory(extension.TypeNameRegistryActive, registry.ZookeeperRegistryFactory)
	extension.SetRegistryFactory(extension.TypeNameRegistryZookeeper, registry.ZookeeperRegistryFactory)
	extension.SetRegistryFactory(extension.TypeNameRegistryEcho, registry.EchoRegistryFactory)
	// exchanges
	extension.SetExchange(flux.ProtocolEcho, echoex.NewEchoExchange())
	extension.SetExchange(flux.ProtocolHttp, http.NewHttpExchange())
	extension.SetExchange(flux.ProtocolDubbo, dubbo.NewDubboExchange())
	// filters
	extension.SetFactory(filter.TypeNameFilterJWTVerification, filter.JwtVerificationFilterFactory)
	extension.SetFactory(filter.TypeNameFilterPermissionVerification, filter.PermissionVerificationFactory)
	extension.SetFactory(filter.TypeNameFilterJWTVerification, filter.JwtVerificationFilterFactory)
	// global filters
	extension.SetGlobalFilter(filter.NewParameterParsingFilter())
}

const (
	DefaultServerName = "Flux-GO"
	FluxBanner        = "Flux-GO // Fast gateway for microservice: dubbo, grpc, http"
	FluxVersion       = "Version // git.commit=%s, build.version=%s, build.date=%s"
)

const (
	defaultHttpVersionHeader = "X-Version"
	defaultHttpBodyLimit     = "100K"
	defaultHttpVersionValue  = "v1"
)

const (
	configHttpSectionName   = "HttpServer"
	configHttpBodyLimit     = "body-limit"
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

// Application
type Application struct {
	httpServer        *echo.Echo
	httpVisits        *expvar.Int
	httpAdapter       HttpAdapter
	httpVersionHeader string
	dispatcher        *Dispatcher
	endpointMvMap     map[string]*MultiVersionEndpoint
	contextPool       sync.Pool
	globals           flux.Config
}

func NewApplication() *Application {
	return &Application{
		httpVisits:    expvar.NewInt("HttpVisits"),
		dispatcher:    newDispatcher(),
		endpointMvMap: make(map[string]*MultiVersionEndpoint),
		contextPool:   sync.Pool{New: newContext},
	}
}

func (a *Application) Init() error {
	a.globals = extension.GetGlobalConfig()
	if err := a.dispatcher.Init(); nil != err {
		return err
	}
	// Http server
	httpConfig := a.globals.Config(configHttpSectionName)
	a.httpVersionHeader = httpConfig.StringOrDefault(configHttpVersionHeader, defaultHttpVersionHeader)
	a.httpServer = echo.New()
	a.httpServer.HideBanner = true
	a.httpServer.HidePort = true
	a.httpServer.Use(middleware.CORS())
	a.httpServer.Use(middleware.BodyLimit(httpConfig.StringOrDefault(configHttpBodyLimit, defaultHttpBodyLimit)))
	// Http debug features
	if httpConfig.BooleanOrDefault(configHttpDebugEnable, false) {
		a.enabledDebugFeatures(httpConfig)
	}
	return nil
}

// Start server
func (a *Application) Start(version flux.BuildInfo) error {
	logger.Info(FluxBanner)
	logger.Infof(FluxVersion, version.CommitId, version.Version, version.Date)
	if err := a.dispatcher.Startup(); nil != err {
		return err
	}
	events := make(chan flux.EndpointEvent, 2)
	defer close(events)
	if err := a.dispatcher.WatchRegistry(events); nil != err {
		return fmt.Errorf("start registry watching: %w", err)
	} else {
		go a.handleHttpRouteEvent(events)
	}
	// Start http server at last
	httpConfig := a.globals.Config(configHttpSectionName)
	address := fmt.Sprintf("%s:%d", httpConfig.String("address"), httpConfig.Int64("port"))
	certFile := httpConfig.String(configHttpTlsCertFile)
	keyFile := httpConfig.String(configHttpTlsKeyFile)
	if certFile != "" && keyFile != "" {
		logger.Infof("HttpServer(HTTP/2 TLS) starting: %s", address)
		return a.httpServer.StartTLS(address, certFile, keyFile)
	} else {
		logger.Infof("HttpServer starting: %s", address)
		return a.httpServer.Start(address)
	}
}

// Shutdown to cleanup resources
func (a *Application) Shutdown() {
	// Stop http server
	ctx, cancel := stdContext.WithTimeout(stdContext.Background(), 10*time.Second)
	defer cancel()
	_ = a.httpServer.Shutdown(ctx)
	if ctx.Err() == stdContext.DeadlineExceeded {
		logger.Error("HTTP Server forcefully stopped after timeout")
	} else {
		logger.Info("HTTP Server gracefully stopped")
	}
	// Stop dispatcher
	if err := a.dispatcher.Shutdown(); nil != err {
		logger.Error("Dispatcher shutdown:", err)
	}
}

func (a *Application) handleHttpRouteEvent(events <-chan flux.EndpointEvent) {
	for event := range events {
		pattern := a.httpAdapter.pattern(event.HttpPattern)
		routeKey := fmt.Sprintf("%s#%s", event.HttpMethod, pattern)
		vEndpoint, isNew := a.getVersionEndpoint(routeKey)
		// Check http method
		event.Endpoint.HttpMethod = strings.ToUpper(event.Endpoint.HttpMethod)
		eEndpoint := event.Endpoint
		switch eEndpoint.HttpMethod {
		case httplib.MethodGet, httplib.MethodPost, httplib.MethodDelete, httplib.MethodPut,
			httplib.MethodHead, httplib.MethodOptions, httplib.MethodPatch, httplib.MethodTrace:
		default:
			// http.MethodConnect, and Others
			logger.Errorf("Unsupported http method: %s, ignore", eEndpoint.HttpMethod)
			continue
		}
		switch event.Type {
		case flux.EndpointEventAdded:
			logger.Infof("Add HTTP endpoint: %s@%s %s", eEndpoint.Version, event.HttpMethod, pattern)
			vEndpoint.Update(eEndpoint.Version, &eEndpoint)
			if isNew {
				logger.Infof("Register HTTP routing: %s %s", event.HttpMethod, pattern)
				a.httpServer.Add(event.HttpMethod, pattern, a.generateHttpRouting(vEndpoint))
			}
		case flux.EndpointEventUpdated:
			logger.Infof("Update HTTP endpoint: %s@%s %s", eEndpoint.Version, event.HttpMethod, pattern)
			vEndpoint.Update(eEndpoint.Version, &eEndpoint)
		case flux.EndpointEventRemoved:
			logger.Infof("Remove HTTP endpoint: %s %s", event.HttpMethod, pattern)
			vEndpoint.Delete(eEndpoint.Version)
		}
	}
}

func (a *Application) acquire(c echo.Context, endpoint *flux.Endpoint) *context {
	context := a.contextPool.Get().(*context)
	context.reattach(c, endpoint)
	return context
}

func (a *Application) release(c *context) {
	c.release()
	a.contextPool.Put(c)
}

func (a *Application) generateHttpRouting(mvEndpoint *MultiVersionEndpoint) echo.HandlerFunc {
	return func(c echo.Context) error {
		defer func() {
			if err := recover(); err != nil {
				logger.Error(err)
			}
		}()
		a.httpVisits.Add(1)
		httpRequest := c.Request()
		version := httpRequest.Header.Get(a.httpVersionHeader)
		endpoint, found := mvEndpoint.Get(version)
		if !found {
			version = defaultHttpVersionValue
			endpoint, found = mvEndpoint.Get(version)
		}
		context := a.acquire(c, endpoint)
		defer a.release(context)
		logger.Infof("Received request, id: %s, method: %s, uri: %s, version: %s", context.RequestId(), httpRequest.Method, httpRequest.RequestURI, version)
		if !found {
			return a.httpAdapter.error(context, errEndpointVersionNotFound)
		}
		if err := httpRequest.ParseForm(); nil != err {
			logger.Errorf("Parse http req-form, method: %s, uri:%s", httpRequest.Method, httpRequest.RequestURI)
			return a.httpAdapter.error(context, &flux.InvokeError{
				StatusCode: flux.StatusBadRequest,
				Message:    "REQUEST:FORM_PARSING",
				Internal:   err,
			})
		}
		logger.Infof("Dispatch request, id: %s, endpoint: %s#%s", context.RequestId(), endpoint.HttpMethod, endpoint.HttpPattern)
		if err := a.dispatcher.Dispatch(context); nil != err {
			return a.httpAdapter.error(context, err)
		} else {
			return a.httpAdapter.response(context)
		}
	}
}

func (a *Application) getVersionEndpoint(routeKey string) (*MultiVersionEndpoint, bool) {
	if mve, ok := a.endpointMvMap[routeKey]; ok {
		return mve, false
	} else {
		mve = &MultiVersionEndpoint{
			data: new(sync.Map),
		}
		a.endpointMvMap[routeKey] = mve
		return mve, true
	}
}
