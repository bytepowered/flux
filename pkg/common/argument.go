package common

import (
	"github.com/bytepowered/fluxgo/pkg/ext"
	"github.com/bytepowered/fluxgo/pkg/flux"
	"github.com/bytepowered/fluxgo/pkg/internal"
)

var argumentValueLoaderExtKey = extkey{id: "argument.value.loader.func"}

// ArgumentValueLoaderFunc 参数值直接加载函数
type ArgumentValueLoaderFunc func() flux.ValueObject

// SetArgumentValueLoader 设置参数值加载函数
func SetArgumentValueLoader(arg *flux.ServiceArgumentSpec, f ArgumentValueLoaderFunc) {
	arg.SetExtends(argumentValueLoaderExtKey, f)
}

// ArgumentValueLoader 获取参数值加载函数
func ArgumentValueLoader(arg *flux.ServiceArgumentSpec) (ArgumentValueLoaderFunc, bool) {
	v, ok := arg.GetExtends(argumentValueLoaderExtKey)
	if ok {
		f, is := v.(ArgumentValueLoaderFunc)
		return f, is
	}
	return nil, false
}

// 构建参数值对象工具函数

func NewPrimitiveArgument(typeClass, argName string) flux.ServiceArgumentSpec {
	return NewPrimitiveArgumentWithLoader(typeClass, argName, nil)
}

func NewPrimitiveArgumentWithLoader(typeClass, argName string, valLoader func() flux.ValueObject) flux.ServiceArgumentSpec {
	arg := flux.ServiceArgumentSpec{
		ClassType:  flux.MustNotEmpty(typeClass, "<type-class> in argument MUST NOT empty"),
		StructType: flux.ServiceArgumentTypePrimitive,
		Name:       flux.MustNotEmpty(argName, "<argument-name> in argument MUST NOT empty"),
		HttpName:   argName,
		HttpScope:  flux.ScopeAuto,
	}
	SetArgumentValueLoader(&arg, valLoader)
	return arg
}

func NewComplexArgument(typeClass, argName string) flux.ServiceArgumentSpec {
	return flux.ServiceArgumentSpec{
		ClassType:  flux.MustNotEmpty(typeClass, "<type-class> in argument MUST NOT empty"),
		StructType: flux.ServiceArgumentTypeComplex,
		Name:       flux.MustNotEmpty(argName, "<argument-name> in argument MUST NOT empty"),
		HttpName:   argName,
		HttpScope:  flux.ScopeAuto,
	}
}

func NewSliceArrayArgument(argName string, generic string) flux.ServiceArgumentSpec {
	arg := NewPrimitiveArgument(internal.JavaUtilListClassName, argName)
	arg.GenericTypes = []string{generic}
	return arg
}

func NewStringArgument(argName string) flux.ServiceArgumentSpec {
	return NewPrimitiveArgument(internal.JavaLangStringClassName, argName)
}

func NewStringArgumentWith(argName string, value string) flux.ServiceArgumentSpec {
	return NewPrimitiveArgumentWithLoader(internal.JavaLangStringClassName, argName, func() flux.ValueObject {
		return ext.NewStringValueObject(value)
	})
}

func NewIntegerArgument(argName string) flux.ServiceArgumentSpec {
	return NewPrimitiveArgument(internal.JavaLangIntegerClassName, argName)
}

func NewIntegerArgumentWith(argName string, value int32) flux.ServiceArgumentSpec {
	return NewPrimitiveArgumentWithLoader(internal.JavaLangIntegerClassName, argName, func() flux.ValueObject {
		return ext.NewObjectValueObject(value)
	})
}

func NewLongArgument(argName string) flux.ServiceArgumentSpec {
	return NewPrimitiveArgument(internal.JavaLangLongClassName, argName)
}

func NewLongArgumentWith(argName string, value int64) flux.ServiceArgumentSpec {
	return NewPrimitiveArgumentWithLoader(internal.JavaLangLongClassName, argName, func() flux.ValueObject {
		return ext.NewObjectValueObject(value)
	})
}

func NewBooleanArgument(argName string) flux.ServiceArgumentSpec {
	return NewPrimitiveArgument(internal.JavaLangBooleanClassName, argName)
}

func NewBooleanArgumentWith(argName string, value bool) flux.ServiceArgumentSpec {
	return NewPrimitiveArgumentWithLoader(internal.JavaLangBooleanClassName, argName, func() flux.ValueObject {
		return ext.NewObjectValueObject(value)
	})
}

func NewFloatArgument(argName string) flux.ServiceArgumentSpec {
	return NewPrimitiveArgument(internal.JavaLangFloatClassName, argName)
}

func NewFloatArgumentWith(argName string, value float64) flux.ServiceArgumentSpec {
	return NewPrimitiveArgumentWithLoader(internal.JavaLangFloatClassName, argName, func() flux.ValueObject {
		return ext.NewObjectValueObject(value)
	})
}

func NewDoubleArgument(argName string) flux.ServiceArgumentSpec {
	return NewPrimitiveArgument(internal.JavaLangDoubleClassName, argName)
}

func NewStringMapArgument(argName string) flux.ServiceArgumentSpec {
	return NewComplexArgument(internal.JavaUtilMapClassName, argName)
}

func NewHashMapArgument(argName string) flux.ServiceArgumentSpec {
	return NewComplexArgument(internal.JavaUtilMapClassName, argName)
}
