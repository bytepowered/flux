package server

import (
	goctx "context"
	"fmt"
	dubgo "github.com/apache/dubbo-go/config"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/internal"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/toolkit"
	"github.com/jinzhu/copier"
	"golang.org/x/net/context"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"reflect"
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
	ListenerIdWebapi    = ListenerIdDefault
	ListenServerIdAdmin = "admin"
)

type (
	// GenericOptionFunc 配置HttpServeEngine函数
	GenericOptionFunc func(gs *GenericServer)
	// VersionLookupFunc Http请求版本查找函数
	VersionLookupFunc func(webex flux.ServerWebContext) (version string)
)

// GenericServer Server
type GenericServer struct {
	listeners      map[string]flux.WebListener
	onContextHooks []flux.OnContextHookFunc
	versionFunc    VersionLookupFunc
	dispatcher     *Dispatcher
	pooled         *sync.Pool
	started        chan struct{}
	stopped        chan struct{}
	banner         string
}

// WithOnContextHooks 配置请求Hook函数列表
func WithOnContextHooks(hooks ...flux.OnContextHookFunc) GenericOptionFunc {
	return func(bs *GenericServer) {
		bs.onContextHooks = append(bs.onContextHooks, hooks...)
	}
}

// WithVersionLookupFunc 配置Web请求版本选择函数
func WithVersionLookupFunc(fun VersionLookupFunc) GenericOptionFunc {
	return func(bs *GenericServer) {
		bs.versionFunc = fun
	}
}

// WithServerBanner 配置服务Banner
func WithServerBanner(banner string) GenericOptionFunc {
	return func(bs *GenericServer) {
		bs.banner = banner
	}
}

// WithNewWebListener 配置WebListener
func WithNewWebListener(server flux.WebListener) GenericOptionFunc {
	return func(bs *GenericServer) {
		bs.AddWebListener(server.ListenerId(), server)
	}
}

// WithServeResponseWriter 配置ResponseWriter
func WithServeResponseWriter(writer flux.ServeResponseWriter) GenericOptionFunc {
	return func(bs *GenericServer) {
		bs.dispatcher.setResponseWriter(writer)
	}
}

func WithOnBeforeFilterHookFunc(hooks ...flux.OnBeforeFilterHookFunc) GenericOptionFunc {
	return func(bs *GenericServer) {
		for _, h := range hooks {
			bs.dispatcher.addOnBeforeFilterHook(h)
		}
	}
}

func WithOnBeforeTransportHookFunc(hooks ...flux.OnBeforeTransportHookFunc) GenericOptionFunc {
	return func(bs *GenericServer) {
		for _, h := range hooks {
			bs.dispatcher.addOnBeforeTransportHook(h)
		}
	}
}

func NewGenericServer(opts ...GenericOptionFunc) *GenericServer {
	server := &GenericServer{
		dispatcher:     NewDispatcher(),
		listeners:      make(map[string]flux.WebListener, 2),
		onContextHooks: make([]flux.OnContextHookFunc, 0, 4),
		pooled:         &sync.Pool{New: func() interface{} { return flux.NewContext() }},
		started:        make(chan struct{}),
		stopped:        make(chan struct{}),
		banner:         defaultBanner,
	}
	for _, opt := range opts {
		opt(server)
	}
	return server
}

// Prepare Call before init and startup
func (gs *GenericServer) Prepare() error {
	logger.Info("SERVER:EVEN:PREPARE")
	for _, hook := range ext.PrepareHooks() {
		if err := hook.OnPrepare(); nil != err {
			return err
		}
	}
	logger.Info("SERVER:EVEN:PREPARE:OK")
	return nil
}

