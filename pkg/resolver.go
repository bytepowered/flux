package pkg

import (
	"fmt"
	"github.com/spf13/cast"
)

const (
	defaultResolverName = "default"
)

var (
	_resolvers = make(map[string]ValueResolver)
)

// RegisterValueResolver 添加值类型解析函数
func RegisterValueResolver(valueTypeName string, resolver ValueResolver) {
	_resolvers[valueTypeName] = resolver
}

// GetValueResolver 获取值类型解析函数
func GetValueResolver(valueTypeName string) ValueResolver {
	return _resolvers[valueTypeName]
}

// GetDefaultResolver 获取默认的值类型解析函数
func GetDefaultResolver() ValueResolver {
	return _resolvers[defaultResolverName]
}

var (
	stringResolver = SimpleValueResolver(func(value interface{}) (interface{}, error) {
		return cast.ToString(value), nil
	}).ResolveFunc
	integerResolver = SimpleValueResolver(func(value interface{}) (interface{}, error) {
		return cast.ToIntE(value)
	}).ResolveFunc
	longResolver = SimpleValueResolver(func(value interface{}) (interface{}, error) {
		return cast.ToInt64E(value)
	}).ResolveFunc
	float32Resolver = SimpleValueResolver(func(value interface{}) (interface{}, error) {
		return cast.ToFloat32E(value)
	}).ResolveFunc
	float64Resolver = SimpleValueResolver(func(value interface{}) (interface{}, error) {
		return cast.ToFloat64E(value)
	}).ResolveFunc
	booleanResolver = SimpleValueResolver(func(value interface{}) (interface{}, error) {
		return cast.ToBool(value), nil
	}).ResolveFunc
	mapResolver = SimpleValueResolver(func(value interface{}) (interface{}, error) {
		return _toMap(value)
	}).ResolveFunc
	listResolver = ValueResolver(func(_ string, genericTypes []string, value interface{}) (interface{}, error) {
		return _toList(genericTypes, value)
	})
	defaultResolver = ValueResolver(func(className string, genericTypes []string, value interface{}) (interface{}, error) {
		return map[string]interface{}{
			"class":   className,
			"generic": genericTypes,
			"value":   value,
		}, nil
	})
)

func init() {
	RegisterValueResolver("string", stringResolver)
	RegisterValueResolver("String", stringResolver)
	RegisterValueResolver(JavaLangStringClassName, stringResolver)

	RegisterValueResolver("int", integerResolver)
	RegisterValueResolver("Integer", integerResolver)
	RegisterValueResolver(JavaLangIntegerClassName, integerResolver)

	RegisterValueResolver("int64", longResolver)
	RegisterValueResolver("long", longResolver)
	RegisterValueResolver("Long", longResolver)
	RegisterValueResolver(JavaLangLongClassName, longResolver)

	RegisterValueResolver("float", float32Resolver)
	RegisterValueResolver("Float", float32Resolver)
	RegisterValueResolver(JavaLangFloatClassName, float32Resolver)

	RegisterValueResolver("double", float64Resolver)
	RegisterValueResolver("Double", float64Resolver)
	RegisterValueResolver(JavaLangDoubleClassName, float64Resolver)

	RegisterValueResolver("bool", booleanResolver)
	RegisterValueResolver("Boolean", booleanResolver)
	RegisterValueResolver(JavaLangBooleanClassName, booleanResolver)

	RegisterValueResolver("map", mapResolver)
	RegisterValueResolver("Map", mapResolver)
	RegisterValueResolver(JavaUtilMapClassName, mapResolver)

	RegisterValueResolver("slice", listResolver)
	RegisterValueResolver("List", listResolver)
	RegisterValueResolver(JavaUtilListClassName, listResolver)

	RegisterValueResolver(defaultResolverName, defaultResolver)
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
		v, _ := GetValueResolver(eleTypeName)(eleTypeName, []string{}, value)
		return []interface{}{v}, nil
	} else {
		return []interface{}{value}, nil
	}
}
