package dubbo

import (
	"context"
	"errors"
	"fmt"

	jsoniter "github.com/json-iterator/go"
	"reflect"
	"sync"
	"time"
)

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/backend"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/pkg"
)

import (
	"github.com/apache/dubbo-go/common"
	"github.com/apache/dubbo-go/common/constant"
	dubgo "github.com/apache/dubbo-go/config"
	"github.com/apache/dubbo-go/protocol"
	"github.com/apache/dubbo-go/protocol/dubbo"
)

const (
	ConfigKeyTraceEnable    = "trace-enable"
	ConfigKeyReferenceDelay = "reference-delay"
)

func init() {
	ext.SetBackendTransport(flux.ProtoDubbo, NewBackendTransportService())
}

var (
	ErrDecodeInvalidHeaders = errors.New(flux.ErrorMessageDubboDecodeInvalidHeader)
	ErrDecodeInvalidStatus  = errors.New(flux.ErrorMessageDubboDecodeInvalidStatus)
)

var (
	_     flux.BackendTransport = new(BackendTransportService)
	_json                       = jsoniter.ConfigCompatibleWithStandardLibrary
)

type (
	// Option func to set option
	Option func(service *BackendTransportService)
)

type (
	// ArgumentsAssembleFunc Dubbo调用参数封装函数，可外部化配置为其它协议的值对象
	ArgumentsAssembleFunc func(arguments []flux.Argument, context flux.Context) (types []string, values interface{}, err error)
	// AttachmentAssembleFun 封装Attachment附件的函数
	AttachmentAssembleFun func(context flux.Context) (interface{}, error)
)

type (
	// DubboGenericOptionsFunc DubboReference配置函数，可外部化配置Dubbo Reference
	DubboGenericOptionsFunc func(*flux.BackendService, *flux.Configuration, *dubgo.ReferenceConfig) *dubgo.ReferenceConfig
	// DubboGenericServiceFunc 用于构建DubboGo泛型调用Service实例
	DubboGenericServiceFunc func(*flux.BackendService) common.RPCService
	// DubboGenericInvokeFunc 用于执行Dubbo泛调用方法，返回统一Result数据结构
	DubboGenericInvokeFunc func(ctx context.Context, args []interface{}, rpc common.RPCService) protocol.Result
)

// BackendTransportService 集成DubboRPC框架的BackendService
type BackendTransportService struct {
	// 可外部配置
	defaults          map[string]interface{}        // 配置默认值
	registryAlias     map[string]string             // Registry的别名
	dubboOptionsFunc  []DubboGenericOptionsFunc     // Dubbo Reference 配置函数
	dubboServiceFunc  DubboGenericServiceFunc       // Dubbo Service 构建函数
	dubboInvokeFunc   DubboGenericInvokeFunc        // 执行Dubbo泛调用的函数
	argAssembleFunc   ArgumentsAssembleFunc         // Dubbo参数封装函数
	attAssembleFunc   AttachmentAssembleFun         // Attachment封装函数
	responseCodecFunc flux.BackendResponseCodecFunc // 解析响应结果的函数
	// 内部私有
	traceEnable   bool
	configuration *flux.Configuration
	serviceMutex  sync.RWMutex
}

// WithArgumentAssembleFunc 用于配置Dubbo参数封装实现函数
func WithArgumentAssembleFunc(fun ArgumentsAssembleFunc) Option {
	return func(service *BackendTransportService) {
		service.argAssembleFunc = fun
	}
}

// WithAttachmentAssembleFunc 用于配置Attachment封装实现函数
func WithAttachmentAssembleFunc(fun AttachmentAssembleFun) Option {
	return func(service *BackendTransportService) {
		service.attAssembleFunc = fun
	}
}

// WithResponseCodecFunc 用于配置响应数据解析实现函数
func WithResponseCodecFunc(fun flux.BackendResponseCodecFunc) Option {
	return func(service *BackendTransportService) {
		service.responseCodecFunc = fun
	}
}

// WithGenericOptionsFunc 用于配置DubboReference的参数配置函数
func WithGenericOptionsFunc(fun DubboGenericOptionsFunc) Option {
	return func(service *BackendTransportService) {
		service.dubboOptionsFunc = append(service.dubboOptionsFunc, fun)
	}
}

// WithGenericServiceFunc 用于构建DubboRPCService的函数
func WithGenericServiceFunc(fun DubboGenericServiceFunc) Option {
	return func(service *BackendTransportService) {
		service.dubboServiceFunc = fun
	}
}

