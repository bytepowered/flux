package ext

import (
	flux2 "github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-pkg"
)

// 提供一种可扩展的参数查找实现。
// 通过替换参数值查找函数，可以允许某些非规范Http参数系统的自定义参数值查找逻辑。
var (
	argumentLookupFunc flux2.ArgumentLookupFunc
)

func SetArgumentLookupFunc(f flux2.ArgumentLookupFunc) {
	argumentLookupFunc = fluxpkg.MustNotNil(f, "ArgumentLookupFunc is nil").(flux2.ArgumentLookupFunc)
}

func ArgumentLookupFunc() flux2.ArgumentLookupFunc {
	return argumentLookupFunc
}

//// 构建参数值对象工具函数

func NewPrimitiveArgument(typeClass, argName string) flux2.Argument {
	return NewPrimitiveArgumentWithLoader(typeClass, argName, nil)
}

func NewPrimitiveArgumentWithLoader(typeClass, argName string, valLoader func() flux2.MTValue) flux2.Argument {
	name := fluxpkg.MustNotEmpty(argName, "argName is empty")
	return flux2.Argument{
		Class:         fluxpkg.MustNotEmpty(typeClass, "typeClass is empty"),
		Type:          flux2.ArgumentTypePrimitive,
		Name:          name,
		HttpName:      name,
		HttpScope:     flux2.ScopeAuto,
		ValueLoader:   valLoader,
		LookupFunc:    ArgumentLookupFunc(),
		ValueResolver: MTValueResolverByType(typeClass),
	}
}

func NewComplexArgument(typeClass, argName string) flux2.Argument {
	name := fluxpkg.MustNotEmpty(argName, "argName is empty")
	return flux2.Argument{
		Class:         fluxpkg.MustNotEmpty(typeClass, "typeClass is empty"),
		Type:          flux2.ArgumentTypeComplex,
		Name:          name,
		HttpName:      name,
		HttpScope:     flux2.ScopeAuto,
		LookupFunc:    ArgumentLookupFunc(),
		ValueResolver: MTValueResolverByType(typeClass),
	}
}

func NewSliceArrayArgument(argName string, generic string) flux2.Argument {
	arg := NewPrimitiveArgument(flux2.JavaUtilListClassName, argName)
	arg.Generic = []string{generic}
	return arg
}

func NewStringArgument(argName string) flux2.Argument {
	return NewPrimitiveArgument(flux2.JavaLangStringClassName, argName)
}

func NewStringArgumentWith(argName string, value string) flux2.Argument {
	return NewPrimitiveArgumentWithLoader(flux2.JavaLangStringClassName, argName, func() flux2.MTValue {
		return flux2.WrapStringMTValue(value)
	})
}

func NewIntegerArgument(argName string) flux2.Argument {
	return NewPrimitiveArgument(flux2.JavaLangIntegerClassName, argName)
}

func NewIntegerArgumentWith(argName string, value int32) flux2.Argument {
	return NewPrimitiveArgumentWithLoader(flux2.JavaLangIntegerClassName, argName, func() flux2.MTValue {
		return flux2.WrapObjectMTValue(value)
	})
}

func NewLongArgument(argName string) flux2.Argument {
	return NewPrimitiveArgument(flux2.JavaLangLongClassName, argName)
}

func NewLongArgumentWith(argName string, value int64) flux2.Argument {
	return NewPrimitiveArgumentWithLoader(flux2.JavaLangLongClassName, argName, func() flux2.MTValue {
		return flux2.WrapObjectMTValue(value)
	})
}

func NewBooleanArgument(argName string) flux2.Argument {
	return NewPrimitiveArgument(flux2.JavaLangBooleanClassName, argName)
}

func NewBooleanArgumentWith(argName string, value bool) flux2.Argument {
	return NewPrimitiveArgumentWithLoader(flux2.JavaLangBooleanClassName, argName, func() flux2.MTValue {
		return flux2.WrapObjectMTValue(value)
	})
}

func NewFloatArgument(argName string) flux2.Argument {
	return NewPrimitiveArgument(flux2.JavaLangFloatClassName, argName)
}

func NewFloatArgumentWith(argName string, value float64) flux2.Argument {
	return NewPrimitiveArgumentWithLoader(flux2.JavaLangFloatClassName, argName, func() flux2.MTValue {
		return flux2.WrapObjectMTValue(value)
	})
}

func NewDoubleArgument(argName string) flux2.Argument {
	return NewPrimitiveArgument(flux2.JavaLangDoubleClassName, argName)
}

func NewStringMapArgument(argName string) flux2.Argument {
	return NewComplexArgument(flux2.JavaUtilMapClassName, argName)
}

func NewHashMapArgument(argName string) flux2.Argument {
	return NewComplexArgument(flux2.JavaUtilMapClassName, argName)
}
