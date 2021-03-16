package dubbo

import (
	"context"
	"errors"
	"fmt"
	"github.com/bytepowered/flux/flux-node"
	jsoniter "github.com/json-iterator/go"
	"reflect"
	"sync"
	"time"
)

import (
	"github.com/bytepowered/flux/flux-node/backend"
	"github.com/bytepowered/flux/flux-node/ext"
	"github.com/bytepowered/flux/flux-node/logger"
	"github.com/bytepowered/flux/flux-pkg"
)

import (
	"github.com/apache/dubbo-go/common"
	"github.com/apache/dubbo-go/common/constant"
	dubgo "github.com/apache/dubbo-go/config"
	"github.com/apache/dubbo-go/protocol"
	"github.com/apache/dubbo-go/protocol/dubbo"
)

const (
	ConfigKeyTraceEnable    = "trace_enable"
	ConfigKeyReferenceDelay = "reference_delay"
)

func init() {
	ext.RegisterTransporter(flux.ProtoDubbo, NewTransporter())
}

var (
	ErrDecodeInvalidHeaders = errors.New(flux.ErrorMessageDubboDecodeInvalidHeader)
	ErrDecodeInvalidStatus  = errors.New(flux.ErrorMessageDubboDecodeInvalidStatus)
)

var (
	_     flux.Transporter = new(RpcDubboTransporter)
	_json                  = jsoniter.ConfigCompatibleWithStandardLibrary
)

type (
	// Option func to set option
	Option func(*RpcDubboTransporter)
	// ArgumentResolver Dubbo调用参数封装函数，可外部化配置为其它协议的值对象
	ArgumentResolver func(arguments []flux.Argument, context *flux.Context) (types []string, values interface{}, err error)
	// AttachmentResolver 封装Attachment附件的函数
	AttachmentResolver func(context *flux.Context) (interface{}, error)
)

type (
	// GenericOptionsFunc DubboReference配置函数，可外部化配置Dubbo Reference
	GenericOptionsFunc func(*flux.TransporterService, *flux.Configuration, *dubgo.ReferenceConfig) *dubgo.ReferenceConfig
	// GenericServiceFunc 用于构建DubboGo泛型调用Service实例
	GenericServiceFunc func(*flux.TransporterService) common.RPCService
	// GenericInvokeFunc 用于执行Dubbo泛调用方法，返回统一Result数据结构
	GenericInvokeFunc func(ctx context.Context, args []interface{}, rpc common.RPCService) protocol.Result
)

// RpcDubboTransporter 集成DubboRPC框架的BackendService
type RpcDubboTransporter struct {
	// 可外部配置
	defaults  map[string]interface{} // 配置默认值
	registry  map[string]string      // Registry的别名
	optionsf  []GenericOptionsFunc   // Dubbo Reference 配置函数
	servicef  GenericServiceFunc     // Dubbo Service 构建函数
	invokef   GenericInvokeFunc      // 执行Dubbo泛调用的函数
	aresolver ArgumentResolver       // Dubbo参数封装函数
	tresolver AttachmentResolver     // Attachment封装函数
	codec     flux.TransportCodec    // 解析响应结果的函数
	writer    flux.TransportWriter   // Writer
	// 内部私有
	trace         bool
	configuration *flux.Configuration
	servmx        sync.RWMutex
}

// WithArgumentResolver 用于配置Dubbo参数封装实现函数
func WithArgumentResolver(fun ArgumentResolver) Option {
	return func(service *RpcDubboTransporter) {
		service.aresolver = fun
	}
}

// WithAttachmentResolver 用于配置Attachment封装实现函数
func WithAttachmentResolver(fun AttachmentResolver) Option {
	return func(service *RpcDubboTransporter) {
		service.tresolver = fun
	}
}

// WithTransportCodec 用于配置响应数据解析实现函数
func WithTransportCodec(fun flux.TransportCodec) Option {
	return func(service *RpcDubboTransporter) {
		service.codec = fun
	}
}

