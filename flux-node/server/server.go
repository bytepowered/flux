package server

import (
	goctx "context"
	"fmt"
	dubgo "github.com/apache/dubbo-go/config"
	"github.com/bytepowered/flux/flux-inspect"
	"github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-node/ext"
	"github.com/bytepowered/flux/flux-node/listener"
	"github.com/bytepowered/flux/flux-node/logger"
	"github.com/bytepowered/flux/flux-pkg"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/net/context"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strings"
	"time"
)

const (
	defaultBanner = "Flux-GO // Fast gateway for microservice: dubbo, grpc, http"
	VersionFormat = "Version // git.commit=%s, build.version=%s, build.date=%s"
)

const (
	DefaultHttpHeaderVersion = "X-Version"

	ListenerIdDefault   = "default"
	ListenServerIdAdmin = "admin"
)

type (
	// Option 配置HttpServeEngine函数
	Option func(bs *BootstrapServer)
	// VersionLookupFunc Http请求版本查找函数
	VersionLookupFunc func(webex flux.ServerWebContext) (version string)
)

// BootstrapServer
type BootstrapServer struct {
	listener    map[string]flux.WebListener
	hookFunc    []flux.ContextHookFunc
	versionFunc VersionLookupFunc
	dispatcher  *Dispatcher
	started     chan struct{}
	stopped     chan struct{}
	banner      string
}

// WithContextHooks 配置请求Hook函数列表
func WithContextHooks(hooks ...flux.ContextHookFunc) Option {
	return func(bs *BootstrapServer) {
		bs.hookFunc = append(bs.hookFunc, hooks...)
	}
}

// WithVersionLookupFunc 配置Web请求版本选择函数
func WithVersionLookupFunc(fun VersionLookupFunc) Option {
	return func(bs *BootstrapServer) {
		bs.versionFunc = fun
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
		bs.dispatcher.hooks = append(bs.dispatcher.hooks, hooks...)
	}
}

func WithWebListener(server flux.WebListener) Option {
	return func(bs *BootstrapServer) {
		bs.AddWebListener(server.ListenerId(), server)
	}
}

func NewDefaultBootstrapServer(options ...Option) *BootstrapServer {
	opts := []Option{
		WithServerBanner(defaultBanner),
		// Lookup version
		WithVersionLookupFunc(func(webex flux.ServerWebContext) string {
			return webex.HeaderVar(DefaultHttpHeaderVersion)
		}),
		// Default WebListener
		WithWebListener(listener.New(ListenerIdDefault, LoadWebListenerConfig(ListenerIdDefault), nil)),
		// Admin WebListener
		WithWebListener(listener.New(ListenServerIdAdmin, LoadWebListenerConfig(ListenServerIdAdmin), nil,
			// 内部元数据查询
			listener.WithWebHandlers([]listener.WebHandlerTuple{
				{Method: "GET", Pattern: "/inspect/endpoints", Handler: fluxinspect.EndpointsHandler},
				{Method: "GET", Pattern: "/inspect/services", Handler: fluxinspect.ServicesHandler},
				{Method: "GET", Pattern: "/inspect/metrics", Handler: flux.WrapHttpHandler(promhttp.Handler())},
			}),
		)),
	}
	return NewBootstrapServerWith(append(opts, options...)...)
}

func NewBootstrapServerWith(opts ...Option) *BootstrapServer {
	srv := &BootstrapServer{
		dispatcher: NewDispatcher(),
		listener:   make(map[string]flux.WebListener, 2),
		hookFunc:   make([]flux.ContextHookFunc, 0, 4),
		started:    make(chan struct{}),
		stopped:    make(chan struct{}),
		banner:     defaultBanner,
	}
	for _, opt := range opts {
		opt(srv)
	}
	return srv
}

// Prepare Call before init and startup
func (s *BootstrapServer) Prepare() error {
	return s.dispatcher.Prepare()
}

