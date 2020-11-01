package ext

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/pkg"
	"reflect"
)

// 提供一种可扩展的参数查找实现。
// 通过替换参数值查找函数，可以允许某些非规范Http参数系统的自定义化参数值查找逻辑。
var (
	_argumentValueLookupFunc flux.ArgumentValueResolver
)

func SetArgumentValueResolver(r flux.ArgumentValueResolver) {
	_argumentValueLookupFunc = pkg.RequireNotNil(r, "ArgumentValueResolver is nil").(flux.ArgumentValueResolver)
}

func GetArgumentValueResolver() flux.ArgumentValueResolver {
	return _argumentValueLookupFunc
}

//// 构建参数值对象工具函数

func NewPrimitiveArgument(typeClass, argName string, value interface{}) flux.Argument {
	return flux.Argument{
		TypeClass: typeClass,
		Type:      flux.ArgumentTypePrimitive,
		Name:      argName,
		Value:     flux.NewWrapValue(value),
	}
}

func NewComplexArgument(typeClass, argName string, value interface{}) flux.Argument {
	return flux.Argument{
		TypeClass: typeClass,
		Type:      flux.ArgumentTypeComplex,
		Name:      argName,
		Value:     flux.NewWrapValue(value),
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

func NewBooleanArgument(argName string, value bool) flux.Argument {
	return NewPrimitiveArgument(flux.JavaLangBooleanClassName, argName, value)
}

func NewFloatArgument(argName string, value float32) flux.Argument {
	return NewPrimitiveArgument(flux.JavaLangFloatClassName, argName, value)
}

func NewDoubleArgument(argName string, value float64) flux.Argument {
	return NewPrimitiveArgument(flux.JavaLangDoubleClassName, argName, value)
}

func NewStringMapArgument(argName string, value map[string]interface{}) flux.Argument {
	return NewComplexArgument(flux.JavaUtilMapClassName, argName, value)
}

func NewHashMapArgument(argName string, value interface{}) flux.Argument {
	// Allow nil
	if nil == value {
		return NewComplexArgument(flux.JavaUtilMapClassName, argName, value)
	}
	switch reflect.TypeOf(value).Kind() {
	case reflect.Map:
		return NewComplexArgument(flux.JavaUtilMapClassName, argName, value)
	default:
		panic("value is not a hashmap")
	}
}

func NewSliceArrayArgument(argName string, value interface{}) flux.Argument {
	// allow nil
	if nil == value {
		return NewComplexArgument(flux.JavaUtilListClassName, argName, value)
	}
	switch reflect.TypeOf(value).Kind() {
	case reflect.Slice, reflect.Array:
		return NewComplexArgument(flux.JavaUtilListClassName, argName, value)
	default:
		panic("value is not a hashmap")
	}
}
