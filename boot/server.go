package boot

import (
	goctx "context"
	"fmt"
	dubgo "github.com/apache/dubbo-go/config"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/admin"
	"github.com/bytepowered/flux/context"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/listen"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/pkg"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"time"
)

const (
	defaultBanner = "Flux-GO // Fast gateway for microservice: dubbo, grpc, http"
	VersionFormat = "Version // git.commit=%s, build.version=%s, build.date=%s"
)

const (
	DefaultHttpHeaderVersion = "X-Version"

	ListenServerIdDefault = "default"
	ListenServerIdAdmin   = "admin"
)

type (
	// Option 配置HttpServeEngine函数
	Option func(bs *BootstrapServer)
	// EndpointSelectFunc Http请求版本查找函数
	EndpointSelectFunc func(webc flux.WebContext) string
)

// BootstrapServer
type BootstrapServer struct {
	listenServers      map[string]flux.ListenServer
	ctxHooks           []flux.ContextHook
	endpointSelectFunc EndpointSelectFunc
	router             *Router
	started            chan struct{}
	stopped            chan struct{}
	banner             string
	routeTraceEnabled  bool
}

// WithServerContextHooks 配置请求Hook函数列表
func WithServerContextHooks(hooks ...flux.ContextHook) Option {
	return func(bs *BootstrapServer) {
		bs.ctxHooks = append(bs.ctxHooks, hooks...)
	}
}

// WithEndpointSelectFunc 配置Web请求版本选择函数
func WithEndpointSelectFunc(fun EndpointSelectFunc) Option {
	return func(bs *BootstrapServer) {
		bs.endpointSelectFunc = fun
	}
}

// WithBanner 配置服务Banner
func WithServerBanner(banner string) Option {
	return func(bs *BootstrapServer) {
		bs.banner = banner
	}
}

// WithPrepareHooks 配置服务启动预备阶段Hook函数列表
func WithPrepareHooks(hooks ...flux.PrepareHookFunc) Option {
	return func(bs *BootstrapServer) {
		bs.router.hooks = append(bs.router.hooks, hooks...)
	}
}

func WithListenServer(id string, server flux.ListenServer) Option {
	return func(bs *BootstrapServer) {
		bs.AddListenServer(id, server)
	}
}

func NewDefaultBootstrapServer(options ...Option) *BootstrapServer {
	opts := []Option{
		WithServerBanner(defaultBanner),
		WithEndpointSelectFunc(func(webc flux.WebContext) string {
			return webc.HeaderVar(DefaultHttpHeaderVersion)
		}),
		// Default ListenServer
		WithListenServer(ListenServerIdDefault,
			listen.NewServer(LoadListenServerConfig(ListenServerIdDefault), nil)),
		// Admin ListenServer
		WithListenServer(ListenServerIdAdmin,
			listen.NewServer(LoadListenServerConfig(ListenServerIdAdmin), nil,
				// 内部元数据查询
				listen.WithWebHandlers([]listen.WebHandlerTuple{
					{Method: "GET", Pattern: "/inspect/endpoints", Handler: admin.InspectEndpointsHandler},
					{Method: "GET", Pattern: "/inspect/services", Handler: admin.InspectServicesHandler},
				}),
			)),
	}
	return NewBootstrapServerWith(append(opts, options...)...)
}

func NewBootstrapServerWith(opts ...Option) *BootstrapServer {
	srv := &BootstrapServer{
		router:        NewRouter(),
		listenServers: make(map[string]flux.ListenServer, 2),
		ctxHooks:      make([]flux.ContextHook, 0, 4),
		started:       make(chan struct{}),
		stopped:       make(chan struct{}),
		banner:        defaultBanner,
	}
	for _, opt := range opts {
		opt(srv)
	}
	return srv
}

// Prepare Call before init and startup
func (s *BootstrapServer) Prepare() error {
	return s.router.Prepare()
}

// Initial
func (s *BootstrapServer) Initial() error {
	// Listen Server
	for id, srv := range s.listenServers {
		if err := srv.Init(LoadListenServerConfig(id)); nil != err {
			return err
		}
	}
	// Discovery
	for _, dis := range ext.GetEndpointDiscoveries() {
		if err := s.router.AddInitHook(dis, LoadEndpointDiscoveryConfig(dis.Id())); nil != err {
			return err
		}
	}
	return s.router.Initial()
}

func (s *BootstrapServer) Startup(build flux.Build) error {
	logger.Infof(VersionFormat, build.CommitId, build.Version, build.Date)
	if "" != s.banner {
		logger.Info(s.banner)
	}
	return s.start()
}

func (s *BootstrapServer) start() error {
	s.defaultListenServer()
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
	var errch chan error
	for _id, ls := range s.listenServers {
		go func(id string, server flux.ListenServer) {
			logger.Infow("ListenServer starting, server-id: " + id)
			errch <- server.Listen()
		}(_id, ls)
	}
	close(s.started)
	return <-errch
}