// WithTransportWriter 用于配置响应数据解析实现函数
func WithTransportWriter(fun flux.TransportWriter) Option {
	return func(service *RpcDubboTransporter) {
		service.writer = fun
	}
}

// WithGenericOptionsFunc 用于配置DubboReference的参数配置函数
func WithGenericOptionsFunc(fun GenericOptionsFunc) Option {
	return func(service *RpcDubboTransporter) {
		service.optionsf = append(service.optionsf, fun)
	}
}

// WithGenericServiceFunc 用于构建DubboRPCService的函数
func WithGenericServiceFunc(fun GenericServiceFunc) Option {
	return func(service *RpcDubboTransporter) {
		service.servicef = fun
	}
}

// WithGenericInvokeFunc 用于调用DubboRPCService执行方法的函数
func WithGenericInvokeFunc(fun GenericInvokeFunc) Option {
	return func(service *RpcDubboTransporter) {
		service.invokef = fun
	}
}

// WithRegistryAlias 用于配置DubboRegistry注册中心的配置别名
func WithRegistryAlias(alias map[string]string) Option {
	return func(service *RpcDubboTransporter) {
		service.registry = alias
	}
}

// WithDefaults 用于配置默认配置值
func WithDefaults(defaults map[string]interface{}) Option {
	return func(service *RpcDubboTransporter) {
		service.defaults = defaults
	}
}

// NewTransporterWith New dubbo backend service with optionsf
func NewTransporterWith(opts ...Option) flux.Transporter {
	bts := &RpcDubboTransporter{
		optionsf: make([]GenericOptionsFunc, 0),
	}
	for _, opt := range opts {
		opt(bts)
	}
	return bts
}

// NewTransporter New dubbo backend instance
func NewTransporter() flux.Transporter {
	return NewTransporterOverride()
}

// NewTransporterOverride New dubbo backend instance
func NewTransporterOverride(overrides ...Option) flux.Transporter {
	opts := []Option{
		WithRegistryAlias(map[string]string{
			"id":       "dubbo.registry.id",
			"protocol": "dubbo.registry.protocol",
			"group":    "dubbo.registry.group",
			"timeout":  "dubbo.registry.timeout",
			"address":  "dubbo.registry.address",
			"username": "dubbo.registry.username",
			"password": "dubbo.registry.password",
		}),
		WithDefaults(map[string]interface{}{
			ConfigKeyReferenceDelay: time.Millisecond * 10,
			ConfigKeyTraceEnable:    false,
			"timeout":               "5000",
			"retries":               "0",
			"cluster":               "failover",
			"load_balance":          "random",
			"protocol":              dubbo.DUBBO,
		}),
		WithGenericServiceFunc(func(backend *flux.TransporterService) common.RPCService {
			return dubgo.NewGenericService(backend.Interface)
		}),
		WithGenericInvokeFunc(func(ctx context.Context, args []interface{}, rpc common.RPCService) protocol.Result {
			srv := rpc.(*dubgo.GenericService)
			data, err := srv.Invoke(ctx, args)
			return &protocol.RPCResult{
				Attrs: nil,
				Err:   err,
				Rest:  data,
			}
		}),
		WithArgumentResolver(DefaultArgumentResolver),
		WithAttachmentResolver(DefaultAttachmentResolver),
		WithTransportCodec(NewTransportCodecFunc()),
		WithTransportWriter(new(backend.RpcTransportWriter)),
	}
	return NewTransporterWith(append(opts, overrides...)...)
}

func (b *RpcDubboTransporter) Writer() flux.TransportWriter {
	return b.writer
}

