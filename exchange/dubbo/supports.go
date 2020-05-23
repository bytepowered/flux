package dubbo

import (
	hessian "github.com/apache/dubbo-go-hessian2"
	dubbogo "github.com/apache/dubbo-go/config"
	"github.com/apache/dubbo-go/protocol/dubbo"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/pkg"
	"time"
)

// DubboReference配置函数，可外部化配置Dubbo Reference
type ReferenceOptionFunc func(*flux.Endpoint, flux.Configuration, *dubbogo.ReferenceConfig) *dubbogo.ReferenceConfig

// 参数封装函数，可外部化配置为其它协议的值对象
type ArgumentAssembleFunc func(arguments []flux.Argument) (types []string, values interface{})

var (
	_refOptionFuncs  = make([]ReferenceOptionFunc, 0)
	_argAssembleFunc ArgumentAssembleFunc
)

// 添加DubboReference配置函数
func AddReferenceOptionFunc(opts ...ReferenceOptionFunc) {
	_refOptionFuncs = append(_refOptionFuncs, opts...)
}

// 外部化配置参数封装函数
func SetArgumentAssembleFunc(f ArgumentAssembleFunc) {
	_argAssembleFunc = f
}

func GetArgumentAssembleFunc() ArgumentAssembleFunc {
	return _argAssembleFunc
}

func init() {
	// 默认将参数值转换为Hession2对象
	SetArgumentAssembleFunc(ArgumentAssembleHessian)
}

func ArgumentAssembleHessian(arguments []flux.Argument) ([]string, interface{}) {
	size := len(arguments)
	types := make([]string, size)
	values := make([]hessian.Object, size)
	for i, arg := range arguments {
		types[i] = arg.TypeClass
		if flux.ArgumentTypePrimitive == arg.Type {
			values[i] = arg.HttpValue.Value()
		} else if flux.ArgumentTypeComplex == arg.Type {
			values[i] = ComplexToMap(arg)
		} else {
			logger.Warnf("Unsupported parameter type: %s", arg.Type)
		}
	}
	return types, values
}

func ComplexToMap(arg flux.Argument) map[string]interface{} {
	m := make(map[string]interface{}, 1+len(arg.Fields))
	m["class"] = arg.TypeClass
	for _, field := range arg.Fields {
		if flux.ArgumentTypePrimitive == field.Type {
			m[field.Name] = field.HttpValue.Value()
		} else if flux.ArgumentTypeComplex == arg.Type {
			m[field.Name] = ComplexToMap(field)
		} else {
			logger.Warnf("Unsupported parameter type: %s", arg.Type)
		}
	}
	return m
}

func NewReference(endpoint *flux.Endpoint, config flux.Configuration) *dubbogo.ReferenceConfig {
	ifaceName := endpoint.UpstreamUri
	timeout := getConfig(endpoint.RpcTimeout, "timeout", config, "3000")
	retries := getConfig(endpoint.RpcRetries, "retries", config, "0")
	cluster := config.GetStringDefault("cluster", "failover")
	logger.Infof("Create dubbo reference-config, iface: %s, PREPARING", ifaceName)
	reference := &dubbogo.ReferenceConfig{
		InterfaceName:  ifaceName,
		Version:        endpoint.RpcVersion,
		Group:          endpoint.RpcGroup,
		RequestTimeout: timeout,
		Cluster:        cluster,
		Retries:        retries,
		Protocol:       dubbo.DUBBO,
		Generic:        true,
	}
	// Options
	for _, optfun := range _refOptionFuncs {
		if nil == optfun {
			continue
		}
		const msg = "Dubbo option-func return nil reference"
		reference = pkg.RequireNotNil(optfun(endpoint, config, reference), msg).(*dubbogo.ReferenceConfig)
	}
	reference.GenericLoad(ifaceName)
	<-time.After(time.Millisecond * 50)
	logger.Infof("Create dubbo reference-config, iface: %s, LOADED OK", ifaceName)
	return reference
}

func getConfig(specifiedValue string, configKey string, config flux.Configuration, defValue string) string {
	if "" != specifiedValue {
		return specifiedValue
	}
	return config.GetStringDefault(configKey, defValue)
}
