package dubbo

import (
	hessian "github.com/apache/dubbo-go-hessian2"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/backend"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
)

// Dubbo默认参数封装处理：转换成hession协议对象。
// 注意：不能使用 interface{} 值类型。在Dubbogo 1.5.1 / hessian2 v1.6.1中，序列化值类型会被识别为 Ljava.util.List
// 注意：函数定义的返回值类型不指定为hessian.Object，避免外部化实现或者其它协议实现时，直接依赖hessian.Object类型；
// Ref: dubbo-go-hessian2@v1.7.0/request.go:36
func DefaultArgumentsAssembleFunc(arguments []flux.Argument, ctx flux.Context) ([]string, interface{}, error) {
	size := len(arguments)
	types := make([]string, size)
	values := make([]hessian.Object, size)
	lookup := ext.LoadArgumentValueLookupFunc()
	resolver := ext.LoadArgumentValueResolveFunc()
	for i, argument := range arguments {
		types[i] = argument.Class
		if flux.ArgumentTypePrimitive == argument.Type {
			if value, err := backend.LookupResolveWith(argument, lookup, resolver, ctx); nil != err {
				return nil, nil, err
			} else {
				values[i] = value
			}
		} else if flux.ArgumentTypeComplex == argument.Type {
			if value, err := ArgumentsComplex(argument, lookup, resolver, ctx); nil != err {
				return nil, nil, err
			} else {
				values[i] = value
			}
		} else {
			logger.TraceContext(ctx).Warnw("Unsupported parameter type", "argument-type", argument.Type)
		}
	}
	return types, values, nil
}

func ArgumentsComplex(argument flux.Argument, lookup flux.ArgumentValueLookupFunc, resolver flux.ArgumentValueResolveFunc, ctx flux.Context) (map[string]interface{}, error) {
	m := make(map[string]interface{}, 1+len(argument.Fields))
	m["class"] = argument.Class
	for _, field := range argument.Fields {
		if flux.ArgumentTypePrimitive == field.Type {
			if value, err := backend.LookupResolveWith(field, lookup, resolver, ctx); nil != err {
				return nil, err
			} else {
				m[field.Name] = value
			}
		} else if flux.ArgumentTypeComplex == field.Type {
			if value, err := ArgumentsComplex(field, lookup, resolver, ctx); nil != err {
				return nil, err
			} else {
				m[field.Name] = value
			}
		} else {
			logger.TraceContext(ctx).Warnw("Unsupported parameter type", "argument", argument.Name, "field-type", field.Type)
		}
	}
	return m, nil
}
