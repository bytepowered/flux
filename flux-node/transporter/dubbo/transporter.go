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
	"github.com/bytepowered/flux/flux-node/ext"
	"github.com/bytepowered/flux/flux-node/logger"
	"github.com/bytepowered/flux/flux-node/transporter"
	"github.com/bytepowered/flux/flux-pkg"
)

import (
	"github.com/apache/dubbo-go/common"
	"github.com/apache/dubbo-go/common/constant"
	dubgo "github.com/apache/dubbo-go/config"
	"github.com/apache/dubbo-go/protocol"
	"github.com/apache/dubbo-go/protocol/dubbo"
)

import (
	_ "github.com/apache/dubbo-go/cluster/cluster_impl"
	_ "github.com/apache/dubbo-go/cluster/loadbalance"
	_ "github.com/apache/dubbo-go/filter/filter_impl"
	_ "github.com/apache/dubbo-go/registry/protocol"
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
	_     flux.Transporter = new(RpcTransporter)
	_json                  = jsoniter.ConfigCompatibleWithStandardLibrary
)

type (
	// Option func to set option
	Option func(*RpcTransporter)
	// ArgumentResolver Dubbo调用参数封装函数，可外部化配置为其它协议的值对象
	ArgumentResolver func(arguments []flux.Argument, context *flux.Context) (types []string, values interface{}, err error)
	// AttachmentResolver 封装Attachment附件的函数
	AttachmentResolver func(context *flux.Context) (interface{}, error)
)

type (
	// GenericOptionsFunc DubboReference配置函数，可外部化配置Dubbo Reference
	GenericOptionsFunc func(*flux.Service, *flux.Configuration, *dubgo.ReferenceConfig) *dubgo.ReferenceConfig
	// GenericServiceFunc 用于构建DubboGo泛型调用Service实例
	GenericServiceFunc func(*flux.Service) common.RPCService
	// GenericInvokeFunc 用于执行Dubbo泛调用方法，返回统一Result数据结构
	GenericInvokeFunc func(ctx context.Context, args []interface{}, rpc common.RPCService) protocol.Result
)

// RpcTransporter 集成DubboRPC框架的TransporterService
type RpcTransporter struct {
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
	return func(service *RpcTransporter) {
		service.aresolver = fun
	}
}

// WithAttachmentResolver 用于配置Attachment封装实现函数
func WithAttachmentResolver(fun AttachmentResolver) Option {
	return func(service *RpcTransporter) {
		service.tresolver = fun
	}
}

// WithTransportCodec 用于配置响应数据解析实现函数
func WithTransportCodec(fun flux.TransportCodec) Option {
	return func(service *RpcTransporter) {
		service.codec = fun
	}
}

// WithTransportWriter 用于配置响应数据解析实现函数
func WithTransportWriter(fun flux.TransportWriter) Option {
	return func(service *RpcTransporter) {
		service.writer = fun
	}
}

// WithGenericOptionsFunc 用于配置DubboReference的参数配置函数
func WithGenericOptionsFunc(fun GenericOptionsFunc) Option {
	return func(service *RpcTransporter) {
		service.optionsf = append(service.optionsf, fun)
	}
}

// WithGenericServiceFunc 用于构建DubboRPCService的函数
func WithGenericServiceFunc(fun GenericServiceFunc) Option {
	return func(service *RpcTransporter) {
		service.servicef = fun
	}
}

// WithGenericInvokeFunc 用于调用DubboRPCService执行方法的函数
func WithGenericInvokeFunc(fun GenericInvokeFunc) Option {
	return func(service *RpcTransporter) {
		service.invokef = fun
	}
}

// WithRegistryAlias 用于配置DubboRegistry注册中心的配置别名
func WithRegistryAlias(alias map[string]string) Option {
	return func(service *RpcTransporter) {
		service.registry = alias
	}
}

// WithDefaults 用于配置默认配置值
func WithDefaults(defaults map[string]interface{}) Option {
	return func(service *RpcTransporter) {
		service.defaults = defaults
	}
}

// NewTransporterWith New dubbo transporter service with options
func NewTransporterWith(opts ...Option) flux.Transporter {
	bts := &RpcTransporter{
		optionsf: make([]GenericOptionsFunc, 0),
	}
	for _, opt := range opts {
		opt(bts)
	}
	return bts
}

// NewTransporter New dubbo transporter instance
func NewTransporter() flux.Transporter {
	return NewTransporterOverride()
}

// NewTransporterOverride New dubbo transporter instance
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
		// 使用带Result结果的RPCService实现
		WithGenericServiceFunc(func(service *flux.Service) common.RPCService {
			return NewResultRPCService(service.Interface)
		}),
		// 转换为 ResultRPCService 的调用
		WithGenericInvokeFunc(func(ctx context.Context, args []interface{}, service common.RPCService) protocol.Result {
			return service.(*ResultRPCService).Invoke(ctx, args)
		}),
		WithArgumentResolver(DefaultArgumentResolver),
		WithAttachmentResolver(DefaultAttachmentResolver),
		WithTransportCodec(NewTransportCodecFunc()),
		WithTransportWriter(new(transporter.DefaultTransportWriter)),
	}
	return NewTransporterWith(append(opts, overrides...)...)
}

func (b *RpcTransporter) Writer() flux.TransportWriter {
	return b.writer
}

// Init init transporter
func (b *RpcTransporter) Init(config *flux.Configuration) error {
	logger.Info("Dubbo transporter transporter initializing")
	config.SetDefaults(b.defaults)
	b.configuration = config
	b.trace = config.GetBool(ConfigKeyTraceEnable)
	logger.Infow("Dubbo transporter transporter request trace", "enable", b.trace)
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
	registry.SetKeyAlias(b.registry)
	if id, rconfig := newConsumerRegistry(registry); id != "" && nil != rconfig {
		consumerc.Registries[id] = rconfig
		logger.Infow("Dubbo transporter transporter setup registry", "id", id, "config", rconfig)
	}
	dubgo.SetConsumerConfig(consumerc)
	return nil
}

