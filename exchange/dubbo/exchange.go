package dubbo

import (
	"context"
	"errors"
	_ "github.com/apache/dubbo-go/cluster/cluster_impl"
	_ "github.com/apache/dubbo-go/cluster/loadbalance"
	"github.com/apache/dubbo-go/common/constant"
	_ "github.com/apache/dubbo-go/common/proxy/proxy_factory"
	dubbogo "github.com/apache/dubbo-go/config"
	_ "github.com/apache/dubbo-go/filter/filter_impl"
	_ "github.com/apache/dubbo-go/registry/nacos"
	_ "github.com/apache/dubbo-go/registry/protocol"
	_ "github.com/apache/dubbo-go/registry/zookeeper"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/internal"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/pkg"
	"github.com/spf13/cast"
	"sync"
	"time"
)

const (
	configKeyTraceEnable    = "trace-enable"
	configKeyReferenceDelay = "reference-delay"
)

var (
	ErrInvalidHeaders = errors.New("DUBBO_RPC:INVALID_HEADERS")
	ErrInvalidStatus  = errors.New("DUBBO_RPC:INVALID_STATUS")
	ErrMessageInvoke  = "DUBBO_RPC:INVOKE"
)

// DubboReference配置函数，可外部化配置Dubbo Reference
type OptionFunc func(*flux.Endpoint, *flux.Configuration, *dubbogo.ReferenceConfig) *dubbogo.ReferenceConfig

// 参数封装函数，可外部化配置为其它协议的值对象
type AssembleFunc func(arguments []flux.Argument) (types []string, values interface{})

// 集成DubboRPC框架的Exchange
type DubboExchange struct {
	// 可外部配置
	OptionFuncs  []OptionFunc
	AssembleFunc AssembleFunc
	// 内部私有
	traceEnable   bool
	configuration *flux.Configuration
	referenceMap  map[string]*dubbogo.ReferenceConfig
	referenceMu   sync.RWMutex
}

func NewDubboExchange() flux.Exchange {
	return &DubboExchange{
		referenceMap: make(map[string]*dubbogo.ReferenceConfig),
		OptionFuncs:  make([]OptionFunc, 0),
		AssembleFunc: assembleHessianValues,
	}
}

func (ex *DubboExchange) Configuration() *flux.Configuration {
	return ex.configuration
}

func (ex *DubboExchange) Init(config *flux.Configuration) error {
	logger.Infof("Dubbo Exchange initializing")
	config.SetDefaults(map[string]interface{}{
		configKeyReferenceDelay: time.Millisecond * 30,
		configKeyTraceEnable:    false,
		"timeout":               "3000",
		"retries":               "1",
		"cluster":               "default",
	})
	ex.configuration = config
	ex.traceEnable = config.GetBool(configKeyTraceEnable)
	logger.Infof("Dubbo Exchange request trace enable: %s", ex.traceEnable)
	if nil == ex.referenceMap {
		ex.referenceMap = make(map[string]*dubbogo.ReferenceConfig)
	}
	// Set default impl if not present
	if nil == ex.OptionFuncs {
		ex.OptionFuncs = make([]OptionFunc, 0)
	}
	if nil == ex.AssembleFunc {
		ex.AssembleFunc = assembleHessianValues
	}
	return nil
}

func (ex *DubboExchange) Exchange(ctx flux.Context) *flux.InvokeError {
	return internal.InvokeExchanger(ctx, ex)
}

func (ex *DubboExchange) Invoke(target *flux.Endpoint, fxctx flux.Context) (interface{}, *flux.InvokeError) {
	types, values := ex.AssembleFunc(target.Arguments)
	// 在测试场景中，fluxContext可能为nil
	attachments := make(map[string]interface{})
	if nil != fxctx {
		attachments = fxctx.Attributes()
	}
	if ex.traceEnable {
		logger.Infof("Dubbo invoke, service:<%s$%s>, value.types: %v, values: %+v, attachments: %v",
			target.UpstreamUri, target.UpstreamMethod, types, values, attachments)
	}
	args := []interface{}{target.UpstreamMethod, types, values}
	reference := ex.lookupReference(target)
	goctx := context.Background()
	if nil != fxctx {
		// Note: must be map[string]string
		// See: dubbo-go@v1.5.1/common/proxy/proxy.go:150
		ssmap, err := cast.ToStringMapStringE(attachments)
		if nil != err {
			return nil, &flux.InvokeError{StatusCode: flux.StatusServerError, Message: ErrMessageInvoke, Internal: err}
		}
		goctx = context.WithValue(goctx, constant.AttachmentKey, ssmap)
	}
	if resp, err := reference.GetRPCService().(*dubbogo.GenericService).Invoke(goctx, args); err != nil {
		logger.Infof("Dubbo rpc error, service: %s, method: %s, err: %s", target.UpstreamUri, target.UpstreamMethod, err)
		return nil, &flux.InvokeError{
			StatusCode: flux.StatusBadGateway,
			Message:    ErrMessageInvoke,
			Internal:   err,
		}
	} else {
		return resp, nil
	}
}

func (ex *DubboExchange) lookupReference(endpoint *flux.Endpoint) *dubbogo.ReferenceConfig {
	ex.referenceMu.Lock()
	defer ex.referenceMu.Unlock()
	interfaceName := endpoint.UpstreamUri
	if ref, ok := ex.referenceMap[interfaceName]; ok {
		return ref
	}
	ref := NewReference(endpoint, ex.configuration)
	// Options
	const msg = "Dubbo option-func return nil reference"
	for _, opt := range ex.OptionFuncs {
		if nil != opt {
			ref = pkg.RequireNotNil(opt(endpoint, ex.configuration, ref), msg).(*dubbogo.ReferenceConfig)
		}
	}
	logger.Infof("Create dubbo reference-config, iface: %s, LOADING", interfaceName)
	ref.GenericLoad(interfaceName)
	t := ex.configuration.GetDuration(configKeyReferenceDelay)
	if t == 0 {
		t = time.Millisecond * 30
	}
	<-time.After(t)
	logger.Infof("Create dubbo reference-config, iface: %s, LOADED OK", interfaceName)
	ex.referenceMap[interfaceName] = ref
	return ref
}
