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
	"sync"
)

const (
	ResponseKeyHttpStatus  = "@net.bytepowered.flux.http-status"
	ResponseKeyHttpHeaders = "@net.bytepowered.flux.http-headers"
	ResponseKeyHttpBody    = "@net.bytepowered.flux.http-body"
)

const (
	ExchangeNamespaceDubbo = "EXCHANGE.DUBBO"
)

var (
	ErrInvalidHeaders = errors.New("DUBBO_RPC:INVALID_HEADERS")
	ErrInvalidStatus  = errors.New("DUBBO_RPC:INVALID_STATUS")
	ErrMessageInvoke  = "DUBBO_RPC:INVOKE"
)

var (
	gDecoderConfig DecoderConfig
)

// 定义Decoder的解析响应结构的字段
type DecoderConfig struct {
	KeyCode   string
	KeyHeader string
	KeyBody   string
}

// 集成DubboRPC框架的Exchange
type exchange struct {
	config       flux.Configuration
	traceEnabled bool // 日志打印
	referenceMap map[string]*dubbogo.ReferenceConfig
	referenceMu  sync.RWMutex
}

func NewDubboExchange() flux.Exchange {
	return &exchange{
		referenceMap: make(map[string]*dubbogo.ReferenceConfig),
	}
}

func (ex *exchange) Init() error {
	config := flux.NewNamespaceConfiguration(ExchangeNamespaceDubbo)
	logger.Infof("Dubbo Exchange initializing")
	ex.traceEnabled = config.GetBoolDefault("trace-enable", false)
	gDecoderConfig = DecoderConfig{
		KeyCode:   config.GetStringDefault("decoder-key-code", ResponseKeyHttpStatus),
		KeyHeader: config.GetStringDefault("decoder-key-header", ResponseKeyHttpHeaders),
		KeyBody:   config.GetStringDefault("decoder-key-body", ResponseKeyHttpBody),
	}
	return nil
}

func (ex *exchange) Exchange(ctx flux.Context) *flux.InvokeError {
	return internal.InvokeExchanger(ctx, ex)
}

func (ex *exchange) Invoke(target *flux.Endpoint, fxctx flux.Context) (interface{}, *flux.InvokeError) {
	types, args := assemble(target.Arguments)
	reference := ex.lookup(target)
	goctx := context.Background()
	if nil != fxctx {
		goctx = context.WithValue(goctx, constant.AttachmentKey, pkg.ToStringKVMap(fxctx.AttrValues()))
	}
	if ex.traceEnabled {
		attrs := make(flux.StringMap)
		if fxctx != nil {
			attrs = fxctx.AttrValues()
		}
		logger.Infof("Dubbo invoke, service:<%s$%s>, args.type:[%v], args.value:[%v], attrs: %v",
			target.UpstreamUri, target.UpstreamMethod, types, args, attrs)
	}
	if resp, err := reference.GetRPCService().(*dubbogo.GenericService).
		Invoke(goctx, []interface{}{target.UpstreamMethod, types, args}); err != nil {
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

func (ex *exchange) lookup(endpoint *flux.Endpoint) *dubbogo.ReferenceConfig {
	ex.referenceMu.Lock()
	defer ex.referenceMu.Unlock()
	interfaceName := endpoint.UpstreamUri
	if ref, ok := ex.referenceMap[interfaceName]; ok {
		return ref
	} else {
		newRef := newReference(endpoint, ex.config)
		ex.referenceMap[interfaceName] = newRef
		return newRef
	}
}
