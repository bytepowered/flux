package extension

import (
	"github.com/bytepowered/flux/pkg"
	"reflect"
)

const (
	defaultResolverName = "default"
)

var (
	_resolvers   = make(map[string]pkg.ValueResolver)
	_numberTypes = make([]reflect.Type, 0)
)

// RegisterValueResolver 添加值类型解析函数
func RegisterValueResolver(valueTypeName string, resolver pkg.ValueResolver) {
	_resolvers[valueTypeName] = resolver
}

// GetValueResolver 获取值类型解析函数
func GetValueResolver(valueTypeName string) pkg.ValueResolver {
	return _resolvers[valueTypeName]
}

// GetDefaultResolver 获取默认的值类型解析函数
func GetDefaultResolver() pkg.ValueResolver {
	return _resolvers[defaultResolverName]
}

var (
	stringResolver = pkg.ValueResolver(func(_ string, _ []string, value interface{}) (interface{}, error) {
		return pkg.ToString(value), nil
	})
	integerResolver = pkg.ValueResolver(func(_ string, _ []string, value interface{}) (interface{}, error) {
		return pkg.ToInt(value)
	})
	longResolver = pkg.ValueResolver(func(_ string, _ []string, value interface{}) (interface{}, error) {
		return pkg.ToInt64(value)
	})
	float32Resolver = pkg.ValueResolver(func(_ string, _ []string, value interface{}) (interface{}, error) {
		return pkg.ToFloat32(value)
	})
	float64Resolver = pkg.ValueResolver(func(_ string, _ []string, value interface{}) (interface{}, error) {
		return pkg.ToFloat64(value)
	})
	booleanResolver = pkg.ValueResolver(func(_ string, _ []string, value interface{}) (interface{}, error) {
		return pkg.ToBool(value), nil
	})
	mapResolver = pkg.ValueResolver(func(_ string, genericTypes []string, value interface{}) (interface{}, error) {
		return _toMap(value)
	})
	listResolver = pkg.ValueResolver(func(_ string, genericTypes []string, value interface{}) (interface{}, error) {
		return _toList(genericTypes, value)
	})
	defaultResolver = pkg.ValueResolver(func(className string, genericTypes []string, value interface{}) (interface{}, error) {
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
	RegisterValueResolver(pkg.JavaLangStringClassName, stringResolver)

	RegisterValueResolver("int", integerResolver)
	RegisterValueResolver("Integer", integerResolver)
	RegisterValueResolver(pkg.JavaLangIntegerClassName, integerResolver)

	RegisterValueResolver("int64", longResolver)
	RegisterValueResolver("long", longResolver)
	RegisterValueResolver("Long", longResolver)
	RegisterValueResolver(pkg.JavaLangLongClassName, longResolver)

	RegisterValueResolver("float", float32Resolver)
	RegisterValueResolver("Float", float32Resolver)
	RegisterValueResolver(pkg.JavaLangFloatClassName, float32Resolver)

	RegisterValueResolver("double", float64Resolver)
	RegisterValueResolver("Double", float64Resolver)
	RegisterValueResolver(pkg.JavaLangDoubleClassName, float64Resolver)

	RegisterValueResolver("bool", booleanResolver)
	RegisterValueResolver("Boolean", booleanResolver)
	RegisterValueResolver(pkg.JavaLangBooleanClassName, booleanResolver)

	RegisterValueResolver("map", mapResolver)
	RegisterValueResolver("Map", mapResolver)
	RegisterValueResolver(pkg.JavaUtilMapClassName, mapResolver)

	RegisterValueResolver("slice", listResolver)
	RegisterValueResolver("List", listResolver)
	RegisterValueResolver(pkg.JavaUtilListClassName, listResolver)

	RegisterValueResolver(defaultResolverName, defaultResolver)
}

func _toMap(value interface{}) (interface{}, error) {
	if sm, ok := value.(map[string]interface{}); ok {
		return sm, nil
	}
	if om, ok := value.(map[interface{}]interface{}); ok {
		return om, nil
	}
	GetLogger().Warnf("输入类型与目标类型map不匹配, input: %+v, type:%t", value, value)
	return make(map[interface{}]interface{}, 0), nil
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