// Init init backend
func (b *RpcDubboTransporter) Init(config *flux.Configuration) error {
	logger.Info("Dubbo backend transport initializing")
	config.SetDefaults(b.defaults)
	b.configuration = config
	b.trace = config.GetBool(ConfigKeyTraceEnable)
	logger.Infow("Dubbo backend transport request trace", "enable", b.trace)
	// Set default impl if not present
	if nil == b.optionsf {
		b.optionsf = make([]GenericOptionsFunc, 0)
	}
	if fluxpkg.IsNil(b.aresolver) {
		b.aresolver = DefaultArgumentResolver
	}
	// 修改默认Consumer配置
	consumerc := dubgo.GetConsumerConfig()
	// 支持定义Registry
	registry := b.configuration.Sub("registry")
	registry.SetGlobalAlias(b.registry)
	if id, rconfig := newConsumerRegistry(registry); id != "" && nil != rconfig {
		consumerc.Registries[id] = rconfig
		logger.Infow("Dubbo backend transport setup registry", "id", id, "config", rconfig)
	}
	dubgo.SetConsumerConfig(consumerc)
	return nil
}

// Startup startup service
func (b *RpcDubboTransporter) Startup() error {
	return nil
}

// Shutdown shutdown service
func (b *RpcDubboTransporter) Shutdown(_ context.Context) error {
	dubgo.BeforeShutdown()
	return nil
}

// Transport do exchange with context
func (b *RpcDubboTransporter) Transport(ctx *flux.Context) {
	backend.DoTransport(ctx, b)
}

// Invoke invoke backend service with context
func (b *RpcDubboTransporter) Invoke(ctx *flux.Context, service flux.TransporterService) (interface{}, *flux.ServeError) {
	types, values, err := b.aresolver(service.Arguments, ctx)
	if nil != err {
		return nil, &flux.ServeError{
			StatusCode: flux.StatusServerError,
			ErrorCode:  flux.ErrorCodeGatewayInternal,
			Message:    flux.ErrorMessageDubboAssembleFailed,
			CauseError: err,
		}
	} else {
		return b.DoInvoke(types, values, service, ctx)
	}
}

func (b *RpcDubboTransporter) InvokeCodec(ctx *flux.Context, service flux.TransporterService) (*flux.ResponseBody, *flux.ServeError) {
	raw, serr := b.Invoke(ctx, service)
	select {
	case <-ctx.Context().Done():
		logger.TraceContext(ctx).Infow("BACKEND:DUBBO:RPC_CANCELED",
			"backend-service", service.ServiceID(), "error", ctx.Context().Err())
		return nil, serr
	default:
		break
	}
	if nil != serr {
		logger.TraceContext(ctx).Errorw("BACKEND:DUBBO:RPC_ERROR",
			"backend-service", service.ServiceID(), "error", serr.CauseError)
		return nil, serr
	}
	// decode response
	result, err := b.codec(ctx, raw)
	if nil != err {
		return nil, &flux.ServeError{
			StatusCode: flux.StatusServerError,
			ErrorCode:  flux.ErrorCodeGatewayInternal,
			Message:    flux.ErrorMessageBackendDecodeResponse,
			CauseError: fmt.Errorf("decode dubbo response, err: %w", err),
		}
	}
	fluxpkg.AssertNotNil(result, "dubbo: <result> must not nil, request.id: "+ctx.RequestId())
	return result, nil
}

