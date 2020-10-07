package support

import (
	"bytes"
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/spf13/cast"
	"io"
	"io/ioutil"
	"net/url"
	"strings"
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
	mapResolver = flux.TypedValueResolver(func(_ string, genericTypes []string, value flux.MIMETypeValue) (interface{}, error) {
		return _toHashMap(value)
	})
	listResolver = flux.TypedValueResolver(func(_ string, genericTypes []string, value flux.MIMETypeValue) (interface{}, error) {
		return _toArrayList(genericTypes, value)
	})
	defaultResolver = flux.TypedValueResolver(func(typeClass string, typeGeneric []string, value flux.MIMETypeValue) (interface{}, error) {
		return map[string]interface{}{
			"class":   typeClass,
			"generic": typeGeneric,
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

// FIXME need test
func _toHashMap(mimeV flux.MIMETypeValue) (interface{}, error) {
	switch mimeV.MIMEType {
	case flux.ValueMIMETypeLangStringMap:
		return mimeV.Value, nil
	case flux.ValueMIMETypeLangText:
		decoder := ext.GetSerializer(ext.TypeNameSerializerJson)
		var hashmap = map[string]interface{}{}
		err := decoder.Unmarshal([]byte(mimeV.Value.(string)), &hashmap)
		return hashmap, fmt.Errorf("cannot decode text to hashmap, text: %s, error:%w", mimeV.Value, err)
	case flux.ValueMIMETypeLangObject:
		if sm, ok := mimeV.Value.(map[string]interface{}); ok {
			return sm, nil
		}
		if om, ok := mimeV.Value.(map[interface{}]interface{}); ok {
			return om, nil
		}
		return nil, fmt.Errorf("cannot cast object to hashmap, object: %+v, object.type:%T", mimeV.Value, mimeV.Value)
	default:
		var data []byte
		if strings.Contains(mimeV.MIMEType, "application/json") {
			if bs, err := _toBytes(mimeV.Value); nil != err {
				return nil, err
			} else {
				data = bs
			}
		} else if strings.Contains(mimeV.MIMEType, "application/x-www-form-urlencoded") {
			if bs, err := _toBytes(mimeV.Value); nil != err {
				return nil, err
			} else if jbs, err := _queryToJsonBytes(bs); nil != err {
				return nil, err
			} else {
				data = jbs
			}
		} else {
			return nil, fmt.Errorf("unsupported mime-type to hashmap, value: %+v, value.type:%T, mime-type: %s",
				mimeV.Value, mimeV.Value, mimeV.MIMEType)
		}
		decoder := ext.GetSerializer(ext.TypeNameSerializerJson)
		var hashmap = map[string]interface{}{}
		err := decoder.Unmarshal(data, &hashmap)
		return hashmap, err
	}
}

// FIXME need test
func _toArrayList(genericTypes []string, mimeV flux.MIMETypeValue) (interface{}, error) {
	if len(genericTypes) > 0 {
		elementType := genericTypes[0]
		resolver := ext.GetTypedValueResolver(elementType)
		if v, err := resolver(elementType, []string{}, mimeV); nil != err {
			return nil, err
		} else {
			return []interface{}{v}, nil
		}
	} else {
		return []interface{}{mimeV}, nil
	}
}

func _toBytes(v interface{}) ([]byte, error) {
	switch v.(type) {
	case []byte:
		return v.([]byte), nil
	case string:
		return []byte(v.(string)), nil
	case io.Reader:
		bs, err := ioutil.ReadAll(v.(io.Reader))
		if closer, ok := v.(io.Closer); ok {
			_ = closer.Close()
		}
		if nil != err {
			return nil, err
		} else {
			return bs, nil
		}
	default:
		return nil, fmt.Errorf("cannot convert value to []byte, value: %+v, value.type:%T", v, v)
	}
}

// Tested
func _queryToJsonBytes(queryStr []byte) ([]byte, error) {
	query, err := url.ParseQuery(string(queryStr))
	if nil != err {
		return nil, err
	}
	fields := make([]string, 0, len(query))
	for key, values := range query {
		if len(values) > 1 {
			// wrap with ""
			copied := make([]string, len(values))
			for i, val := range values {
				copied[i] = `"` + strings.Replace(val, "\"", "\\\"", -1) + `"`
			}
			fields = append(fields, fmt.Sprintf(`"%s":[%s]`, key, strings.Join(copied, ",")))
		} else {
			fields = append(fields, fmt.Sprintf(`"%s":"%s"`, key, strings.Replace(values[0], "\"", "\\\"", -1)))
		}
	}
	bf := new(bytes.Buffer)
	bf.WriteByte('{')
	bf.WriteString(strings.Join(fields, ","))
	bf.WriteByte('}')
	return bf.Bytes(), nil
}
