package ext

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/pkg"
)

func NewPrimitiveArgument(typeClass, argName string, value interface{}) flux.Argument {
	return flux.Argument{
		TypeClass: typeClass,
		ArgType:   flux.ArgumentTypePrimitive,
		ArgName:   argName,
		ArgValue:  flux.NewWrapValue(value),
	}
}

func NewStringArgument(argName string, value string) flux.Argument {
	return NewPrimitiveArgument(pkg.JavaLangStringClassName, argName, value)
}

func NewIntegerArgument(argName string, value int) flux.Argument {
	return NewPrimitiveArgument(pkg.JavaLangIntegerClassName, argName, value)
}

func NewLongArgument(argName string, value int64) flux.Argument {
	return NewPrimitiveArgument(pkg.JavaLangLongClassName, argName, value)
}

func NewFloatArgument(argName string, value float32) flux.Argument {
	return NewPrimitiveArgument(pkg.JavaLangFloatClassName, argName, value)
}

func NewDoubleArgument(argName string, value float64) flux.Argument {
	return NewPrimitiveArgument(pkg.JavaLangDoubleClassName, argName, value)
}

func NewHashMapArgument(argName string, value interface{}) flux.Argument {
	return NewPrimitiveArgument(pkg.JavaUtilMapClassName, argName, value)
}
