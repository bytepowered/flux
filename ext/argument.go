package ext

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/pkg"
)

// 提供一种可扩展的参数查找实现。
// 通过替换参数值查找函数，可以允许某些非规范Http参数系统的自定义化参数值查找逻辑。
var (
	_argumentValueLookupFunc  flux.ArgumentValueLookupFunc
	_argumentValueResolveFunc flux.ArgumentValueResolveFunc
)

func SetArgumentValueLookupFunc(r flux.ArgumentValueLookupFunc) {
	_argumentValueLookupFunc = pkg.RequireNotNil(r, "ArgumentValueLookupFunc is nil").(flux.ArgumentValueLookupFunc)
}

func GetArgumentValueLookupFunc() flux.ArgumentValueLookupFunc {
	return _argumentValueLookupFunc
}

func SetArgumentValueResolveFunc(r flux.ArgumentValueResolveFunc) {
	_argumentValueResolveFunc = pkg.RequireNotNil(r, "ArgumentValueResolveFunc is nil").(flux.ArgumentValueResolveFunc)
}

func GetArgumentValueResolveFunc() flux.ArgumentValueResolveFunc {
	return _argumentValueResolveFunc
}

//// 构建参数值对象工具函数

func NewPrimitiveArgument(typeClass, argName string) flux.Argument {
	return flux.Argument{
		TypeClass: typeClass,
		Type:      flux.ArgumentTypePrimitive,
		Name:      argName,
		HttpScope: flux.ScopeAuto,
	}
}

func NewComplexArgument(typeClass, argName string) flux.Argument {
	return flux.Argument{
		TypeClass: typeClass,
		Type:      flux.ArgumentTypeComplex,
		Name:      argName,
	}
}

func NewStringArgument(argName string) flux.Argument {
	return NewPrimitiveArgument(flux.JavaLangStringClassName, argName)
}

func NewIntegerArgument(argName string) flux.Argument {
	return NewPrimitiveArgument(flux.JavaLangIntegerClassName, argName)
}

func NewLongArgument(argName string) flux.Argument {
	return NewPrimitiveArgument(flux.JavaLangLongClassName, argName)
}

func NewBooleanArgument(argName string) flux.Argument {
	return NewPrimitiveArgument(flux.JavaLangBooleanClassName, argName)
}

func NewFloatArgument(argName string) flux.Argument {
	return NewPrimitiveArgument(flux.JavaLangFloatClassName, argName)
}

func NewDoubleArgument(argName string) flux.Argument {
	return NewPrimitiveArgument(flux.JavaLangDoubleClassName, argName)
}

func NewStringMapArgument(argName string) flux.Argument {
	return NewComplexArgument(flux.JavaUtilMapClassName, argName)
}

func NewHashMapArgument(argName string) flux.Argument {
	return NewComplexArgument(flux.JavaUtilMapClassName, argName)
}

func NewSliceArrayArgument(argName string) flux.Argument {
	return NewComplexArgument(flux.JavaUtilListClassName, argName)
}
