package dubbo

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	_ "github.com/apache/dubbo-go/cluster/cluster_impl"
	_ "github.com/apache/dubbo-go/cluster/loadbalance"
	"github.com/apache/dubbo-go/common/constant"
	_ "github.com/apache/dubbo-go/common/proxy/proxy_factory"
	dubgo "github.com/apache/dubbo-go/config"
	_ "github.com/apache/dubbo-go/filter/filter_impl"
	"github.com/apache/dubbo-go/protocol/dubbo"
	_ "github.com/apache/dubbo-go/registry/nacos"
	_ "github.com/apache/dubbo-go/registry/protocol"
	_ "github.com/apache/dubbo-go/registry/zookeeper"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/backend"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/pkg"
	"github.com/spf13/cast"
)

const (
	configKeyTraceEnable    = "trace-enable"
	configKeyReferenceDelay = "reference-delay"
)

var (
	ErrDecodeInvalidHeaders = errors.New("DUBBRPC:DECODE:INVALID_HEADERS")
	ErrDecodeInvalidStatus  = errors.New("DUBBRPC:DECODE:INVALID_STATUS")
	ErrMessageInvoke        = "DUBBRPC:INVOKE"
	ErrMessageAssemble      = "DUBBRPC:ASSEMBLE"
)

var (
	registryGlobalAlias map[string]string
)

var (
	_ flux.Backend = new(BackendService)
)

func init() {
	SetRegistryGlobalAlias(map[string]string{
		"id":       "dubbo.registry.id",
		"protocol": "dubbo.registry.protocol",
		"group":    "dubbo.registry.protocol",
		"timeout":  "dubbo.registry.timeout",
		"address":  "dubbo.registry.address",
		"username": "dubbo.registry.username",
		"password": "dubbo.registry.password",
	})
}

type (
	// ReferenceOptionsFunc DubboReference配置函数，可外部化配置Dubbo Reference
	ReferenceOptionsFunc func(*flux.BackendService, *flux.Configuration, *dubgo.ReferenceConfig) *dubgo.ReferenceConfig
	// ParameterAssembleFunc Dubbo调用参数封装函数，可外部化配置为其它协议的值对象
	ParameterAssembleFunc func(arguments []flux.Argument, context flux.Context) (types []string, values interface{}, err error)
)

// GetRegistryGlobalAlias 获取默认DubboRegistry全局别名配置
func GetRegistryGlobalAlias() map[string]string {
	return registryGlobalAlias
}

// SetRegistryGlobalAlias 设置DubboRegistry全局别名配置
func SetRegistryGlobalAlias(alias map[string]string) {
	registryGlobalAlias = pkg.RequireNotNil(alias, "alias is nil").(map[string]string)
}

// BackendService 集成DubboRPC框架的BackendService
type BackendService struct {
	// 可外部配置
	ReferenceOptionsFuncs []ReferenceOptionsFunc
	ParameterAssembleFunc ParameterAssembleFunc
	// 内部私有
	traceEnable   bool
	configuration *flux.Configuration
	referenceMu   sync.RWMutex
}

// NewDubboBackend New dubbo backend instance
func NewDubboBackend() flux.Backend {
	return &BackendService{
		ReferenceOptionsFuncs: make([]ReferenceOptionsFunc, 0),
		ParameterAssembleFunc: AssembleHessianArguments,
	}
}

// Configuration get config instance
func (b *BackendService) Configuration() *flux.Configuration {
	return b.configuration
}

// Init init backend
func (b *BackendService) Init(config *flux.Configuration) error {
	logger.Info("Dubbo backend initializing")
	config.SetDefaults(map[string]interface{}{
		configKeyReferenceDelay: time.Millisecond * 30,
		configKeyTraceEnable:    false,
		"timeout":               "5000",
		"retries":               "0",
		"cluster":               "failover",
		"load-balance":          "random",
		"protocol":              dubbo.DUBBO,
	})
	b.configuration = config
	b.traceEnable = config.GetBool(configKeyTraceEnable)
	logger.Infow("Dubbo backend request trace", "enable", b.traceEnable)
	// Set default impl if not present
	if nil == b.ReferenceOptionsFuncs {
		b.ReferenceOptionsFuncs = make([]ReferenceOptionsFunc, 0)
	}
	if pkg.IsNil(b.ParameterAssembleFunc) {
		b.ParameterAssembleFunc = AssembleHessianArguments
	}
	// 修改默认Consumer配置
	consumerc := dubgo.GetConsumerConfig()
	// 支持定义Registry
	registry := b.configuration.Sub("registry")
	registry.SetGlobalAlias(GetRegistryGlobalAlias())
	if id, rconfig := newConsumerRegistry(registry); id != "" && nil != rconfig {
		consumerc.Registries[id] = rconfig
		logger.Infow("Dubbo backend setup registry", "id", id, "config", rconfig)
	}
	dubgo.SetConsumerConfig(consumerc)
	return nil
}

