package server

import (
	goctx "context"
	"fmt"
	dubgo "github.com/apache/dubbo-go/config"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/toolkit"
	"golang.org/x/net/context"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"sync"
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
	// GenericOptionFunc 配置HttpServeEngine函数
	GenericOptionFunc func(gs *GenericServer)
	// VersionLookupFunc Http请求版本查找函数
	VersionLookupFunc func(webex flux.ServerWebContext) (version string)
)

// GenericServer
type GenericServer struct {
	listener     map[string]flux.WebListener
	contextHooks []flux.ContextHookFunc
	prepareHooks []flux.PrepareHookFunc
	versionFunc  VersionLookupFunc
	dispatcher   *Dispatcher
	pooled       *sync.Pool
	started      chan struct{}
	stopped      chan struct{}
	banner       string
}

// WithContextHooks 配置请求Hook函数列表
func WithContextHooks(hooks ...flux.ContextHookFunc) GenericOptionFunc {
	return func(bs *GenericServer) {
		bs.contextHooks = append(bs.contextHooks, hooks...)
	}
}

// WithPrepareHooks 配置服务启动预备阶段Hook函数列表
func WithPrepareHooks(hooks ...flux.PrepareHookFunc) GenericOptionFunc {
	return func(bs *GenericServer) {
		bs.prepareHooks = append(bs.prepareHooks, hooks...)
	}
}

// WithVersionLookupFunc 配置Web请求版本选择函数
func WithVersionLookupFunc(fun VersionLookupFunc) GenericOptionFunc {
	return func(bs *GenericServer) {
		bs.versionFunc = fun
	}
}

// WithBanner 配置服务Banner
func WithServerBanner(banner string) GenericOptionFunc {
	return func(bs *GenericServer) {
		bs.banner = banner
	}
}

func WithWebListener(server flux.WebListener) GenericOptionFunc {
	return func(bs *GenericServer) {
		bs.AddWebListener(server.ListenerId(), server)
	}
}

func NewGenericServer(opts ...GenericOptionFunc) *GenericServer {
	server := &GenericServer{
		dispatcher:   NewDispatcher(),
		listener:     make(map[string]flux.WebListener, 2),
		contextHooks: make([]flux.ContextHookFunc, 0, 4),
		prepareHooks: make([]flux.PrepareHookFunc, 0, 4),
		pooled:       &sync.Pool{New: func() interface{} { return flux.NewContext() }},
		started:      make(chan struct{}),
		stopped:      make(chan struct{}),
		banner:       defaultBanner,
	}
	for _, opt := range opts {
		opt(server)
	}
	return server
}

// Prepare Call before init and startup
func (gs *GenericServer) Prepare() error {
	logger.Info("SERVER:EVEN:PREPARE")
	for _, hook := range append(ext.PrepareHooks(), gs.prepareHooks...) {
		if err := hook(); nil != err {
			return err
		}
	}
	logger.Info("SERVER:EVEN:PREPARE:OK")
	return nil
}

// Initial
func (gs *GenericServer) Initial() error {
	logger.Info("SERVER:EVEN:INIT")
	defer logger.Info("SERVER:EVEN:INIT:OK")
	// Listen Server
	for id, webListener := range gs.listener {
		config := NewWebListenerConfig(id)
		if err := webListener.Init(config); nil != err {
			return err
		}
	}
	// Discovery
	for _, dis := range ext.EndpointDiscoveries() {
		config := flux.NewConfigurationByKeys(flux.NamespaceDiscoveries, dis.Id())
		err := gs.dispatcher.AddInitHook(dis, config)
		if nil != err {
			return err
		}
	}
	return gs.dispatcher.Init()
}

func (gs *GenericServer) Startup(build flux.Build) error {
	logger.Infof(VersionFormat, build.CommitId, build.Version, build.Date)
	if gs.banner != "" {
		logger.Info(gs.banner)
	}
	return gs.start()
}