func (s *BootstrapServer) startDiscovery(endpoints chan flux.HttpEndpointEvent, services chan flux.BackendServiceEvent) error {
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
		defer logger.Info("Discovery event loop: STOP")
		for {
			select {
			case epEvt, ok := <-endpoints:
				if !ok {
					return
				}
				s.onHttpEndpointEvent(epEvt)

			case esEvt, ok := <-services:
				if !ok {
					return
				}
				s.onBackendServiceEvent(esEvt)
			}
		}
	}()
	return nil
}

func (s *BootstrapServer) route(webc flux.WebContext, server flux.ListenServer, endpoints *flux.MultiEndpoint) error {
	version := s.endpointSelectFunc(webc)
	endpoint, found := endpoints.LookupByVersion(version)
	defer func(requestId string) {
		if r := recover(); r != nil {
			trace := logger.Trace(requestId)
			if err, ok := r.(error); ok {
				trace.Errorw("SERVER:ROUTE:CRITICAL_PANIC", "error", err)
			} else {
				trace.Errorw("SERVER:ROUTE:CRITICAL_PANIC", "recover", r)
			}
			trace.Error(string(debug.Stack()))
		}
	}(webc.RequestId())
	if !found {
		if s.routeTraceEnabled {
			logger.Trace(webc.RequestId()).Infow("SERVER:ROUTE:NOT_FOUND",
				"http-pattern", []string{webc.Method(), webc.URI(), webc.URL().Path},
			)
		}
		return flux.ErrRouteNotFound
	}
	ctxw := context.NewAttachableContext(webc, endpoint)
	// route call
	logger.TraceContext(ctxw).Infow("SERVER:ROUTE:START")
	endcall := func(code int, start time.Time) {
		logger.TraceContext(ctxw).Infow("SERVER:ROUTE:END",
			"metric", ctxw.Metrics(),
			"elapses", time.Since(start).String(), "response.code", code)
	}
	start := time.Now()
	// Context hook
	for _, hook := range s.ctxHooks {
		hook(webc, ctxw)
	}
	// route and response
	if err := s.router.Route(ctxw); nil != err {
		defer endcall(err.StatusCode, start)
		if goctx.Canceled == err.CauseError || strings.Contains(err.CauseError.Error(), goctx.Canceled.Error()) {
			logger.TraceContext(ctxw).Infow("SERVER:ROUTE:CANCELED")
			return goctx.Canceled
		} else {
			logger.TraceContext(ctxw).Errorw("SERVER:ROUTE:ERROR", "error", err)
			err.MergeHeader(ctxw.Response().HeaderVars())
			return err
		}
	}
	r := ctxw.Response()
	defer endcall(r.StatusCode(), start)
	return server.Write(webc, r.HeaderVars(), r.StatusCode(), r.Payload())
}

func (s *BootstrapServer) onBackendServiceEvent(event flux.BackendServiceEvent) {
	service := event.Service
	initArguments(service.Arguments)
	switch event.EventType {
	case flux.EventTypeAdded:
		logger.Infow("SERVER:META:SERVICE:ADD",
			"service-id", service.ServiceId, "alias-id", service.AliasId)
		ext.SetBackendService(service)
		if "" != service.AliasId {
			ext.SetBackendServiceById(service.AliasId, service)
		}
	case flux.EventTypeUpdated:
		logger.Infow("SERVER:META:SERVICE:UPDATE",
			"service-id", service.ServiceId, "alias-id", service.AliasId)
		ext.SetBackendService(service)
		if "" != service.AliasId {
			ext.SetBackendServiceById(service.AliasId, service)
		}
	case flux.EventTypeRemoved:
		logger.Infow("SERVER:META:SERVICE:REMOVE",
			"service-id", service.ServiceId, "alias-id", service.AliasId)
		ext.RemoveBackendService(service.ServiceId)
		if "" != service.AliasId {
			ext.RemoveBackendService(service.AliasId)
		}
	}
}

func (s *BootstrapServer) onHttpEndpointEvent(event flux.HttpEndpointEvent) {
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
		logger.Infow("SERVER:META:ENDPOINT:ADD", "version", endpoint.Version, "method", method, "pattern", pattern)
		bind.Update(endpoint.Version, &endpoint)
		// 根据Endpoint属性，选择ListenServer来绑定
		if isreg {
			id := endpoint.GetAttr(flux.EndpointAttrTagServerId).GetString()
			if id == "" {
				id = ListenServerIdDefault
			}
			server, ok := s.GetListenServer(id)
			if ok {
				logger.Infow("SERVER:META:ENDPOINT:HTTP_HANDLER/"+id, "method", method, "pattern", pattern)
				server.AddHandler(method, pattern, s.newEndpointHandler(server, bind))
			} else {
				logger.Errorw("SERVER:META:ENDPOINT:LISTENER_MISSED/"+id, "method", method, "pattern", pattern)
			}
		}
	case flux.EventTypeUpdated:
		logger.Infow("SERVER:META:ENDPOINT:UPDATE", "version", endpoint.Version, "method", method, "pattern", pattern)
		bind.Update(endpoint.Version, &endpoint)
	case flux.EventTypeRemoved:
		logger.Infow("SERVER:META:ENDPOINT:REMOVE", "method", method, "pattern", pattern)
		bind.Delete(endpoint.Version)
	}
}

