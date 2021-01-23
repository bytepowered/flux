package dubbo

import (
	"context"
	"errors"
	"fmt"
	"github.com/apache/dubbo-go/protocol/dubbo"
	"github.com/bytepowered/flux/ext"
	jsoniter "github.com/json-iterator/go"
	"reflect"
	"sync"
	"time"

	"github.com/apache/dubbo-go/common/constant"
	dubgo "github.com/apache/dubbo-go/config"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/backend"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/pkg"
)

const (
	ConfigKeyTraceEnable    = "trace-enable"
	ConfigKeyReferenceDelay = "reference-delay"
)

func init() {
	ext.StoreBackendTransport(flux.ProtoDubbo, NewBackendTransportService())
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
	// ReferenceOptionsFunc DubboReference配置函数，可外部化配置Dubbo Reference
	ReferenceOptionsFunc func(*flux.BackendService, *flux.Configuration, *dubgo.ReferenceConfig) *dubgo.ReferenceConfig
	// ArgumentsAssembleFunc Dubbo调用参数封装函数，可外部化配置为其它协议的值对象
	ArgumentsAssembleFunc func(arguments []flux.Argument, context flux.Context) (types []string, values interface{}, err error)
	// AttachmentAssembleFun 封装Attachment附件的函数
	AttachmentAssembleFun func(context flux.Context) (interface{}, error)
)

// BackendTransportService 集成DubboRPC框架的BackendService
type BackendTransportService struct {
	// 可外部配置
	defaults             map[string]interface{}       // 配置默认值
	registryAlias        map[string]string            // Registry的别名
	referenceOptionsFunc []ReferenceOptionsFunc       // Dubbo Reference 配置函数
	argAssembleFunc      ArgumentsAssembleFunc        // Dubbo参数封装函数
	attAssembleFunc      AttachmentAssembleFun        // Attachment封装函数
	resultDecodeFunc     flux.BackendResultDecodeFunc // 解析响应结果的函数
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

// WithResultDecodeFunc 用于配置响应数据解析实现函数
func WithResultDecodeFunc(fun flux.BackendResultDecodeFunc) Option {
	return func(service *BackendTransportService) {
		service.resultDecodeFunc = fun
	}
}

// WithReferenceOptionsFunc 用于配置DubboReference的参数配置函数
func WithReferenceOptionsFunc(fun ReferenceOptionsFunc) Option {
	return func(service *BackendTransportService) {
		service.referenceOptionsFunc = append(service.referenceOptionsFunc, fun)
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
		referenceOptionsFunc: make([]ReferenceOptionsFunc, 0),
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
		WithResultDecodeFunc(NewBackendResultDecodeFunc()),
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
	}
	return NewBackendTransportServiceWith(append(opts, overrides...)...)
}

// Configuration get config instance
func (b *BackendTransportService) Configuration() *flux.Configuration {
	return b.configuration
}

// GetResultDecodeFunc returns result decode func
func (b *BackendTransportService) GetResultDecodeFunc() flux.BackendResultDecodeFunc {
	return b.resultDecodeFunc
}

// Init init backend
func (b *BackendTransportService) Init(config *flux.Configuration) error {
	logger.Info("Dubbo backend transport initializing")
	config.SetDefaults(b.defaults)
	b.configuration = config
	b.traceEnable = config.GetBool(ConfigKeyTraceEnable)
	logger.Infow("Dubbo backend transport request trace", "enable", b.traceEnable)
	// Set default impl if not present
	if nil == b.referenceOptionsFunc {
		b.referenceOptionsFunc = make([]ReferenceOptionsFunc, 0)
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
	return backend.Exchange(ctx, b)
}

// Invoke invoke backend service with context
func (b *BackendTransportService) Invoke(service flux.BackendService, ctx flux.Context) (interface{}, *flux.ServeError) {
	types, values, err := b.argAssembleFunc(service.Arguments, ctx)
	if nil != err {
		return nil, &flux.ServeError{
			StatusCode: flux.StatusServerError,
			ErrorCode:  flux.ErrorCodeGatewayInternal,
			Message:    flux.ErrorMessageDubboAssembleFailed,
			Internal:   err,
		}
	} else {
		return b.ExecuteWith(types, values, service, ctx)
	}
}

// ExecuteWith execute backend service with arguments
func (b *BackendTransportService) ExecuteWith(types []string, values interface{}, service flux.BackendService, ctx flux.Context) (interface{}, *flux.ServeError) {
	if b.traceEnable {
		logger.TraceContext(ctx).Infow("Dubbo invoking",
			"backend-service", service.ServiceID(), "arg-values", values, "arg-types", types, "attrs", ctx.Attributes())
	}
	att, err := b.attAssembleFunc(ctx)
	if nil != err {
		logger.TraceContext(ctx).Errorw("Dubbo attachment error",
			"backend-service", service.ServiceID(),
			"error", err)
		return nil, &flux.ServeError{
			StatusCode: flux.StatusServerError,
			ErrorCode:  flux.ErrorCodeGatewayInternal,
			Message:    flux.ErrorMessageDubboAssembleFailed,
			Internal:   err,
		}
	}
	goctx := context.WithValue(ctx.Context(), constant.AttachmentKey, att)
	generic := b.LoadGenericService(&service)
	if resp, err := generic.Invoke(goctx, []interface{}{service.Method, types, values}); err != nil {
		logger.TraceContext(ctx).Errorw("Dubbo rpc error",
			"backend-service", service.ServiceID(), "error", err)
		return nil, &flux.ServeError{
			StatusCode: flux.StatusBadGateway,
			ErrorCode:  flux.ErrorCodeGatewayBackend,
			Message:    flux.ErrorMessageDubboInvokeFailed,
			Internal:   err,
		}
	} else {
		if b.traceEnable {
			text, err := _json.MarshalToString(resp)
			ctxLogger := logger.TraceContext(ctx)
			if nil == err {
				ctxLogger.Infow("Dubbo received response", "backend-service", service.ServiceID(), "response.json", text)
			} else {
				ctxLogger.Infow("Dubbo received response",
					"backend-service", service.ServiceID(), "response.type", reflect.TypeOf(resp), "response.data", fmt.Sprintf("%+v", resp))
			}
		}
		return resp, nil
	}
}

// LoadGenericService create and cache dubbo generic service
func (b *BackendTransportService) LoadGenericService(definition *flux.BackendService) *dubgo.GenericService {
	b.serviceMutex.Lock()
	defer b.serviceMutex.Unlock()
	if service := dubgo.GetConsumerService(definition.Interface); nil != service {
		return service.(*dubgo.GenericService)
	}
	newRef := NewReference(definition.Interface, definition, b.configuration)
	// Options
	const msg = "Dubbo option-func return nil reference"
	for _, optsFunc := range b.referenceOptionsFunc {
		if nil != optsFunc {
			newRef = pkg.RequireNotNil(optsFunc(definition, b.configuration, newRef), msg).(*dubgo.ReferenceConfig)
		}
	}
	logger.Infow("Create dubbo generic service: ING", "interface", definition.Interface)
	service := dubgo.NewGenericService(definition.Interface)
	dubgo.SetConsumerService(service)
	newRef.Refer(service)
	newRef.Implement(service)
	t := b.configuration.GetDuration(ConfigKeyReferenceDelay)
	if t == 0 {
		t = time.Millisecond * 10
	}
	<-time.After(t)
	logger.Infow("Create dubbo generic service: OK", "interface", definition.Interface)
	return service
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