func (gs *GenericServer) start() error {
	toolkit.AssertNotNil(gs.defaultListener(), "<default-listener> MUST NOT nil")
	// Dispatcher
	logger.Info("SERVER:EVEN:STARTUP")
	if err := gs.dispatcher.Startup(); nil != err {
		return err
	}
	logger.Info("SERVER:EVEN:STARTUP:OK")
	// Discovery
	endpoints := make(chan flux.EndpointEvent, 2)
	services := make(chan flux.ServiceEvent, 2)
	defer func() {
		close(endpoints)
		close(services)
	}()
	logger.Info("SERVER:EVEN:DISCOVERY:START")
	ctx, canceled := context.WithCancel(context.Background())
	defer canceled()
	go gs.startEventLoop(ctx, endpoints, services)
	if err := gs.startEventWatch(ctx, endpoints, services); nil != err {
		return err
	}
	logger.Info("SERVER:EVEN:DISCOVERY:OK")
	// Listeners
	var errch chan error
	for id, web := range gs.listener {
		logger.Infow("SERVER:EVEN:LISTENER:START", "listener-id", web.ListenerId())
		go func(id string, server flux.WebListener) {
			errch <- server.Listen()
			logger.Infow("SERVER:EVEN:LISTENER:STOP", "listener-id", id)
		}(id, web)
	}
	close(gs.started)
	return <-errch
}

func (gs *GenericServer) startEventLoop(ctx context.Context, endpoints chan flux.EndpointEvent, services chan flux.ServiceEvent) {
	logger.Info("SERVER:EVEN:EVENTLOOP:START")
	defer logger.Info("SERVER:EVEN:EVENTLOOP:STOP")
	for {
		select {
		case epEvt, ok := <-endpoints:
			if ok {
				gs.onEndpointEvent(epEvt)
			}

		case esEvt, ok := <-services:
			if ok {
				gs.onServiceEvent(esEvt)
			}

		case <-ctx.Done():
			return
		}
	}
}

func (gs *GenericServer) startEventWatch(ctx context.Context, endpoints chan flux.EndpointEvent, services chan flux.ServiceEvent) error {
	for _, discovery := range ext.EndpointDiscoveries() {
		logger.Infow("SERVER:EVEN:DISCOVERY:WATCH", "discovery-id", discovery.Id())
		if err := discovery.WatchEndpoints(ctx, endpoints); nil != err {
			return err
		}
		if err := discovery.WatchServices(ctx, services); nil != err {
			return err
		}
		logger.Infow("SERVER:EVEN:DISCOVERY:WATCH/OK", "discovery-id", discovery.Id())
	}
	return nil
}

func (gs *GenericServer) route(webex flux.ServerWebContext, server flux.WebListener, endpoints *flux.MVCEndpoint) (err error) {
	defer func(id string) {
		if rvr := recover(); rvr != nil {
			logger.Trace(id).Errorw("SERVER:EVEN:ROUTE:CRITICAL_PANIC", "error", rvr, "debug", string(debug.Stack()))
			err = fmt.Errorf("SERVER:EVEN:ROUTE:%gs", rvr)
		}
	}(webex.RequestId())
	endpoint, found := endpoints.Lookup(gs.versionFunc(webex))
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
		logger.Trace(webex.RequestId()).Infow("SERVER:EVEN:ROUTE:NOT_FOUND",
			"http-pattern", []string{webex.Method(), webex.URI(), webex.URL().Path},
		)
		// Endpoint节点版本被删除，需要重新路由到NotFound处理函数
		return server.HandleNotfound(webex)
	} else {
		toolkit.Assert(endpoint.IsValid(), "<endpoint> must valid when routing")
	}
	ctxw := gs.pooled.Get().(*flux.Context)
	defer gs.pooled.Put(ctxw)
	ctxw.Reset(webex, &endpoint)
	ctxw.SetAttribute(flux.XRequestTime, ctxw.StartAt().Unix())
	ctxw.SetAttribute(flux.XRequestId, webex.RequestId())
	trace := logger.TraceContext(ctxw)
	trace.Infow("SERVER:EVEN:ROUTE:START")
	// hook
	for _, hook := range gs.contextHooks {
		hook(webex, ctxw)
	}
	defer func(start time.Time) {
		trace.Infow("SERVER:EVEN:ROUTE:END", "metric", ctxw.Metrics(), "elapses", time.Since(start).String())
	}(ctxw.StartAt())
	// route
	if serr := gs.dispatcher.Route(ctxw); nil != serr {
		server.HandleError(webex, serr)
	}
	return nil
}

