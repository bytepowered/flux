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
		return CastToStringMap(value)
	})
	listResolver = flux.TypedValueResolver(func(_ string, genericTypes []string, value flux.MIMETypeValue) (interface{}, error) {
		return CastToArrayList(genericTypes, value)
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

func CastToStringMap(mimeV flux.MIMETypeValue) (map[string]interface{}, error) {
	switch mimeV.MIMEType {
	case flux.ValueMIMETypeLangStringMap:
		return cast.ToStringMap(mimeV.Value), nil
	case flux.ValueMIMETypeLangText:
		decoder := ext.GetSerializer(ext.TypeNameSerializerJson)
		var hashmap = map[string]interface{}{}
		if err := decoder.Unmarshal([]byte(mimeV.Value.(string)), &hashmap); nil != err {
			return nil, fmt.Errorf("cannot decode text to hashmap, text: %s, error:%w", mimeV.Value, err)
		} else {
			return hashmap, nil
		}
	case flux.ValueMIMETypeLangObject:
		if sm, err := cast.ToStringMapE(mimeV.Value); nil != err {
			return nil, fmt.Errorf("cannot cast object to hashmap, object: %+v, object.type:%T", mimeV.Value, mimeV.Value)
		} else {
			return sm, nil
		}
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
			} else if jbs, err := JSONBytesFromQueryString(bs); nil != err {
				return nil, err
			} else {
				data = jbs
			}
		} else {
			if sm, err := cast.ToStringMapE(mimeV.Value); nil == err {
				return sm, nil
			} else {
				return nil, fmt.Errorf("unsupported mime-type to hashmap, value: %+v, value.type:%T, mime-type: %s",
					mimeV.Value, mimeV.Value, mimeV.MIMEType)
			}
		}
		decoder := ext.GetSerializer(ext.TypeNameSerializerJson)
		var hashmap = map[string]interface{}{}
		err := decoder.Unmarshal(data, &hashmap)
		return hashmap, err
	}
}

func CastToArrayList(genericTypes []string, mimeV flux.MIMETypeValue) ([]interface{}, error) {
	// SingleValue to arraylist
	if len(genericTypes) > 0 {
		typeClass := genericTypes[0]
		resolver := ext.GetTypedValueResolver(typeClass)
		if v, err := resolver(typeClass, []string{}, mimeV); nil != err {
			return nil, err
		} else {
			return []interface{}{v}, nil
		}
	} else {
		return []interface{}{mimeV.Value}, nil
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
func JSONBytesFromQueryString(queryStr []byte) ([]byte, error) {
	queryValues, err := url.ParseQuery(string(queryStr))
	if nil != err {
		return nil, err
	}
	fields := make([]string, 0, len(queryValues))
	for key, values := range queryValues {
		if len(values) > 1 {
			// quote with ""
			copied := make([]string, len(values))
			for i, val := range values {
				copied[i] = "\"" + string(JSONStringValueEncode(&val)) + "\""
			}
			fields = append(fields, "\""+key+"\":["+strings.Join(copied, ",")+"]")
		} else {
			fields = append(fields, "\""+key+"\":\""+string(JSONStringValueEncode(&values[0]))+"\"")
		}
	}
	bf := new(bytes.Buffer)
	bf.WriteByte('{')
	bf.WriteString(strings.Join(fields, ","))
	bf.WriteByte('}')
	return bf.Bytes(), nil
}

func JSONStringValueEncode(str *string) []byte {
	return []byte(strings.Replace(*str, `"`, `\"`, -1))
}