// Initial
func (s *BootstrapServer) Initial() error {
	// Listen Server
	for id, webListener := range s.listener {
		if err := webListener.Init(LoadWebListenerConfig(id)); nil != err {
			return err
		}
	}
	// Discovery
	for _, dis := range ext.EndpointDiscoveries() {
		if err := s.dispatcher.AddInitHook(dis, LoadEndpointDiscoveryConfig(dis.Id())); nil != err {
			return err
		}
	}
	return s.dispatcher.Initial()
}

func (s *BootstrapServer) Startup(build flux.Build) error {
	logger.Infof(VersionFormat, build.CommitId, build.Version, build.Date)
	if "" != s.banner {
		logger.Info(s.banner)
	}
	return s.start()
}

func (s *BootstrapServer) start() error {
	dl := s.defaultListener()
	fluxpkg.Assert(nil != dl, "<default listener> is required")
	// Dispatcher
	logger.Info("SERVER:START:DISPATCHER:START")
	if err := s.dispatcher.Startup(); nil != err {
		return err
	}
	logger.Info("SERVER:START:DISPATCHER:OK")
	// Discovery
	endpoints := make(chan flux.EndpointEvent, 2)
	services := make(chan flux.ServiceEvent, 2)
	defer func() {
		close(endpoints)
		close(services)
	}()
	logger.Info("SERVER:START:DISCOVERY:START")
	ctx, canceled := context.WithCancel(context.Background())
	defer canceled()
	go s.startEventLoop(ctx, endpoints, services)
	if err := s.startEventWatch(ctx, endpoints, services); nil != err {
		return err
	}
	logger.Info("SERVER:START:DISCOVERY:OK")
	// Listeners
	var errch chan error
	for lid, wl := range s.listener {
		logger.Infow("SERVER:START:LISTENER:START", "listener-id", wl.ListenerId())
		go func(id string, server flux.WebListener) {
			errch <- server.Listen()
			logger.Infow("SERVER:START:LISTENER:STOP", "listener-id", id)
		}(lid, wl)
	}
	close(s.started)
	return <-errch
}

func (s *BootstrapServer) startEventLoop(ctx context.Context, endpoints chan flux.EndpointEvent, services chan flux.ServiceEvent) {
	logger.Info("SERVER:START:DISCOVERY:EVENT_LOOP:START")
	defer logger.Info("SERVER:START:DISCOVERY:EVENT_LOOP:STOP")
	for {
		select {
		case epEvt, ok := <-endpoints:
			if ok {
				s.onEndpointEvent(epEvt)
			}

		case esEvt, ok := <-services:
			if ok {
				s.onServiceEvent(esEvt)
			}

		case <-ctx.Done():
			return
		}
	}
}

func (s *BootstrapServer) startEventWatch(ctx context.Context, endpoints chan flux.EndpointEvent, services chan flux.ServiceEvent) error {
	for _, discovery := range ext.EndpointDiscoveries() {
		logger.Infow("SERVER:START:DISCOVERY:WATCH", "discovery-id", discovery.Id())
		if err := discovery.WatchEndpoints(ctx, endpoints); nil != err {
			return err
		}
		if err := discovery.WatchServices(ctx, services); nil != err {
			return err
		}
		logger.Infow("SERVER:START:DISCOVERY:WATCH/OK", "discovery-id", discovery.Id())
	}
	return nil
}