// Startup startup service
func (b *BackendService) Startup() error {
	return nil
}

// Shutdown shutdown service
func (b *BackendService) Shutdown(_ context.Context) error {
	dubgo.BeforeShutdown()
	return nil
}

// Exchange do exchange with context
func (b *BackendService) Exchange(ctx flux.Context) *flux.StateError {
	return backend.DoExchange(ctx, b)
}

// Invoke invoke backend service with context
func (b *BackendService) Invoke(service flux.BackendService, ctx flux.Context) (interface{}, *flux.StateError) {
	types, values, err := b.ParameterAssembleFunc(service.Arguments, ctx)
	if nil != err {
		return nil, &flux.StateError{
			StatusCode: flux.StatusServerError,
			ErrorCode:  flux.ErrorCodeGatewayInternal,
			Message:    ErrMessageAssemble,
			Internal:   err,
		}
	} else {
		return b.ExecuteWith(types, values, service, ctx)
	}
}

// ExecuteWith execute backend service with arguments
func (b *BackendService) ExecuteWith(types []string, values interface{}, service flux.BackendService, ctx flux.Context) (interface{}, *flux.StateError) {
	serviceName := service.Interface + "." + service.Method
	if b.traceEnable {
		logger.TraceContext(ctx).Infow("Dubbo invoking",
			"service", serviceName, "values", values, "types", types, "attrs", ctx.Attributes(),
		)
	}
	// Note: must be map[string]string
	// See: dubbo-go@v1.5.1/common/proxy/proxy.go:150
	attachments, err := cast.ToStringMapStringE(ctx.Attributes())
	if nil != err {
		logger.TraceContext(ctx).Errorw("Dubbo attachment error", "service", serviceName, "error", err)
		return nil, &flux.StateError{
			StatusCode: flux.StatusServerError,
			ErrorCode:  flux.ErrorCodeGatewayInternal,
			Message:    ErrMessageAssemble,
			Internal:   err,
		}
	}
	goctx := context.WithValue(ctx.Context(), constant.AttachmentKey, attachments)
	generic := b.LoadGenericService(&service)
	if resp, err := generic.Invoke(goctx, []interface{}{service.Method, types, values}); err != nil {
		logger.TraceContext(ctx).Errorw("Dubbo rpc error", "service", serviceName, "error", err)
		return nil, &flux.StateError{
			StatusCode: flux.StatusBadGateway,
			ErrorCode:  flux.ErrorCodeGatewayBackend,
			Message:    ErrMessageInvoke,
			Internal:   err,
		}
	} else {
		if b.traceEnable {
			logger.TraceContext(ctx).Infow("Dubbo received response", "service", serviceName,
				"data.type", reflect.TypeOf(resp), "data.value", fmt.Sprintf("%+v", resp))
		}
		return resp, nil
	}
}

// LoadGenericService create and cache dubbo generic service
func (b *BackendService) LoadGenericService(service *flux.BackendService) *dubgo.GenericService {
	b.referenceMu.Lock()
	defer b.referenceMu.Unlock()
	if cached := dubgo.GetConsumerService(service.Interface); nil != cached {
		return cached.(*dubgo.GenericService)
	}
	newRef := NewReference(service.Interface, service, b.configuration)
	// Options
	const msg = "Dubbo option-func return nil reference"
	for _, optsFunc := range b.ReferenceOptionsFuncs {
		if nil != optsFunc {
			newRef = pkg.RequireNotNil(optsFunc(service, b.configuration, newRef), msg).(*dubgo.ReferenceConfig)
		}
	}
	logger.Infow("Create dubbo reference-config, referring", "interface", service.Interface)
	generic := dubgo.NewGenericService(service.Interface)
	dubgo.SetConsumerService(generic)
	newRef.Refer(generic)
	newRef.Implement(generic)
	t := b.configuration.GetDuration(configKeyReferenceDelay)
	if t == 0 {
		t = time.Millisecond * 30
	}
	<-time.After(t)
	logger.Infow("Create dubbo reference-config: OK", "interface", service.Interface)
	return generic
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
