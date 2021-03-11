package ext

import (
	"github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-pkg"
)

// 提供一种可扩展的参数查找实现。
// 通过替换参数值查找函数，可以允许某些非规范Http参数系统的自定义参数值查找逻辑。
var (
	argumentLookupFunc flux.ArgumentLookupFunc
)

func SetArgumentLookupFunc(f flux.ArgumentLookupFunc) {
	argumentLookupFunc = fluxpkg.MustNotNil(f, "ArgumentLookupFunc is nil").(flux.ArgumentLookupFunc)
}

func ArgumentLookupFunc() flux.ArgumentLookupFunc {
	return argumentLookupFunc
}

//// 构建参数值对象工具函数

func NewPrimitiveArgument(typeClass, argName string) flux.Argument {
	return NewPrimitiveArgumentWithLoader(typeClass, argName, nil)
}

func NewPrimitiveArgumentWithLoader(typeClass, argName string, valLoader func() flux.MTValue) flux.Argument {
	name := fluxpkg.MustNotEmpty(argName, "argName is empty")
	return flux.Argument{
		Class:         fluxpkg.MustNotEmpty(typeClass, "typeClass is empty"),
		Type:          flux.ArgumentTypePrimitive,
		Name:          name,
		HttpName:      name,
		HttpScope:     flux.ScopeAuto,
		ValueLoader:   valLoader,
		LookupFunc:    ArgumentLookupFunc(),
		ValueResolver: MTValueResolverByType(typeClass),
	}
}

func NewComplexArgument(typeClass, argName string) flux.Argument {
	name := fluxpkg.MustNotEmpty(argName, "argName is empty")
	return flux.Argument{
		Class:         fluxpkg.MustNotEmpty(typeClass, "typeClass is empty"),
		Type:          flux.ArgumentTypeComplex,
		Name:          name,
		HttpName:      name,
		HttpScope:     flux.ScopeAuto,
		LookupFunc:    ArgumentLookupFunc(),
		ValueResolver: MTValueResolverByType(typeClass),
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
		return flux.WrapStringMTValue(value)
	})
}

func NewIntegerArgument(argName string) flux.Argument {
	return NewPrimitiveArgument(flux.JavaLangIntegerClassName, argName)
}

func NewIntegerArgumentWith(argName string, value int32) flux.Argument {
	return NewPrimitiveArgumentWithLoader(flux.JavaLangIntegerClassName, argName, func() flux.MTValue {
		return flux.WrapObjectMTValue(value)
	})
}

func NewLongArgument(argName string) flux.Argument {
	return NewPrimitiveArgument(flux.JavaLangLongClassName, argName)
}

func NewLongArgumentWith(argName string, value int64) flux.Argument {
	return NewPrimitiveArgumentWithLoader(flux.JavaLangLongClassName, argName, func() flux.MTValue {
		return flux.WrapObjectMTValue(value)
	})
}

func NewBooleanArgument(argName string) flux.Argument {
	return NewPrimitiveArgument(flux.JavaLangBooleanClassName, argName)
}

func NewBooleanArgumentWith(argName string, value bool) flux.Argument {
	return NewPrimitiveArgumentWithLoader(flux.JavaLangBooleanClassName, argName, func() flux.MTValue {
		return flux.WrapObjectMTValue(value)
	})
}

func NewFloatArgument(argName string) flux.Argument {
	return NewPrimitiveArgument(flux.JavaLangFloatClassName, argName)
}

func NewFloatArgumentWith(argName string, value float64) flux.Argument {
	return NewPrimitiveArgumentWithLoader(flux.JavaLangFloatClassName, argName, func() flux.MTValue {
		return flux.WrapObjectMTValue(value)
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