// Startup startup service
func (b *RpcTransporter) Startup() error {
	return nil
}

// Shutdown shutdown service
func (b *RpcTransporter) Shutdown(_ context.Context) error {
	dubgo.BeforeShutdown()
	return nil
}

// Transport do exchange with context
func (b *RpcTransporter) Transport(ctx *flux.Context) {
	transporter.DoTransport(ctx, b)
}

// Invoke invoke transporter service with context
func (b *RpcTransporter) Invoke(ctx *flux.Context, service flux.Service) (interface{}, *flux.ServeError) {
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

func (b *RpcTransporter) InvokeCodec(ctx *flux.Context, service flux.Service) (*flux.ResponseBody, *flux.ServeError) {
	raw, serr := b.Invoke(ctx, service)
	select {
	case <-ctx.Context().Done():
		logger.TraceContext(ctx).Infow("TRANSPORTER:DUBBO:RPC_CANCELED",
			"transporter-service", service.ServiceID(), "error", ctx.Context().Err())
		return nil, serr
	default:
		break
	}
	if nil != serr {
		logger.TraceContext(ctx).Errorw("TRANSPORTER:DUBBO:RPC_ERROR",
			"transporter-service", service.ServiceID(), "error", serr.CauseError)
		return nil, serr
	}
	// decode response
	result, err := b.codec(ctx, raw)
	if nil != err {
		return nil, &flux.ServeError{
			StatusCode: flux.StatusServerError,
			ErrorCode:  flux.ErrorCodeGatewayInternal,
			Message:    flux.ErrorMessageTransportDecodeResponse,
			CauseError: fmt.Errorf("decode dubbo response, err: %w", err),
		}
	}
	fluxpkg.AssertNotNil(result, "dubbo: <result> must not nil, request.id: "+ctx.RequestId())
	return result, nil
}

// DoInvoke execute transporter service with arguments
func (b *RpcTransporter) DoInvoke(types []string, values interface{}, service flux.Service, ctx *flux.Context) (interface{}, *flux.ServeError) {
	att, err := b.tresolver(ctx)
	if nil != err {
		logger.TraceContext(ctx).Errorw("TRANSPORTER:DUBBO:ATTACHMENT",
			"transporter-service", service.ServiceID(), "error", err)
		return nil, &flux.ServeError{
			StatusCode: flux.StatusServerError,
			ErrorCode:  flux.ErrorCodeGatewayInternal,
			Message:    flux.ErrorMessageDubboAssembleFailed,
			CauseError: err,
		}
	}
	if b.trace {
		logger.TraceContext(ctx).Infow("TRANSPORTER:DUBBO:INVOKE",
			"transporter-service", service.ServiceID(), "arg-values", values, "arg-types", types, "attrs", att)
	}
	generic := b.LoadGenericService(&service)
	goctx := context.WithValue(ctx.Context(), constant.AttachmentKey, att)
	resultW := b.invokef(goctx, []interface{}{service.Method, types, values}, generic)
	if cause := resultW.Error(); cause != nil {
		return nil, &flux.ServeError{
			StatusCode: flux.StatusBadGateway,
			ErrorCode:  flux.ErrorCodeGatewayTransporter,
			Message:    flux.ErrorMessageDubboInvokeFailed,
			CauseError: cause,
		}
	} else {
		if b.trace {
			data := resultW.Result()
			text, err := _json.MarshalToString(data)
			ctxLogger := logger.TraceContext(ctx)
			if nil == err {
				ctxLogger.Infow("TRANSPORTER:DUBBO:RECEIVED", "transporter-service", service.ServiceID(), "response.json", text)
			} else {
				ctxLogger.Infow("TRANSPORTER:DUBBO:RECEIVED",
					"transporter-service", service.ServiceID(), "response.type", reflect.TypeOf(data), "response.data", fmt.Sprintf("%+v", data))
			}
		}
		return resultW, nil
	}
}

// LoadGenericService create and cache dubbo generic service
func (b *RpcTransporter) LoadGenericService(service *flux.Service) common.RPCService {
	b.servmx.Lock()
	defer b.servmx.Unlock()
	if srv := dubgo.GetConsumerService(service.Interface); nil != srv {
		return srv
	}
	newRef := NewReference(service.Interface, service, b.configuration)
	// Options
	const msg = "Dubbo option-func return nil reference"
	for _, optsFunc := range b.optionsf {
		if nil != optsFunc {
			newRef = fluxpkg.MustNotNil(optsFunc(service, b.configuration, newRef), msg).(*dubgo.ReferenceConfig)
		}
	}
	logger.Infow("DUBBO:GENERIC:CREATE: PREPARE", "interface", service.Interface)
	srv := b.servicef(service)
	dubgo.SetConsumerService(srv)
	newRef.Refer(srv)
	newRef.Implement(srv)
	t := b.configuration.GetDuration(ConfigKeyReferenceDelay)
	if t == 0 {
		t = time.Millisecond * 10
	}
	<-time.After(t)
	logger.Infow("DUBBO:GENERIC:CREATE: OJBK", "interface", service.Interface)
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

func NewReference(refid string, service *flux.Service, config *flux.Configuration) *dubgo.ReferenceConfig {
	logger.Infow("Create dubbo reference-config",
		"target-service", service.Interface, "target-url", service.Url,
		"rpc-group", service.RpcGroup(), "rpc-version", service.RpcVersion())
	ref := dubgo.NewReferenceConfig(refid, context.Background())
	ref.Url = service.Url
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