func (gs *GenericServer) onServiceEvent(event flux.ServiceEvent) {
	service := event.Service
	switch event.EventType {
	case flux.EventTypeAdded:
		logger.Infow("SERVER:EVENT:SERVICE:ADD",
			"service-id", service.ServiceId, "alias-id", service.AliasId)
		gs.bindArguments(service.Arguments)
		ext.RegisterService(service)
		if service.AliasId != "" {
			ext.RegisterServiceByID(service.AliasId, service)
		}
	case flux.EventTypeUpdated:
		logger.Infow("SERVER:EVENT:SERVICE:UPDATE",
			"service-id", service.ServiceId, "alias-id", service.AliasId)
		gs.bindArguments(service.Arguments)
		ext.RegisterService(service)
		if service.AliasId != "" {
			ext.RegisterServiceByID(service.AliasId, service)
		}
	case flux.EventTypeRemoved:
		logger.Infow("SERVER:EVENT:SERVICE:REMOVE",
			"service-id", service.ServiceId, "alias-id", service.AliasId)
		ext.RemoveServiceByID(service.ServiceId)
		if service.AliasId != "" {
			ext.RemoveServiceByID(service.AliasId)
		}
	}
}

func (gs *GenericServer) onEndpointEvent(event flux.EndpointEvent) {
	method := strings.ToUpper(event.Endpoint.HttpMethod)
	// Check http method
	if !isAllowMethod(method) {
		logger.Warnw("SERVER:EVENT:ENDPOINT:METHOD/IGNORE", "method", method, "pattern", event.Endpoint.HttpPattern)
		return
	}
	endpoint := event.Endpoint
	pattern := event.Endpoint.HttpPattern
	mvce, register := gs.selectMVCEndpoint(&endpoint)
	switch event.EventType {
	case flux.EventTypeAdded:
		logger.Infow("SERVER:EVENT:ENDPOINT:ADD", "version", endpoint.Version, "method", method, "pattern", pattern)
		gs.bindArguments(append(endpoint.Service.Arguments, endpoint.PermissionService.Arguments...))
		mvce.Update(endpoint.Version, &endpoint)
		// 根据Endpoint属性，选择ListenServer来绑定
		if register {
			id := endpoint.Attributes.Single(flux.EndpointAttrTagListenerId).ToString()
			if id == "" {
				id = ListenerIdDefault
			}
			server, ok := gs.WebListenerById(id)
			if ok {
				logger.Infow("SERVER:EVENT:ENDPOINT:HTTP_HANDLER/"+id, "method", method, "pattern", pattern)
				server.AddHandler(method, pattern, gs.newEndpointHandler(server, mvce))
			} else {
				logger.Errorw("SERVER:EVENT:ENDPOINT:LISTENER_MISSED/"+id, "method", method, "pattern", pattern)
			}
		}
	case flux.EventTypeUpdated:
		logger.Infow("SERVER:EVENT:ENDPOINT:UPDATE", "version", endpoint.Version, "method", method, "pattern", pattern)
		gs.bindArguments(append(endpoint.Service.Arguments, endpoint.PermissionService.Arguments...))
		mvce.Update(endpoint.Version, &endpoint)
	case flux.EventTypeRemoved:
		logger.Infow("SERVER:EVENT:ENDPOINT:REMOVE", "method", method, "pattern", pattern)
		mvce.Delete(endpoint.Version)
	}
}

// Shutdown to cleanup resources
func (gs *GenericServer) Shutdown(ctx goctx.Context) error {
	logger.Info("SERVER:EVENT:SHUTDOWN")
	defer close(gs.stopped)
	for id, server := range gs.listener {
		if err := server.Close(ctx); nil != err {
			logger.Warnw("Server["+id+"] shutdown http server", "error", err)
		}
	}
	return gs.dispatcher.Shutdown(ctx)
}

// GracefulShutdown
func (gs *GenericServer) OnSignalShutdown(quit chan os.Signal, to time.Duration) {
	// 接收停止信号
	signal.Notify(quit, dubgo.ShutdownSignals...)
	<-quit
	logger.Infof("SERVER:EVENT:SIGNAL:SHUTDOWN")
	ctx, cancel := goctx.WithTimeout(goctx.Background(), to)
	defer cancel()
	if err := gs.Shutdown(ctx); nil != err {
		logger.Errorw("SERVER:EVENT:SHUTDOWN:ERROR", "error", err)
	}
}