// Shutdown to cleanup resources
func (s *BootstrapServer) Shutdown(ctx goctx.Context) error {
	logger.Info("Server shutdown...")
	defer close(s.stopped)
	for id, server := range s.listenServers {
		if err := server.Close(ctx); nil != err {
			logger.Warnw("Server["+id+"] shutdown http server", "error", err)
		}
	}
	return s.router.Shutdown(ctx)
}

// GracefulShutdown
func (s *BootstrapServer) OnSignalShutdown(quit chan os.Signal, to time.Duration) {
	// 接收停止信号
	signal.Notify(quit, dubgo.ShutdownSignals...)
	<-quit
	logger.Infof("Server received shutdown signal, shutdown...")
	ctx, cancel := goctx.WithTimeout(goctx.Background(), to)
	defer cancel()
	if err := s.Shutdown(ctx); nil != err {
		logger.Error("Server shutdown, error: ", err)
	}
}

// StateStarted 返回一个Channel。当服务启动完成时，此Channel将被关闭。
func (s *BootstrapServer) StateStarted() <-chan struct{} {
	return s.started
}

// StateStopped 返回一个Channel。当服务停止后完成时，此Channel将被关闭。
func (s *BootstrapServer) StateStopped() <-chan struct{} {
	return s.stopped
}

// AddWebInterceptor 添加Http前拦截器到默认ListenerServer。将在Http被路由到对应Handler之前执行
func (s *BootstrapServer) AddWebInterceptor(m flux.WebInterceptor) {
	s.defaultListenServer().AddInterceptor(m)
}

// AddWebHandler 添加Http处理接口到默认ListenerServer。
func (s *BootstrapServer) AddWebHandler(method, pattern string, h flux.WebHandler, m ...flux.WebInterceptor) {
	s.defaultListenServer().AddHandler(method, pattern, h, m...)
}

// AddWebHttpHandler 添加Http处理接口到默认ListenerServer。
func (s *BootstrapServer) AddWebHttpHandler(method, pattern string, h http.Handler, m ...func(http.Handler) http.Handler) {
	s.defaultListenServer().AddHttpHandler(method, pattern, h, m...)
}

// SetWebNotfoundHandler 设置Http路由失败的处理接口到默认ListenerServer
func (s *BootstrapServer) SetWebNotfoundHandler(nfh flux.WebHandler) {
	s.defaultListenServer().SetNotfoundHandler(nfh)
}

// AddListenServer 添加指定ID
func (s *BootstrapServer) AddListenServer(id string, server flux.ListenServer) {
	s.listenServers[id] = pkg.RequireNotNil(server, "ListenServer is nil").(flux.ListenServer)
}

// GetListenServer 返回ListenServer实例
func (s *BootstrapServer) GetListenServer(id string) (flux.ListenServer, bool) {
	ls, ok := s.listenServers[id]
	return ls, ok
}

// AddContextHook 添加Http与Flux的Context桥接函数
func (s *BootstrapServer) AddContextHook(f flux.ContextHook) {
	s.ctxHooks = append(s.ctxHooks, f)
}

func (s *BootstrapServer) newEndpointHandler(server flux.ListenServer, endpoint *flux.MultiEndpoint) flux.WebHandler {
	return func(webc flux.WebContext) error {
		return s.route(webc, server, endpoint)
	}
}

func (s *BootstrapServer) selectMultiEndpoint(routeKey string, endpoint *flux.Endpoint) (*flux.MultiEndpoint, bool) {
	if mve, ok := ext.SelectEndpoint(routeKey); ok {
		return mve, false
	} else {
		return ext.RegisterEndpoint(routeKey, endpoint), true
	}
}

func (s *BootstrapServer) defaultListenServer() flux.ListenServer {
	count := len(s.listenServers)
	if count == 0 {
		logger.Panicf("Call must after InitialServer()")
	} else if count == 1 {
		for _, server := range s.listenServers {
			return server
		}
	}
	server, ok := s.listenServers[ListenServerIdDefault]
	if ok {
		return server
	}
	logger.Panicf("Default ListenServer not found, servers: %d", count)
	return nil
}

func LoadListenServerConfig(id string) *flux.Configuration {
	return flux.NewConfigurationOfNS(flux.NamespaceListenServer + "." + id)
}

func LoadEndpointDiscoveryConfig(id string) *flux.Configuration {
	return flux.NewConfigurationOfNS(flux.NamespaceEndpointDiscoveryServices + "." + id)
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
