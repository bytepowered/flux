package dubbo

import (
	"context"
	"github.com/apache/dubbo-go-hessian2"
	dubgo "github.com/apache/dubbo-go/config"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/logger"
	"go.uber.org/zap"
)

// Dubbo默认参数封装处理：转换成hession协议对象。
// 注意：不能使用 interface{} 值类型。在Dubbogo 1.5.1 / hessian2 v1.6.1中，序列化值类型会被识别为 Ljava.util.List
// 注意：函数定义的返回值类型不指定为hessian.Object，避免外部化实现或者其它协议实现时，直接依赖hessian.Object类型；
func assembleHessianValues(arguments []flux.Argument) ([]string, interface{}) {
	size := len(arguments)
	argTypes := make([]string, size)
	argValues := make([]hessian.Object, size)
	for i, arg := range arguments {
		argTypes[i] = arg.TypeClass
		if flux.ArgumentTypePrimitive == arg.Type {
			argValues[i] = arg.Value.Get()
		} else if flux.ArgumentTypeComplex == arg.Type {
			argValues[i] = ComplexToMap(arg)
		} else {
			logger.Warn("Unsupported parameter", zap.String("type", arg.Type))
		}
	}
	return argTypes, argValues
}

func ComplexToMap(arg flux.Argument) map[string]interface{} {
	m := make(map[string]interface{}, 1+len(arg.Fields))
	m["class"] = arg.TypeClass
	for _, field := range arg.Fields {
		if flux.ArgumentTypePrimitive == field.Type {
			m[field.Name] = field.Value.Get()
		} else if flux.ArgumentTypeComplex == arg.Type {
			m[field.Name] = ComplexToMap(field)
		} else {
			logger.Warn("Unsupported parameter", zap.String("type", arg.Type))
		}
	}
	return m
}

func NewReference(refid string, service *flux.Service, config *flux.Configuration) *dubgo.ReferenceConfig {
	logger.Infow("Create dubbo reference-config",
		"service", service.Interface, "group", service.Group, "version", service.Version)
	ref := dubgo.NewReferenceConfig(refid, context.Background())
	ref.InterfaceName = service.Interface
	ref.Version = service.Version
	ref.Group = service.Group
	ref.RequestTimeout = service.Timeout
	ref.Retries = service.Retries
	ref.Cluster = config.GetString("cluster")
	ref.Protocol = config.GetString("protocol")
	ref.Loadbalance = config.GetString("load-balance")
	ref.Generic = true
	return ref
}