// StateStarted 返回一个Channel。当服务启动完成时，此Channel将被关闭。
func (gs *GenericServer) StateStarted() <-chan struct{} {
	return gs.started
}

// StateStopped 返回一个Channel。当服务停止后完成时，此Channel将被关闭。
func (gs *GenericServer) StateStopped() <-chan struct{} {
	return gs.stopped
}

// AddWebInterceptor 添加Http前拦截器到默认ListenerServer。将在Http被路由到对应Handler之前执行
func (gs *GenericServer) AddWebInterceptor(m flux.WebInterceptor) {
	gs.defaultListener().AddInterceptor(m)
}

// AddWebHandler 添加Http处理接口到默认ListenerServer。
func (gs *GenericServer) AddWebHandler(method, pattern string, h flux.WebHandler, m ...flux.WebInterceptor) {
	gs.defaultListener().AddHandler(method, pattern, h, m...)
}

// AddWebHttpHandler 添加Http处理接口到默认ListenerServer。
func (gs *GenericServer) AddWebHttpHandler(method, pattern string, h http.Handler, m ...func(http.Handler) http.Handler) {
	gs.defaultListener().AddHttpHandler(method, pattern, h, m...)
}

// SetWebNotfoundHandler 设置Http路由失败的处理接口到默认ListenerServer
func (gs *GenericServer) SetWebNotfoundHandler(nfh flux.WebHandler) {
	gs.defaultListener().SetNotfoundHandler(nfh)
}

// AddWebListener 添加指定ID
func (gs *GenericServer) AddWebListener(listenerID string, listener flux.WebListener) {
	toolkit.AssertNotNil(listener, "WebListener Must Not nil")
	toolkit.AssertNotEmpty(listenerID, "WebListener Id Must Not empty")
	gs.listener[strings.ToLower(listenerID)] = listener
}

// WebListenerById 返回ListenServer实例
func (gs *GenericServer) WebListenerById(listenerID string) (flux.WebListener, bool) {
	ls, ok := gs.listener[strings.ToLower(listenerID)]
	return ls, ok
}

// AddContextHookFunc 添加Http与Flux的Context桥接函数
func (gs *GenericServer) AddContextHookFunc(f flux.ContextHookFunc) {
	gs.contextHooks = append(gs.contextHooks, f)
}

func (gs *GenericServer) newEndpointHandler(server flux.WebListener, endpoint *flux.MVCEndpoint) flux.WebHandler {
	return func(webex flux.ServerWebContext) error {
		return gs.route(webex, server, endpoint)
	}
}

func (gs *GenericServer) selectMVCEndpoint(endpoint *flux.Endpoint) (*flux.MVCEndpoint, bool) {
	key := ext.MakeEndpointKey(endpoint.HttpMethod, endpoint.HttpPattern)
	if mve, ok := ext.EndpointByKey(key); ok {
		return mve, false
	} else {
		return ext.RegisterEndpoint(key, endpoint), true
	}
}

func (gs *GenericServer) defaultListener() flux.WebListener {
	count := len(gs.listener)
	if count == 0 {
		return nil
	} else if count == 1 {
		for _, server := range gs.listener {
			return server
		}
	}
	server, ok := gs.listener[ListenerIdDefault]
	if ok {
		return server
	}
	return nil
}

func (gs *GenericServer) bindArguments(args []flux.Argument) {
	for i := range args {
		args[i].ValueResolver = ext.MTValueResolverByType(args[i].Class)
		args[i].LookupFunc = ext.LookupFunc()
		gs.bindArguments(args[i].Fields)
	}
}

// NewWebListenerOptions 根据WebListenerID，返回初始化WebListener实例时的配置
func NewWebListenerConfig(id string) *flux.Configuration {
	return flux.NewConfigurationByKeys(flux.NamespaceWebListeners, id)
}

// NewWebListenerOptions 根据WebListenerID，返回构建WebListener实例时的选项参数
func NewWebListenerOptions(id string) *flux.Configuration {
	return flux.NewConfigurationByKeys(flux.NamespaceWebListeners, id)
}

func isAllowMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodPut,
		http.MethodHead, http.MethodOptions, http.MethodPatch, http.MethodTrace:
		return true
	default:
		logger.Errorw("Ignore unsupported http method:", "method", method)
		return false
	}
}
