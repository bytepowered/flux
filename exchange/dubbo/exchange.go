package dubbo

import (
	"context"
	"errors"
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
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/pkg"
	"github.com/bytepowered/flux/support"
	"github.com/spf13/cast"
	"reflect"
	"sync"
	"time"
)

const (
	configKeyTraceEnable    = "trace-enable"
	configKeyReferenceDelay = "reference-delay"
)

var (
	ErrInvalidHeaders  = errors.New("DUEXRPC:INVALID_HEADERS")
	ErrInvalidStatus   = errors.New("DUEXRPC:INVALID_STATUS")
	ErrMessageInvoke   = "DUEXRPC:INVOKE"
	ErrMessageAssemble = "DUEXRPC:ASSEMBLE"
)

var (
	registryGlobalAlias = map[string]string{
		"id":       "dubbo.registry.id",
		"protocol": "dubbo.registry.protocol",
		"group":    "dubbo.registry.protocol",
		"timeout":  "dubbo.registry.timeout",
		"address":  "dubbo.registry.address",
		"username": "dubbo.registry.username",
		"password": "dubbo.registry.password",
	}
)

// DubboReference配置函数，可外部化配置Dubbo Reference
type OptionFunc func(*flux.Endpoint, *flux.Configuration, *dubgo.ReferenceConfig) *dubgo.ReferenceConfig

// 参数封装函数，可外部化配置为其它协议的值对象
type AssembleFunc func(arguments []flux.Argument) (types []string, values interface{})

// GetRegistryAlias
func GetRegistryGlobalAlias() map[string]string {
	return registryGlobalAlias
}

// SetRegistryGlobalAlias
func SetRegistryGlobalAlias(alias map[string]string) {
	registryGlobalAlias = alias
}

// 集成DubboRPC框架的Exchange
type DubboExchange struct {
	// 可外部配置
	OptionFuncs  []OptionFunc
	AssembleFunc AssembleFunc
	// 内部私有
	traceEnable   bool
	configuration *flux.Configuration
	referenceMu   sync.RWMutex
}

func NewDubboExchange() flux.Exchange {
	return &DubboExchange{
		OptionFuncs:  make([]OptionFunc, 0),
		AssembleFunc: assembleHessianValues,
	}
}

func (ex *DubboExchange) Configuration() *flux.Configuration {
	return ex.configuration
}

func (ex *DubboExchange) Init(config *flux.Configuration) error {
	logger.Info("Dubbo Exchange initializing")
	config.SetDefaults(map[string]interface{}{
		configKeyReferenceDelay: time.Millisecond * 30,
		configKeyTraceEnable:    false,
		"timeout":               "5000",
		"retries":               "0",
		"cluster":               "failover",
		"load-balance":          "random",
		"protocol":              dubbo.DUBBO,
	})
	ex.configuration = config
	ex.traceEnable = config.GetBool(configKeyTraceEnable)
	logger.Infow("Dubbo Exchange request trace", "enable", ex.traceEnable)
	// Set default impl if not present
	if nil == ex.OptionFuncs {
		ex.OptionFuncs = make([]OptionFunc, 0)
	}
	if nil == ex.AssembleFunc {
		ex.AssembleFunc = assembleHessianValues
	}
	// 修改默认Consumer配置
	consumerc := dubgo.GetConsumerConfig()
	// 支持定义Registry
	registry := ex.configuration.Sub("registry")
	registry.SetGlobalAlias(GetRegistryGlobalAlias())
	if id, rconfig := newConsumerRegistry(registry); id != "" && nil != rconfig {
		consumerc.Registries[id] = rconfig
		logger.Infow("Dubbo exchange setup registry", "id", id, "config", rconfig)
	}
	dubgo.SetConsumerConfig(consumerc)
	return nil
}

func (ex *DubboExchange) Startup() error {
	return nil
}

func (ex *DubboExchange) Shutdown(_ context.Context) error {
	dubgo.BeforeShutdown()
	return nil
}

func (ex *DubboExchange) Exchange(ctx flux.Context) *flux.StateError {
	return support.InvokeExchanger(ctx, ex)
}

func (ex *DubboExchange) Invoke(target *flux.Endpoint, fxctx flux.Context) (interface{}, *flux.StateError) {
	types, values := ex.AssembleFunc(target.Arguments)
	// 在测试场景中，fluxContext可能为nil
	attachments := make(map[string]interface{})
	traceId := "no-trace-id"
	if nil != fxctx {
		attachments = fxctx.Attachments()
		traceId = fxctx.RequestId()
	}
	trace := logger.Trace(traceId)
	serviceTag := target.UpstreamUri + "." + target.UpstreamMethod
	if ex.traceEnable {
		trace.Infow("Dubbo invoking",
			"service", serviceTag, "arguments.type", types, "attachments", attachments,
		)
	}
	args := []interface{}{target.UpstreamMethod, types, values}
	service := ex.lookupService(target)
	goctx := context.Background()
	if nil != fxctx {
		// Note: must be map[string]string
		// See: dubbo-go@v1.5.1/common/proxy/proxy.go:150
		ssmap, err := cast.ToStringMapStringE(attachments)
		if nil != err {
			trace.Errorw("Dubbo attachment error", "service", serviceTag, "error", err)
			return nil, &flux.StateError{
				StatusCode: flux.StatusServerError,
				ErrorCode:  flux.ErrorCodeGatewayInternal,
				Message:    ErrMessageAssemble,
				Internal:   err,
			}
		}
		goctx = context.WithValue(goctx, constant.AttachmentKey, ssmap)
	}
	if resp, err := service.Invoke(goctx, args); err != nil {
		trace.Errorw("Dubbo rpc error", "service", serviceTag, "error", err)
		return nil, &flux.StateError{
			StatusCode: flux.StatusBadGateway,
			ErrorCode:  flux.ErrorCodeGatewayExchange,
			Message:    ErrMessageInvoke,
			Internal:   err,
		}
	} else {
		if ex.traceEnable {
			trace.Infow("Dubbo invoked: OK", "service", serviceTag, "return.type", reflect.TypeOf(resp))
		}
		return resp, nil
	}
}

func (ex *DubboExchange) lookupService(endpoint *flux.Endpoint) *dubgo.GenericService {
	ex.referenceMu.Lock()
	defer ex.referenceMu.Unlock()
	id := endpoint.UpstreamUri
	if ref := dubgo.GetConsumerService(id); nil != ref {
		return ref.(*dubgo.GenericService)
	}
	ref := NewReference(id, endpoint, ex.configuration)
	// Options
	const msg = "Dubbo option-func return nil reference"
	for _, opt := range ex.OptionFuncs {
		if nil != opt {
			ref = pkg.RequireNotNil(opt(endpoint, ex.configuration, ref), msg).(*dubgo.ReferenceConfig)
		}
	}
	logger.Infow("Create dubbo reference-config, referring", "service-id", id, "interface", endpoint.UpstreamUri)
	genericService := dubgo.NewGenericService(id)
	dubgo.SetConsumerService(genericService)
	ref.Refer(genericService)
	ref.Implement(genericService)
	t := ex.configuration.GetDuration(configKeyReferenceDelay)
	if t == 0 {
		t = time.Millisecond * 30
	}
	<-time.After(t)
	logger.Infow("Create dubbo reference-config: OK", "service-id", id, "interface", endpoint.UpstreamUri)
	return genericService
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