// WithGenericInvokeFunc 用于调用DubboRPCService执行方法的函数
func WithGenericInvokeFunc(fun DubboGenericInvokeFunc) Option {
	return func(service *BackendTransportService) {
		service.dubboInvokeFunc = fun
	}
}

// WithRegistryAlias 用于配置DubboRegistry注册中心的配置别名
func WithRegistryAlias(alias map[string]string) Option {
	return func(service *BackendTransportService) {
		service.registryAlias = alias
	}
}

// WithDefaults 用于配置默认配置值
func WithDefaults(defaults map[string]interface{}) Option {
	return func(service *BackendTransportService) {
		service.defaults = defaults
	}
}

// NewBackendTransportServiceWith New dubbo backend service with options
func NewBackendTransportServiceWith(opts ...Option) flux.BackendTransport {
	bts := &BackendTransportService{
		dubboOptionsFunc: make([]DubboGenericOptionsFunc, 0),
	}
	for _, opt := range opts {
		opt(bts)
	}
	return bts
}

// NewBackendTransportService New dubbo backend instance
func NewBackendTransportService() flux.BackendTransport {
	return NewBackendTransportServiceOverrides()
}

// NewBackendTransportServiceOverrides New dubbo backend instance
func NewBackendTransportServiceOverrides(overrides ...Option) flux.BackendTransport {
	opts := []Option{
		WithArgumentAssembleFunc(DefaultArgAssembleFunc),
		WithAttachmentAssembleFunc(DefaultAttAssembleFun),
		WithResponseCodecFunc(NewBackendResponseCodecFunc()),
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
			"load-balance":          "random",
			"protocol":              dubbo.DUBBO,
		}),
		WithGenericServiceFunc(func(backend *flux.BackendService) common.RPCService {
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
	}
	return NewBackendTransportServiceWith(append(opts, overrides...)...)
}

// Configuration get config instance
func (b *BackendTransportService) Configuration() *flux.Configuration {
	return b.configuration
}

// GetResultDecodeFunc returns result decode func
func (b *BackendTransportService) GetResponseCodecFunc() flux.BackendResponseCodecFunc {
	return b.responseCodecFunc
}

// Init init backend
func (b *BackendTransportService) Init(config *flux.Configuration) error {
	logger.Info("Dubbo backend transport initializing")
	config.SetDefaults(b.defaults)
	b.configuration = config
	b.traceEnable = config.GetBool(ConfigKeyTraceEnable)
	logger.Infow("Dubbo backend transport request trace", "enable", b.traceEnable)
	// Set default impl if not present
	if nil == b.dubboOptionsFunc {
		b.dubboOptionsFunc = make([]DubboGenericOptionsFunc, 0)
	}
	if pkg.IsNil(b.argAssembleFunc) {
		b.argAssembleFunc = DefaultArgAssembleFunc
	}
	// 修改默认Consumer配置
	consumerc := dubgo.GetConsumerConfig()
	// 支持定义Registry
	registry := b.configuration.Sub("registry")
	registry.SetGlobalAlias(b.registryAlias)
	if id, rconfig := newConsumerRegistry(registry); id != "" && nil != rconfig {
		consumerc.Registries[id] = rconfig
		logger.Infow("Dubbo backend transport setup registry", "id", id, "config", rconfig)
	}
	dubgo.SetConsumerConfig(consumerc)
	return nil
}

// Startup startup service
func (b *BackendTransportService) Startup() error {
	return nil
}

// Shutdown shutdown service
func (b *BackendTransportService) Shutdown(_ context.Context) error {
	dubgo.BeforeShutdown()
	return nil
}

// Exchange do exchange with context
func (b *BackendTransportService) Exchange(ctx flux.Context) *flux.ServeError {
	return backend.DoExchangeTransport(ctx, b)
}

// Invoke invoke backend service with context
func (b *BackendTransportService) Invoke(ctx flux.Context, service flux.BackendService) (interface{}, *flux.ServeError) {
	types, values, err := b.argAssembleFunc(service.Arguments, ctx)
	if nil != err {
		return nil, &flux.ServeError{
			StatusCode: flux.StatusServerError,
			ErrorCode:  flux.ErrorCodeGatewayInternal,
			Message:    flux.ErrorMessageDubboAssembleFailed,
			Internal:   err,
		}
	} else {
		return b.DoInvoke(types, values, service, ctx)
	}
}

