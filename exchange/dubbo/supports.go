package dubbo

import (
	hessian "github.com/apache/dubbo-go-hessian2"
	dubbogo "github.com/apache/dubbo-go/config"
	"github.com/apache/dubbo-go/protocol/dubbo"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/logger"
)

// Dubbo默认参数封装处理：转换成hession协议对象。
// 注意：不能使用 interface{} 值类型。在Dubbogo 1.5.1 / hessian2 v1.6.1中，序列化值类型会被识别为 Ljava.util.List
func assembleHessianValues(arguments []flux.Argument) ([]string, interface{}) {
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

func NewReference(endpoint *flux.Endpoint, config *flux.Configuration) *dubbogo.ReferenceConfig {
	ifaceName := endpoint.UpstreamUri
	logger.Infof("Create dubbo reference-config, iface: %s, PREPARING", ifaceName)
	reference := &dubbogo.ReferenceConfig{
		InterfaceName:  ifaceName,
		Version:        endpoint.RpcVersion,
		Group:          endpoint.RpcGroup,
		RequestTimeout: config.GetString("timeout"),
		Cluster:        config.GetString("cluster"),
		Retries:        config.GetString("retries"),
		Protocol:       dubbo.DUBBO,
		Generic:        true,
	}
	return reference
}