// Init Call components init
func (gs *GenericServer) Init() error {
	logger.Info("SERVER:EVEN:INIT")
	defer logger.Info("SERVER:EVEN:INIT:OK")
	// 1. WebListen Server
	for id, webListener := range gs.listeners {
		config := NewWebListenerConfig(id)
		ext.AddStartupHook(webListener)
		ext.AddShutdownHook(webListener)
		logger.Infow("SERVER:EVENT:INIT:WEBLISTENER", "webl-id", id, "webl-type", reflect.TypeOf(webListener))
		if err := webListener.OnInit(config); nil != err {
			return err
		}
	}
	// 2. EDS
	for _, eds := range ext.EndpointDiscoveries() {
		ext.AddStartupHook(eds)
		ext.AddShutdownHook(eds)
		err := onInitializer(eds, func(initable flux.Initializer) error {
			logger.Infow("SERVER:EVENT:INIT:DISCOVERY", "eds-id", eds.Id(), "eds-type", reflect.TypeOf(eds))
			edsc := flux.NewConfigurationByKeys(flux.NamespaceDiscoveries, eds.Id())
			return initable.OnInit(edsc)
		})
		if nil != err {
			return err
		}
	}
	// 3. Transporter
	for proto, transporter := range ext.Transporters() {
		ext.AddStartupHook(transporter)
		ext.AddShutdownHook(transporter)
		err := onInitializer(transporter, func(initable flux.Initializer) error {
			trc := flux.NewConfigurationByKeys(flux.NamespaceTransporters, proto)
			logger.Infow("SERVER:EVENT:INIT:TRANSPORT", "t-proto", proto, "t-type", reflect.TypeOf(transporter))
			return initable.OnInit(trc)
		})
		if nil != err {
			return err
		}
	}
	// 4. Static Filters
	for _, filter := range append(ext.GlobalFilters(), ext.SelectiveFilters()...) {
		err := onInitializer(filter, func(initable flux.Initializer) error {
			fic := flux.NewConfiguration(filter.FilterId())
			if IsDisabled(fic) {
				logger.Infow("SERVER:EVENT:INIT:FILTER/disabled", "f-id", filter.FilterId())
				return nil
			}
			ext.AddStartupHook(filter)
			ext.AddShutdownHook(filter)
			logger.Infow("SERVER:EVENT:INIT:FILTER", "f-id", filter.FilterId(), "f-type", reflect.TypeOf(filter))
			return initable.OnInit(fic)
		})
		if nil != err {
			return err
		}
	}
	// 5. Dynamic Filters
	dynFilters, err := dynamicFilters()
	if nil != err {
		return err
	}
	for _, dynf := range dynFilters {
		inst := dynf.Factory()
		dynfilter, ok := inst.(flux.Filter)
		var msgs = []interface{}{"dynf-id", dynf.Id, "dynf-type-id", dynf.TypeId, "dynf-type", reflect.TypeOf(inst)}
		if !ok {
			logger.Infow("SERVER:EVENT:INIT:DYNFILTER/malformed", msgs...)
			continue
		}
		logger.Infow("SERVER:EVENT:INIT:DYNFILTER/new", msgs)
		if IsDisabled(dynf.Config) {
			logger.Infow("SERVER:EVENT:INIT:DYNFILTER/disable", "dynf-id", dynf.Id)
			continue
		}
		ext.AddSelectiveFilter(dynfilter)
		ext.AddStartupHook(dynfilter)
		ext.AddShutdownHook(dynfilter)
		err := onInitializer(dynfilter, func(initable flux.Initializer) error {
			logger.Infow("SERVER:EVENT:INIT:FILTER", "dynf-id", dynfilter.FilterId())
			return initable.OnInit(dynf.Config)
		})
		if nil != err {
			return err
		}
	}
	return nil
}

func (gs *GenericServer) Startup(build flux.Build) error {
	fmt.Printf(VersionFormat, build.CommitId, build.Version, build.Date)
	if gs.banner != "" {
		fmt.Println(gs.banner)
	}
	return gs.start()
}