func (b *BackendTransportService) InvokeCodec(ctx flux.Context, service flux.BackendService) (*flux.BackendResponse, *flux.ServeError) {
	raw, serr := b.Invoke(ctx, service)
	if nil != serr {
		return nil, serr
	}
	// decode response
	result, err := b.GetResponseCodecFunc()(ctx, raw)
	if nil != err {
		return nil, &flux.ServeError{
			StatusCode: flux.StatusServerError,
			ErrorCode:  flux.ErrorCodeGatewayInternal,
			Message:    flux.ErrorMessageBackendDecodeResponse,
			Internal:   fmt.Errorf("decode dubbo response, err: %w", err),
		}
	}
	return result, nil
}

// DoInvoke execute backend service with arguments
func (b *BackendTransportService) DoInvoke(types []string, values interface{}, service flux.BackendService, ctx flux.Context) (interface{}, *flux.ServeError) {
	if b.traceEnable {
		logger.WithContext(ctx).Infow("BACKEND:DUBBO:INVOKE",
			"backend-service", service.ServiceID(), "arg-values", values, "arg-types", types, "attrs", ctx.Attributes())
	}
	att, err := b.attAssembleFunc(ctx)
	if nil != err {
		logger.WithContext(ctx).Errorw("Dubbo attachment error",
			"backend-service", service.ServiceID(), "error", err)
		return nil, &flux.ServeError{
			StatusCode: flux.StatusServerError,
			ErrorCode:  flux.ErrorCodeGatewayInternal,
			Message:    flux.ErrorMessageDubboAssembleFailed,
			Internal:   err,
		}
	}
	goctx := context.WithValue(ctx.Context(), constant.AttachmentKey, att)
	generic := b.LoadGenericService(&service)
	resultW := b.dubboInvokeFunc(goctx, []interface{}{service.Method, types, values}, generic)
	if err := resultW.Error(); err != nil {
		logger.WithContext(ctx).Errorw("BACKEND:DUBBO:RPC_ERROR",
			"backend-service", service.ServiceID(), "error", err)
		return nil, &flux.ServeError{
			StatusCode: flux.StatusBadGateway,
			ErrorCode:  flux.ErrorCodeGatewayBackend,
			Message:    flux.ErrorMessageDubboInvokeFailed,
			Internal:   err,
		}
	} else {
		if b.traceEnable {
			data := resultW.Result()
			text, err := _json.MarshalToString(data)
			ctxLogger := logger.WithContext(ctx)
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
func (b *BackendTransportService) LoadGenericService(backend *flux.BackendService) common.RPCService {
	b.serviceMutex.Lock()
	defer b.serviceMutex.Unlock()
	if srv := dubgo.GetConsumerService(backend.Interface); nil != srv {
		return srv
	}
	newRef := NewReference(backend.Interface, backend, b.configuration)
	// Options
	const msg = "Dubbo option-func return nil reference"
	for _, optsFunc := range b.dubboOptionsFunc {
		if nil != optsFunc {
			newRef = pkg.RequireNotNil(optsFunc(backend, b.configuration, newRef), msg).(*dubgo.ReferenceConfig)
		}
	}
	logger.Infow("Create dubbo generic service: PENDING", "interface", backend.Interface)
	srv := b.dubboServiceFunc(backend)
	dubgo.SetConsumerService(srv)
	newRef.Refer(srv)
	newRef.Implement(srv)
	t := b.configuration.GetDuration(ConfigKeyReferenceDelay)
	if t == 0 {
		t = time.Millisecond * 10
	}
	<-time.After(t)
	logger.Infow("Create dubbo generic service: OK", "interface", backend.Interface)
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

func NewReference(refid string, service *flux.BackendService, config *flux.Configuration) *dubgo.ReferenceConfig {
	logger.Infow("Create dubbo reference-config",
		"service", service.Interface, "remote-host", service.RemoteHost,
		"rpc-group", service.AttrRpcGroup(), "rpc-version", service.AttrRpcVersion())
	ref := dubgo.NewReferenceConfig(refid, context.Background())
	ref.Url = service.RemoteHost
	ref.InterfaceName = service.Interface
	ref.Version = service.AttrRpcVersion()
	ref.Group = service.AttrRpcGroup()
	ref.RequestTimeout = service.AttrRpcTimeout()
	ref.Retries = service.AttrRpcRetries()
	ref.Cluster = config.GetString("cluster")
	ref.Protocol = config.GetString("protocol")
	ref.Loadbalance = config.GetString("load-balance")
	ref.Generic = true
	return ref
}