// DoInvoke execute backend service with arguments
func (b *RpcDubboTransporter) DoInvoke(types []string, values interface{}, service flux.TransporterService, ctx *flux.Context) (interface{}, *flux.ServeError) {
	if b.trace {
		logger.TraceContext(ctx).Infow("BACKEND:DUBBO:INVOKE",
			"backend-service", service.ServiceID(), "arg-values", values, "arg-types", types, "attrs", ctx.Attributes())
	}
	att, err := b.tresolver(ctx)
	if nil != err {
		logger.TraceContext(ctx).Errorw("BACKEND:DUBBO:ATTACHMENT",
			"backend-service", service.ServiceID(), "error", err)
		return nil, &flux.ServeError{
			StatusCode: flux.StatusServerError,
			ErrorCode:  flux.ErrorCodeGatewayInternal,
			Message:    flux.ErrorMessageDubboAssembleFailed,
			CauseError: err,
		}
	}
	generic := b.LoadGenericService(&service)
	goctx := context.WithValue(ctx.Context(), constant.AttachmentKey, att)
	resultW := b.invokef(goctx, []interface{}{service.Method, types, values}, generic)
	if cause := resultW.Error(); cause != nil {
		return nil, &flux.ServeError{
			StatusCode: flux.StatusBadGateway,
			ErrorCode:  flux.ErrorCodeGatewayBackend,
			Message:    flux.ErrorMessageDubboInvokeFailed,
			CauseError: cause,
		}
	} else {
		if b.trace {
			data := resultW.Result()
			text, err := _json.MarshalToString(data)
			ctxLogger := logger.TraceContext(ctx)
			if nil == err {
				ctxLogger.Infow("BACKEND:DUBBO:RECEIVED", "backend-service", service.ServiceID(), "response.json", text)
			} else {
				ctxLogger.Infow("BACKEND:DUBBO:RECEIVED",
					"backend-service", service.ServiceID(), "response.type", reflect.TypeOf(data), "response.data", fmt.Sprintf("%+v", data))
			}
		}
		return resultW, nil
	}
}

// LoadGenericService create and cache dubbo generic service
func (b *RpcDubboTransporter) LoadGenericService(backend *flux.TransporterService) common.RPCService {
	b.servmx.Lock()
	defer b.servmx.Unlock()
	if srv := dubgo.GetConsumerService(backend.Interface); nil != srv {
		return srv
	}
	newRef := NewReference(backend.Interface, backend, b.configuration)
	// Options
	const msg = "Dubbo option-func return nil reference"
	for _, optsFunc := range b.optionsf {
		if nil != optsFunc {
			newRef = fluxpkg.MustNotNil(optsFunc(backend, b.configuration, newRef), msg).(*dubgo.ReferenceConfig)
		}
	}
	logger.Infow("DUBBO:GENERIC:CREATE: PREPARE", "interface", backend.Interface)
	srv := b.servicef(backend)
	dubgo.SetConsumerService(srv)
	newRef.Refer(srv)
	newRef.Implement(srv)
	t := b.configuration.GetDuration(ConfigKeyReferenceDelay)
	if t == 0 {
		t = time.Millisecond * 10
	}
	<-time.After(t)
	logger.Infow("DUBBO:GENERIC:CREATE: OJBK", "interface", backend.Interface)
	return srv
}

func newConsumerRegistry(config *flux.Configuration) (string, *dubgo.RegistryConfig) {
	if !config.IsSet("id", "protocol") {
		return "", nil
	}
	return config.GetString("id"), &dubgo.RegistryConfig{
		Protocol:   config.GetString("protocol"),
		TimeoutStr: config.GetString("timeout"),
		Group:      config.GetString("group"),
		TTL:        config.GetString("ttl"),
		Address:    config.GetString("address"),
		Username:   config.GetString("username"),
		Password:   config.GetString("password"),
		Simplified: config.GetBool("simplified"),
		Preferred:  config.GetBool("preferred"),
		Zone:       config.GetString("zone"),
		Weight:     config.GetInt64("weight"),
		Params:     config.GetStringMapString("params"),
	}
}

func NewReference(refid string, service *flux.TransporterService, config *flux.Configuration) *dubgo.ReferenceConfig {
	logger.Infow("Create dubbo reference-config",
		"service", service.Interface, "remote-host", service.RemoteHost,
		"rpc-group", service.RpcGroup(), "rpc-version", service.RpcVersion())
	ref := dubgo.NewReferenceConfig(refid, context.Background())
	ref.Url = service.RemoteHost
	ref.InterfaceName = service.Interface
	ref.Version = service.RpcVersion()
	ref.Group = service.RpcGroup()
	ref.RequestTimeout = service.RpcTimeout()
	ref.Retries = service.RpcRetries()
	ref.Cluster = config.GetString("cluster")
	ref.Protocol = config.GetString("protocol")
	ref.Loadbalance = config.GetString("load_balance")
	ref.Generic = true
	return ref
}
