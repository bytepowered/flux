package ext

import (
	"github.com/bytepowered/flux"
)

// 提供一种可扩展的参数查找实现。
// 通过替换参数值查找函数，可以允许某些非规范Http参数系统的自定义参数值查找逻辑。
var (
	lookupFunc flux.MTValueLookupFunc
)

func SetLookupFunc(f flux.MTValueLookupFunc) {
	lookupFunc = flux.MustNotNil(f, "MTValueLookupFunc is nil").(flux.MTValueLookupFunc)
}

func LookupFunc() flux.MTValueLookupFunc {
	return lookupFunc
}

//// 构建参数值对象工具函数

func NewPrimitiveArgument(typeClass, argName string) flux.Argument {
	return NewPrimitiveArgumentWithLoader(typeClass, argName, nil)
}

func NewPrimitiveArgumentWithLoader(typeClass, argName string, valLoader func() flux.MTValue) flux.Argument {
	arg := flux.Argument{
		Class:     flux.MustNotEmpty(typeClass, "<type-class> in argument MUST NOT empty"),
		Type:      flux.ArgumentTypePrimitive,
		Name:      flux.MustNotEmpty(argName, "<type-class> in argument MUST NOT empty"),
		HttpName:  argName,
		HttpScope: flux.ScopeAuto,
	}
	SetArgumentValueLoader(&arg, valLoader)
	return arg
}

func NewComplexArgument(typeClass, argName string) flux.Argument {
	return flux.Argument{
		Class:     flux.MustNotEmpty(typeClass, "<type-class> in argument MUST NOT empty"),
		Type:      flux.ArgumentTypeComplex,
		Name:      flux.MustNotEmpty(argName, "<type-class> in argument MUST NOT empty"),
		HttpName:  argName,
		HttpScope: flux.ScopeAuto,
	}
}

func NewSliceArrayArgument(argName string, generic string) flux.Argument {
	arg := NewPrimitiveArgument(flux.JavaUtilListClassName, argName)
	arg.Generic = []string{generic}
	return arg
}

func NewStringArgument(argName string) flux.Argument {
	return NewPrimitiveArgument(flux.JavaLangStringClassName, argName)
}

func NewStringArgumentWith(argName string, value string) flux.Argument {
	return NewPrimitiveArgumentWithLoader(flux.JavaLangStringClassName, argName, func() flux.MTValue {
		return flux.NewStringMTValue(value)
	})
}

func NewIntegerArgument(argName string) flux.Argument {
	return NewPrimitiveArgument(flux.JavaLangIntegerClassName, argName)
}

func NewIntegerArgumentWith(argName string, value int32) flux.Argument {
	return NewPrimitiveArgumentWithLoader(flux.JavaLangIntegerClassName, argName, func() flux.MTValue {
		return flux.NewObjectMTValue(value)
	})
}

func NewLongArgument(argName string) flux.Argument {
	return NewPrimitiveArgument(flux.JavaLangLongClassName, argName)
}

func NewLongArgumentWith(argName string, value int64) flux.Argument {
	return NewPrimitiveArgumentWithLoader(flux.JavaLangLongClassName, argName, func() flux.MTValue {
		return flux.NewObjectMTValue(value)
	})
}

func NewBooleanArgument(argName string) flux.Argument {
	return NewPrimitiveArgument(flux.JavaLangBooleanClassName, argName)
}

func NewBooleanArgumentWith(argName string, value bool) flux.Argument {
	return NewPrimitiveArgumentWithLoader(flux.JavaLangBooleanClassName, argName, func() flux.MTValue {
		return flux.NewObjectMTValue(value)
	})
}

func NewFloatArgument(argName string) flux.Argument {
	return NewPrimitiveArgument(flux.JavaLangFloatClassName, argName)
}

func NewFloatArgumentWith(argName string, value float64) flux.Argument {
	return NewPrimitiveArgumentWithLoader(flux.JavaLangFloatClassName, argName, func() flux.MTValue {
		return flux.NewObjectMTValue(value)
	})
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
