package dubbo

import (
	hessian "github.com/apache/dubbo-go-hessian2"
	flux2 "github.com/bytepowered/flux/flux-node"
	"github.com/spf13/cast"
)

// Dubbo默认参数封装处理：转换成hession协议对象。
// 注意：不能使用 interface{} 值类型。在Dubbogo 1.5.1 / hessian2 v1.6.1中，序列化值类型会被识别为 Ljava.util.List
// 注意：函数定义的返回值类型不指定为hessian.Object，避免外部化实现或者其它协议实现时，直接依赖hessian.Object类型；
// Ref: dubbo-go-hessian2@v1.7.0/request.go:36
func DefaultArgAssembleFunc(arguments []flux2.Argument, ctx flux2.Context) ([]string, interface{}, error) {
	size := len(arguments)
	types := make([]string, size)
	outputs := make([]hessian.Object, size)
	for i, arg := range arguments {
		types[i] = arg.Class
		if val, err := arg.Resolve(ctx); nil != err {
			return nil, nil, err
		} else {
			outputs[i] = val
		}
	}
	return types, outputs, nil
}

// DefaultAttAssembleFun 默认实现封装DubboAttachment的函数
func DefaultAttAssembleFun(ctx flux2.Context) (interface{}, error) {
	// Note: must be map[string]string
	// Ref: dubbo-go@v1.5.1/common/proxy/proxy.go:150
	return cast.ToStringMapStringE(ctx.Attributes())
}
