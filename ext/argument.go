package ext

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/pkg"
)

// 提供一种可扩展的参数查找实现。
// 通过替换参数值查找函数，可以允许某些非规范Http参数系统的自定义化参数值查找逻辑。
var (
	_argumentLookupFunc flux.ArgumentLookupFunc
)

func SetArgumentLookupFunc(fun flux.ArgumentLookupFunc) {
	_argumentLookupFunc = pkg.RequireNotNil(fun, "ArgumentLookupFunc is nil").(flux.ArgumentLookupFunc)
}

func GetArgumentLookupFunc() flux.ArgumentLookupFunc {
	return _argumentLookupFunc
}

//// 构建参数值对象工具函数

func NewPrimitiveArgument(typeClass, argName string, value interface{}) flux.Argument {
	return flux.Argument{
		TypeClass: typeClass,
		Type:      flux.ArgumentTypePrimitive,
		Name:      argName,
		HttpValue: flux.NewWrapValue(value),
	}
}

func NewStringArgument(argName string, value string) flux.Argument {
	return NewPrimitiveArgument(flux.JavaLangStringClassName, argName, value)
}

func NewIntegerArgument(argName string, value int) flux.Argument {
	return NewPrimitiveArgument(flux.JavaLangIntegerClassName, argName, value)
}

func NewLongArgument(argName string, value int64) flux.Argument {
	return NewPrimitiveArgument(flux.JavaLangLongClassName, argName, value)
}

func NewFloatArgument(argName string, value float32) flux.Argument {
	return NewPrimitiveArgument(flux.JavaLangFloatClassName, argName, value)
}

func NewDoubleArgument(argName string, value float64) flux.Argument {
	return NewPrimitiveArgument(flux.JavaLangDoubleClassName, argName, value)
}

func NewHashMapArgument(argName string, value interface{}) flux.Argument {
	return NewPrimitiveArgument(flux.JavaUtilMapClassName, argName, value)
}
