package support

import (
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/spf13/cast"
	"io"
	"io/ioutil"
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
		return _toHashMap(value)
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
	ext.SetTypedValueResolver("string", stringResolver)
	ext.SetTypedValueResolver("String", stringResolver)
	ext.SetTypedValueResolver(flux.JavaLangStringClassName, stringResolver)

	ext.SetTypedValueResolver("int", integerResolver)
	ext.SetTypedValueResolver("Integer", integerResolver)
	ext.SetTypedValueResolver(flux.JavaLangIntegerClassName, integerResolver)

	ext.SetTypedValueResolver("int64", longResolver)
	ext.SetTypedValueResolver("long", longResolver)
	ext.SetTypedValueResolver("Long", longResolver)
	ext.SetTypedValueResolver(flux.JavaLangLongClassName, longResolver)

	ext.SetTypedValueResolver("float", float32Resolver)
	ext.SetTypedValueResolver("Float", float32Resolver)
	ext.SetTypedValueResolver(flux.JavaLangFloatClassName, float32Resolver)

	ext.SetTypedValueResolver("double", float64Resolver)
	ext.SetTypedValueResolver("Double", float64Resolver)
	ext.SetTypedValueResolver(flux.JavaLangDoubleClassName, float64Resolver)

	ext.SetTypedValueResolver("bool", booleanResolver)
	ext.SetTypedValueResolver("Boolean", booleanResolver)
	ext.SetTypedValueResolver(flux.JavaLangBooleanClassName, booleanResolver)

	ext.SetTypedValueResolver("map", mapResolver)
	ext.SetTypedValueResolver("Map", mapResolver)
	ext.SetTypedValueResolver(flux.JavaUtilMapClassName, mapResolver)

	ext.SetTypedValueResolver("slice", listResolver)
	ext.SetTypedValueResolver("List", listResolver)
	ext.SetTypedValueResolver(flux.JavaUtilListClassName, listResolver)

	ext.SetTypedValueResolver(ext.DefaultTypedValueResolverName, defaultResolver)
}

func _toHashMap(value interface{}) (interface{}, error) {
	if sm, ok := value.(map[string]interface{}); ok {
		return sm, nil
	}
	if om, ok := value.(map[interface{}]interface{}); ok {
		return om, nil
	}
	decoder := ext.GetSerializer(ext.TypeNameSerializerJson)
	var hashmap = map[string]interface{}{}
	switch v := value.(type) {
	// JSON String
	case string:
		err := decoder.Unmarshal([]byte(v), &hashmap)
		return hashmap, err
	// JSON Bytes
	case []byte:
		err := decoder.Unmarshal(v, &hashmap)
		return hashmap, err
	// JSON Body Reader
	case io.Reader:
		data, err := ioutil.ReadAll(v)
		if c, ok := v.(io.Closer); ok {
			_ = c.Close()
		}
		if nil != err {
			return hashmap, err
		} else {
			err = decoder.Unmarshal(data, &hashmap)
			return hashmap, err
		}
	default:
		return hashmap, fmt.Errorf("输入类型与目标类型map不匹配, input: %+v, type:%T", value, value)
	}
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
