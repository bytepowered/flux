package support

import (
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/spf13/cast"
)

var (
	stringResolver = flux.TypedValueResolveWrapper(func(value interface{}) (interface{}, error) {
		return cast.ToString(value), nil
	}).ResolveFunc
	integerResolver = flux.TypedValueResolveWrapper(func(value interface{}) (interface{}, error) {
		return cast.ToInt(value), nil
	}).ResolveFunc
	longResolver = flux.TypedValueResolveWrapper(func(value interface{}) (interface{}, error) {
		return cast.ToInt64(value), nil
	}).ResolveFunc
	float32Resolver = flux.TypedValueResolveWrapper(func(value interface{}) (interface{}, error) {
		return cast.ToFloat32(value), nil
	}).ResolveFunc
	float64Resolver = flux.TypedValueResolveWrapper(func(value interface{}) (interface{}, error) {
		return cast.ToFloat64(value), nil
	}).ResolveFunc
	booleanResolver = flux.TypedValueResolveWrapper(func(value interface{}) (interface{}, error) {
		return cast.ToBool(value), nil
	}).ResolveFunc
	mapResolver = flux.TypedValueResolveWrapper(func(value interface{}) (interface{}, error) {
		return _toMap(value)
	}).ResolveFunc
	listResolver = flux.TypedValueResolver(func(_ string, genericTypes []string, value interface{}) (interface{}, error) {
		return _toList(genericTypes, value)
	})
	defaultResolver = flux.TypedValueResolver(func(className string, genericTypes []string, value interface{}) (interface{}, error) {
		return map[string]interface{}{
			"class":   className,
			"generic": genericTypes,
			"value":   value,
		}, nil
	})
)

func init() {
	ext.RegisterTypedValueResolver("string", stringResolver)
	ext.RegisterTypedValueResolver("String", stringResolver)
	ext.RegisterTypedValueResolver(flux.JavaLangStringClassName, stringResolver)

	ext.RegisterTypedValueResolver("int", integerResolver)
	ext.RegisterTypedValueResolver("Integer", integerResolver)
	ext.RegisterTypedValueResolver(flux.JavaLangIntegerClassName, integerResolver)

	ext.RegisterTypedValueResolver("int64", longResolver)
	ext.RegisterTypedValueResolver("long", longResolver)
	ext.RegisterTypedValueResolver("Long", longResolver)
	ext.RegisterTypedValueResolver(flux.JavaLangLongClassName, longResolver)

	ext.RegisterTypedValueResolver("float", float32Resolver)
	ext.RegisterTypedValueResolver("Float", float32Resolver)
	ext.RegisterTypedValueResolver(flux.JavaLangFloatClassName, float32Resolver)

	ext.RegisterTypedValueResolver("double", float64Resolver)
	ext.RegisterTypedValueResolver("Double", float64Resolver)
	ext.RegisterTypedValueResolver(flux.JavaLangDoubleClassName, float64Resolver)

	ext.RegisterTypedValueResolver("bool", booleanResolver)
	ext.RegisterTypedValueResolver("Boolean", booleanResolver)
	ext.RegisterTypedValueResolver(flux.JavaLangBooleanClassName, booleanResolver)

	ext.RegisterTypedValueResolver("map", mapResolver)
	ext.RegisterTypedValueResolver("Map", mapResolver)
	ext.RegisterTypedValueResolver(flux.JavaUtilMapClassName, mapResolver)

	ext.RegisterTypedValueResolver("slice", listResolver)
	ext.RegisterTypedValueResolver("List", listResolver)
	ext.RegisterTypedValueResolver(flux.JavaUtilListClassName, listResolver)

	ext.RegisterTypedValueResolver(ext.DefaultTypedValueResolverName, defaultResolver)
}

func _toMap(value interface{}) (interface{}, error) {
	if sm, ok := value.(map[string]interface{}); ok {
		return sm, nil
	}
	if om, ok := value.(map[interface{}]interface{}); ok {
		return om, nil
	}
	return nil, fmt.Errorf("输入类型与目标类型map不匹配, input: %+v, type:%T", value, value)
}

func _toList(genericTypes []string, value interface{}) (interface{}, error) {
	if len(genericTypes) > 0 {
		eleTypeName := genericTypes[0]
		v, _ := ext.GetTypedValueResolver(eleTypeName)(eleTypeName, []string{}, value)
		return []interface{}{v}, nil
	} else {
		return []interface{}{value}, nil
	}
}