func (s *BootstrapServer) route(webex flux.ServerWebContext, server flux.WebListener, endpoints *flux.MVCEndpoint) (err error) {
	defer func(id string) {
		if rvr := recover(); rvr != nil {
			err = fmt.Errorf("SERVER:ROUTE:CRITICAL_PANIC:%w", rvr)
		}
	}(webex.RequestId())
	endpoint, found := endpoints.Lookup(s.versionFunc(webex))
	// 实现动态Endpoint版本选择
	for _, selector := range ext.EndpointSelectors() {
		if selector.Active(webex, server.ListenerId()) {
			endpoint, found = selector.DoSelect(webex, server.ListenerId(), endpoints)
			if found {
				break
			}
		}
	}
	if !found {
		logger.Trace(webex.RequestId()).Infow("SERVER:ROUTE:NOT_FOUND",
			"http-pattern", []string{webex.Method(), webex.URI(), webex.URL().Path},
		)
		// Endpoint节点版本被删除，需要重新路由到NotFound处理函数
		return server.HandleNotfound(webex)
	} else {
		fluxpkg.Assert(endpoint.IsValid(), "<endpoint> must valid when routing")
	}
	ctxw := flux.NewContext()
	ctxw.Reset(webex, &endpoint)
	ctxw.SetAttribute(flux.XRequestTime, ctxw.StartAt().Unix())
	ctxw.SetAttribute(flux.XRequestId, webex.RequestId())
	ctxw.SetAttribute(flux.XRequestHost, webex.Host())
	ctxw.SetAttribute(flux.XRequestAgent, "flux.go")
	trace := logger.TraceContext(ctxw)
	trace.Infow("SERVER:ROUTE:START")
	// hook
	for _, hook := range s.hookFunc {
		hook(webex, ctxw)
	}
	defer func(start time.Time) {
		trace.Infow("SERVER:ROUTE:END", "metric", ctxw.Metrics(), "elapses", time.Since(start).String())
	}(ctxw.StartAt())
	// route
	if serr := s.dispatcher.Route(ctxw); nil != serr {
		server.HandleError(webex, serr)
	}
	return nil
}

func (s *BootstrapServer) onServiceEvent(event flux.ServiceEvent) {
	service := event.Service
	initArguments(service.Arguments)
	switch event.EventType {
	case flux.EventTypeAdded:
		logger.Infow("SERVER:EVENT:SERVICE:ADD",
			"service-id", service.ServiceId, "alias-id", service.AliasId)
		ext.RegisterTransporterService(service)
		if "" != service.AliasId {
			ext.RegisterTransporterServiceById(service.AliasId, service)
		}
	case flux.EventTypeUpdated:
		logger.Infow("SERVER:EVENT:SERVICE:UPDATE",
			"service-id", service.ServiceId, "alias-id", service.AliasId)
		ext.RegisterTransporterService(service)
		if "" != service.AliasId {
			ext.RegisterTransporterServiceById(service.AliasId, service)
		}
	case flux.EventTypeRemoved:
		logger.Infow("SERVER:EVENT:SERVICE:REMOVE",
			"service-id", service.ServiceId, "alias-id", service.AliasId)
		ext.RemoveTransporterService(service.ServiceId)
		if "" != service.AliasId {
			ext.RemoveTransporterService(service.AliasId)
		}
	}
}

func (s *BootstrapServer) onEndpointEvent(event flux.EndpointEvent) {
	method := strings.ToUpper(event.Endpoint.HttpMethod)
	// Check http method
	if !isAllowedHttpMethod(method) {
		logger.Warnw("SERVER:EVENT:ENDPOINT:METHOD/IGNORE", "method", method, "pattern", event.Endpoint.HttpPattern)
		return
	}
	pattern := event.Endpoint.HttpPattern
	routeKey := fmt.Sprintf("%s#%s", method, pattern)
	endpoint := event.Endpoint
	initArguments(endpoint.Service.Arguments)
	initArguments(endpoint.Permission.Arguments)
	bind, isreg := s.selectMultiEndpoint(routeKey, &endpoint)
	switch event.EventType {
	case flux.EventTypeAdded:
		logger.Infow("SERVER:EVENT:ENDPOINT:ADD", "version", endpoint.Version, "method", method, "pattern", pattern)
		bind.Update(endpoint.Version, &endpoint)
		// 根据Endpoint属性，选择ListenServer来绑定
		if isreg {
			id := endpoint.GetAttr(flux.EndpointAttrTagListenerId).GetString()
			if id == "" {
				id = ListenerIdDefault
			}
			server, ok := s.WebListenerById(id)
			if ok {
				logger.Infow("SERVER:EVENT:ENDPOINT:HTTP_HANDLER/"+id, "method", method, "pattern", pattern)
				server.AddHandler(method, pattern, s.newEndpointHandler(server, bind))
			} else {
				logger.Errorw("SERVER:EVENT:ENDPOINT:LISTENER_MISSED/"+id, "method", method, "pattern", pattern)
			}
		}
	case flux.EventTypeUpdated:
		logger.Infow("SERVER:EVENT:ENDPOINT:UPDATE", "version", endpoint.Version, "method", method, "pattern", pattern)
		bind.Update(endpoint.Version, &endpoint)
	case flux.EventTypeRemoved:
		logger.Infow("SERVER:EVENT:ENDPOINT:REMOVE", "method", method, "pattern", pattern)
		bind.Delete(endpoint.Version)
	}
}

