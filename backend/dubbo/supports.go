package dubbo

import (
	"context"
	"github.com/apache/dubbo-go-hessian2"
	dubgo "github.com/apache/dubbo-go/config"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/backend"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
)

// Dubbo默认参数封装处理：转换成hession协议对象。
// 注意：不能使用 interface{} 值类型。在Dubbogo 1.5.1 / hessian2 v1.6.1中，序列化值类型会被识别为 Ljava.util.List
// 注意：函数定义的返回值类型不指定为hessian.Object，避免外部化实现或者其它协议实现时，直接依赖hessian.Object类型；
// Ref: dubbo-go-hessian2@v1.7.0/request.go:36
func assembleHessianArguments(args []flux.Argument, ctx flux.Context) ([]string, interface{}, error) {
	size := len(args)
	types := make([]string, size)
	values := make([]hessian.Object, size)
	lookup := ext.GetArgumentValueLookupFunc()
	resolver := ext.GetArgumentValueResolveFunc()
	for i, arg := range args {
		types[i] = arg.Class
		if flux.ArgumentTypePrimitive == arg.Type {
			if value, err := backend.LookupResolveWith(arg, lookup, resolver, ctx); nil != err {
				return nil, nil, err
			} else {
				values[i] = value
			}
		} else if flux.ArgumentTypeComplex == arg.Type {
			if value, err := ComplexToMap(arg, lookup, resolver, ctx); nil != err {
				return nil, nil, err
			} else {
				values[i] = value
			}
		} else {
			logger.TraceContext(ctx).Warnw("Unsupported parameter type", "arg.type", arg.Type)
		}
	}
	return types, values, nil
}

func ComplexToMap(arg flux.Argument, lookup flux.ArgumentValueLookupFunc, resolver flux.ArgumentValueResolveFunc, ctx flux.Context) (map[string]interface{}, error) {
	m := make(map[string]interface{}, 1+len(arg.Fields))
	m["class"] = arg.Class
	for _, field := range arg.Fields {
		if flux.ArgumentTypePrimitive == field.Type {
			if value, err := backend.LookupResolveWith(arg, lookup, resolver, ctx); nil != err {
				return nil, err
			} else {
				m[field.Name] = value
			}
		} else if flux.ArgumentTypeComplex == arg.Type {
			if value, err := ComplexToMap(arg, lookup, resolver, ctx); nil != err {
				return nil, err
			} else {
				m[field.Name] = value
			}
		} else {
			logger.TraceContext(ctx).Warnw("Unsupported parameter type", "arg.type", arg.Type)
		}
	}
	return m, nil
}

func NewReference(refid string, service *flux.BackendService, config *flux.Configuration) *dubgo.ReferenceConfig {
	logger.Infow("Create dubbo reference-config",
		"service", service.Interface, "group", service.RpcGroup, "version", service.RpcVersion)
	ref := dubgo.NewReferenceConfig(refid, context.Background())
	ref.InterfaceName = service.Interface
	ref.Version = service.RpcVersion
	ref.Group = service.RpcGroup
	ref.RequestTimeout = service.RpcTimeout
	ref.Retries = service.RpcRetries
	ref.Cluster = config.GetString("cluster")
	ref.Protocol = config.GetString("protocol")
	ref.Loadbalance = config.GetString("load-balance")
	ref.Generic = true
	return ref
}