func (gs *GenericServer) start() error {
	flux.AssertNotNil(gs.defaultListener(), "<default-listener> MUST NOT nil")
	flux.AssertNotNil(ext.LookupScopedValueFunc(), "<scope-value-lookup-func> MUST NOT nil")
	logger.Info("SERVER:EVEN:STARTUP")
	for _, startup := range sortedStartup(ext.StartupHooks()) {
		if err := startup.OnStartup(); nil != err {
			return err
		}
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
	errch := make(chan error, 1)
	for id, web := range gs.listeners {
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
		if err := discovery.WatchServices(ctx, services); nil != err {
			return err
		}
		if err := discovery.WatchEndpoints(ctx, endpoints); nil != err {
			return err
		}
		logger.Infow("SERVER:EVEN:DISCOVERY:WATCH/OK", "discovery-id", discovery.Id())
	}
	return nil
}

func (gs *GenericServer) route(webex flux.ServerWebContext, server flux.WebListener, endpoints *flux.MVCEndpoint) (err error) {
	defer func(id string) {
		if panerr := recover(); panerr != nil {
			trace := logger.Trace(id)
			if recerr, ok := panerr.(error); ok {
				trace.Errorw(recerr.Error(), "r-error", recerr, "debug", string(debug.Stack()))
				err = recerr
			} else {
				trace.Errorw("SERVER:EVEN:ROUTE:CRITICAL_PANIC", "r-error", panerr, "debug", string(debug.Stack()))
				err = fmt.Errorf("SERVER:EVEN:ROUTE:%s", panerr)
			}
		}
	}(webex.RequestId())
	var endpoint flux.EndpointSpec
	// 查找匹配版本的Endpoint
	if src, found := gs.lookup(webex, server, endpoints); found {
		// dup to enforce metadata safe
		cperr := gs.dup(&endpoint, src)
		flux.AssertM(cperr == nil, func() string {
			return fmt.Sprintf("duplicate endpoint metadata, error: %s", cperr.Error())
		})
	} else {
		logger.Trace(webex.RequestId()).Infow("SERVER:EVEN:ROUTE:ENDPOINT/NOT_FOUND",
			"http-pattern", []string{webex.Method(), webex.URI(), webex.URL().Path},
		)
		// Endpoint节点版本被删除，需要重新路由到NotFound处理函数
		return server.HandleNotfound(webex)
	}
	// 检查Endpoint/Service绑定
	flux.AssertTrue(endpoint.Valid(), "<endpoint> must valid when routing")
	flux.AssertTrue(endpoint.Service.IsValid(), "<endpoint.service> must valid when routing")
	ctxw := gs.pooled.Get().(*flux.Context)
	defer gs.pooled.Put(ctxw)
	ctxw.Reset(webex, &endpoint)
	ctxw.SetAttribute(flux.XRequestTime, ctxw.StartAt().Unix())
	ctxw.SetAttribute(flux.XRequestId, webex.RequestId())
	logger.TraceContext(ctxw).Infow("SERVER:EVEN:ROUTE:START")
	// hook
	for _, hook := range gs.onContextHooks {
		hook(webex, ctxw)
	}
	defer func(start time.Time) {
		logger.Trace(webex.RequestId()).Infow("SERVER:EVEN:ROUTE:END", "metric", ctxw.Metrics(), "elapses", time.Since(start).String())
	}(ctxw.StartAt())
	// route
	if rouerr := gs.dispatcher.dispatch(ctxw); nil != rouerr {
		server.HandleError(webex, rouerr)
	}
	return
}

func (gs *GenericServer) lookup(webex flux.ServerWebContext, server flux.WebListener, endpoints *flux.MVCEndpoint) (*flux.EndpointSpec, bool) {
	// 动态Endpoint版本选择
	for _, selector := range ext.EndpointSelectors() {
		if selector.Active(webex, server.ListenerId()) {
			if ep, ok := selector.DoSelect(webex, server.ListenerId(), endpoints); ok {
				return ep, true
			}
		}
	}
	// 默认版本选择
	return endpoints.Lookup(gs.versionFunc(webex))
}

// Shutdown to cleanup resources
func (gs *GenericServer) Shutdown(ctx goctx.Context) error {
	logger.Info("SERVER:EVENT:SHUTDOWN")
	defer func() {
		logger.Info("SERVER:EVENT:SHUTDOWN/ok")
		close(gs.stopped)
	}()
	// WebListener
	for id, server := range gs.listeners {
		if err := server.OnShutdown(ctx); nil != err {
			logger.Warnw("Server["+id+"] shutdown http server", "error", err)
		}
	}
	// Components
	for _, shutdown := range sortedShutdown(ext.ShutdownHooks()) {
		if err := shutdown.OnShutdown(ctx); nil != err {
			logger.Warnw("Component shutdown failed", "c-type", reflect.TypeOf(shutdown), "error", err)
		}
	}
	return nil
}

func (gs *GenericServer) onServiceEvent(event flux.ServiceEvent) {
	service := event.Service
	var epvars = []interface{}{"service-id", service.ServiceID(), "alias-id", service.AliasId}
	if err := internal.VerifyAnnotations(service.Annotations); err != nil {
		logger.Warnw("SERVER:EVENT:SERVICE:ANNOTATION/invalid", epvars...)
		return
	}
	switch event.EventType {
	case flux.EventTypeAdded:
		logger.Infow("SERVER:EVENT:SERVICE:ADD", epvars...)
		gs.syncEndpoint(&service)
		ext.RegisterService(service)
		if service.AliasId != "" {
			ext.RegisterServiceByID(service.AliasId, service)
		}

	case flux.EventTypeUpdated:
		logger.Infow("SERVER:EVENT:SERVICE:UPDATE", epvars...)
		gs.syncEndpoint(&service)
		ext.RegisterService(service)
		if service.AliasId != "" {
			ext.RegisterServiceByID(service.AliasId, service)
		}
	case flux.EventTypeRemoved:
		logger.Infow("SERVER:EVENT:SERVICE:REMOVE", epvars...)
		ext.RemoveServiceByID(service.ServiceID())
		if service.AliasId != "" {
			ext.RemoveServiceByID(service.AliasId)
		}
	}
}

func (gs *GenericServer) onEndpointEvent(event flux.EndpointEvent) {
	ep := event.Endpoint
	var epvars = []interface{}{"ep-app", ep.Application, "ep-version", ep.Version, "ep-method", ep.HttpMethod, "ep-pattern", ep.HttpPattern}
	// Check http method
	method := strings.ToUpper(event.Endpoint.HttpMethod)
	if !SupportedHttpMethod(method) {
		logger.Warnw("SERVER:EVENT:ENDPOINT:METHOD/IGNORE", epvars...)
		return
	}
	if err := internal.VerifyAnnotations(ep.Annotations); err != nil {
		logger.Warnw("SERVER:EVENT:ENDPOINT:ANNOTATION/invalid", epvars...)
		return
	}
	mvce, register := gs.selectMVCEndpoint(&ep)
	switch event.EventType {
	case flux.EventTypeAdded:
		logger.Infow("SERVER:EVENT:ENDPOINT:ADD", epvars...)
		gs.syncService(&ep)
		mvce.Update(ep.Version, &ep)
		if register {
			// 根据Endpoint注解属性，选择ListenServer来绑定
			var listenerId = ListenerIdDefault
			if anno, ok := ep.AnnotationEx(flux.EndpointAnnotationListenerSel); ok && anno.IsValid() {
				listenerId = anno.GetString()
			}
			if webListener, ok := gs.WebListenerById(listenerId); ok {
				logger.Infow("SERVER:EVENT:ENDPOINT:HTTP_HANDLER/"+listenerId, epvars...)
				webListener.AddHandler(ep.HttpMethod, ep.HttpPattern, gs.newEndpointHandler(webListener, mvce))
			} else {
				logger.Errorw("SERVER:EVENT:ENDPOINT:LISTENER_MISSED/"+listenerId, epvars...)
			}
		}
	case flux.EventTypeUpdated:
		logger.Infow("SERVER:EVENT:ENDPOINT:UPDATE", epvars...)
		gs.syncService(&ep)
		mvce.Update(ep.Version, &ep)
	case flux.EventTypeRemoved:
		logger.Infow("SERVER:EVENT:ENDPOINT:REMOVE", epvars...)
		mvce.Delete(ep.Version)
	}
}

// AwaitSignal GracefulShutdown
func (gs *GenericServer) AwaitSignal(quit chan os.Signal, to time.Duration) {
	// 接收停止信号
	signal.Notify(quit, dubgo.ShutdownSignals...)
	<-quit
	logger.Infof("SERVER:EVENT:SIGNAL:SHUTDOWN")
	ctx, cancel := goctx.WithTimeout(goctx.Background(), to)
	defer cancel()
	if err := gs.Shutdown(ctx); nil != err {
		logger.Errorw("SERVER:EVENT:SHUTDOWN/error", "error", err)
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
	flux.AssertNotNil(listener, "WebListener Must Not nil")
	flux.AssertNotEmpty(listenerID, "WebListener Id Must Not empty")
	gs.listeners[strings.ToLower(listenerID)] = listener
}

// WebListenerById 返回ListenServer实例
func (gs *GenericServer) WebListenerById(listenerID string) (flux.WebListener, bool) {
	ls, ok := gs.listeners[strings.ToLower(listenerID)]
	return ls, ok
}

// AddContextHookFunc 添加Http与Flux的Context桥接函数
func (gs *GenericServer) AddContextHookFunc(f flux.OnContextHookFunc) {
	gs.onContextHooks = append(gs.onContextHooks, f)
}

func (gs *GenericServer) newEndpointHandler(server flux.WebListener, endpoint *flux.MVCEndpoint) flux.WebHandler {
	return func(webex flux.ServerWebContext) error {
		return gs.route(webex, server, endpoint)
	}
}

func (gs *GenericServer) selectMVCEndpoint(endpoint *flux.EndpointSpec) (*flux.MVCEndpoint, bool) {
	key := ext.MakeEndpointKey(endpoint.HttpMethod, endpoint.HttpPattern)
	if mve, ok := ext.EndpointByKey(key); ok {
		return mve, false
	} else {
		return ext.RegisterEndpoint(key, endpoint), true
	}
}

func (gs *GenericServer) defaultListener() flux.WebListener {
	count := len(gs.listeners)
	if count == 0 {
		return nil
	} else if count == 1 {
		for _, server := range gs.listeners {
			return server
		}
	}
	server, ok := gs.listeners[ListenerIdDefault]
	if ok {
		return server
	}
	return nil
}

// syncService 将Endpoint与Service建立绑定映射；
// 此处绑定的为原始元数据的引用；
func (gs *GenericServer) syncService(ep *flux.EndpointSpec) {
	// Endpoint为静态模型，不支持动态更新
	if ep.AnnotationExists(flux.EndpointAnnotationStaticModel) {
		logger.Infow("SERVER:EVENT:MAPMETA/ignore:static", "ep-pattern", ep.HttpPattern, "ep-service", ep.ServiceId)
		return
	}
	service, ok := ext.ServiceByID(ep.ServiceId)
	if !ok {
		return
	}
	logger.Infow("SERVER:EVENT:MAPMETA/sync-service", "ep-pattern", ep.HttpPattern, "ep-service", ep.ServiceId)
	ep.Service = service
}

// syncEndpoint 将Endpoint与Service建立绑定映射；
// 此处绑定的为原始元数据的引用；
func (gs *GenericServer) syncEndpoint(srv *flux.ServiceSpec) {
	for _, mvce := range ext.Endpoints() {
		for _, ep := range mvce.Endpoints() {
			// Endpoint为静态模型，不支持动态更新
			if ep.AnnotationExists(flux.EndpointAnnotationStaticModel) {
				logger.Infow("SERVER:EVENT:MAPMETA/ignore:static", "ep-pattern", ep.HttpPattern, "ep-service", ep.ServiceId)
				continue
			}
			if toolkit.MatchEqual([]string{srv.ServiceID(), srv.AliasId}, ep.ServiceId) {
				logger.Infow("SERVER:EVENT:MAPMETA/sync-endpoint", "ep-pattern", ep.HttpPattern, "ep-service", ep.ServiceId)
				ep.Service = *srv
			}
		}
	}
}

var copierconf = copier.Option{
	DeepCopy: true,
}

func (gs *GenericServer) dup(toep *flux.EndpointSpec, fromep *flux.EndpointSpec) error {
	return copier.CopyWithOption(toep, fromep, copierconf)
}

func onInitializer(v interface{}, f func(initable flux.Initializer) error) error {
	if init, ok := v.(flux.Initializer); ok {
		return f(init)
	}
	return nil
}

// NewWebListenerConfig 根据WebListenerID，返回初始化WebListener实例时的配置
func NewWebListenerConfig(id string) *flux.Configuration {
	return flux.NewConfigurationByKeys(flux.NamespaceWebListeners, id)
}

// NewWebListenerOptions 根据WebListenerID，返回构建WebListener实例时的选项参数
func NewWebListenerOptions(id string) *flux.Configuration {
	return flux.NewConfigurationByKeys(flux.NamespaceWebListeners, id)
}

func SupportedHttpMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodPut,
		http.MethodHead, http.MethodOptions, http.MethodPatch, http.MethodTrace:
		return true
	default:
		logger.Errorw("Ignore unsupported http method:", "method", method)
		return false
	}
}