// Shutdown to cleanup resources
func (s *BootstrapServer) Shutdown(ctx goctx.Context) error {
	logger.Info("Server shutdown...")
	defer close(s.stopped)
	for id, server := range s.listener {
		if err := server.Close(ctx); nil != err {
			logger.Warnw("Server["+id+"] shutdown http server", "error", err)
		}
	}
	return s.dispatcher.Shutdown(ctx)
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
	s.defaultListener().AddInterceptor(m)
}

// AddWebHandler 添加Http处理接口到默认ListenerServer。
func (s *BootstrapServer) AddWebHandler(method, pattern string, h flux.WebHandler, m ...flux.WebInterceptor) {
	s.defaultListener().AddHandler(method, pattern, h, m...)
}

// AddWebHttpHandler 添加Http处理接口到默认ListenerServer。
func (s *BootstrapServer) AddWebHttpHandler(method, pattern string, h http.Handler, m ...func(http.Handler) http.Handler) {
	s.defaultListener().AddHttpHandler(method, pattern, h, m...)
}

// SetWebNotfoundHandler 设置Http路由失败的处理接口到默认ListenerServer
func (s *BootstrapServer) SetWebNotfoundHandler(nfh flux.WebHandler) {
	s.defaultListener().SetNotfoundHandler(nfh)
}

// AddWebListener 添加指定ID
func (s *BootstrapServer) AddWebListener(id string, server flux.WebListener) {
	s.listener[id] = fluxpkg.MustNotNil(server, "WebListener is nil").(flux.WebListener)
}

// WebListenerById 返回ListenServer实例
func (s *BootstrapServer) WebListenerById(listenerId string) (flux.WebListener, bool) {
	ls, ok := s.listener[listenerId]
	return ls, ok
}

// AddContextHookFunc 添加Http与Flux的Context桥接函数
func (s *BootstrapServer) AddContextHookFunc(f flux.ContextHookFunc) {
	s.hookFunc = append(s.hookFunc, f)
}

func (s *BootstrapServer) newEndpointHandler(server flux.WebListener, endpoint *flux.MVCEndpoint) flux.WebHandler {
	return func(webex flux.ServerWebContext) error {
		return s.route(webex, server, endpoint)
	}
}

func (s *BootstrapServer) selectMultiEndpoint(routeKey string, endpoint *flux.Endpoint) (*flux.MVCEndpoint, bool) {
	if mve, ok := ext.EndpointByKey(routeKey); ok {
		return mve, false
	} else {
		return ext.RegisterEndpoint(routeKey, endpoint), true
	}
}

func (s *BootstrapServer) defaultListener() flux.WebListener {
	count := len(s.listener)
	if count == 0 {
		return nil
	} else if count == 1 {
		for _, server := range s.listener {
			return server
		}
	}
	server, ok := s.listener[ListenerIdDefault]
	if ok {
		return server
	}
	return nil
}

func LoadWebListenerConfig(id string) *flux.Configuration {
	return flux.NewConfigurationOfNS(flux.NamespaceWebListeners + "." + id)
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
		args[i].ValueResolver = ext.MTValueResolverByType(args[i].Class)
		args[i].LookupFunc = ext.ArgumentLookupFunc()
		initArguments(args[i].Fields)
	}
}
