package backend

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-node/ext"
	"github.com/spf13/cast"
	"io"
	"io/ioutil"
	"net/url"
	"reflect"
	"strings"
)

var (
	errCastToByteTypeNotSupported = errors.New("cannot convert value to []byte")
)

var (
	stringResolver = flux.MTValueResolver(func(mtValue flux.MTValue, _ string, genericTypes []string) (interface{}, error) {
		return CastDecodeMTValueToString(mtValue)
	})
	integerResolver = flux.WrapMTValueResolver(func(value interface{}) (interface{}, error) {
		if isEmptyOrNil(value) {
			return int(0), nil
		}
		return cast.ToIntE(value)
	}).ResolveMT
	longResolver = flux.WrapMTValueResolver(func(value interface{}) (interface{}, error) {
		if isEmptyOrNil(value) {
			return int64(0), nil
		}
		return cast.ToInt64E(value)
	}).ResolveMT
	float32Resolver = flux.WrapMTValueResolver(func(value interface{}) (interface{}, error) {
		if isEmptyOrNil(value) {
			return float32(0), nil
		}
		return cast.ToFloat32E(value)
	}).ResolveMT
	float64Resolver = flux.WrapMTValueResolver(func(value interface{}) (interface{}, error) {
		if isEmptyOrNil(value) {
			return float64(0), nil
		}
		return cast.ToFloat64E(value)
	}).ResolveMT
	booleanResolver = flux.WrapMTValueResolver(func(value interface{}) (interface{}, error) {
		if isEmptyOrNil(value) {
			return false, nil
		}
		return cast.ToBoolE(value)
	}).ResolveMT
	mapResolver = flux.MTValueResolver(func(value flux.MTValue, _ string, genericTypes []string) (interface{}, error) {
		return ToStringMapE(value)
	})
	listResolver = flux.MTValueResolver(func(value flux.MTValue, _ string, genericTypes []string) (interface{}, error) {
		return ToGenericListE(genericTypes, value)
	})
	complexObjectResolver = flux.MTValueResolver(func(mtValue flux.MTValue, class string, generic []string) (interface{}, error) {
		if isEmptyOrNil(mtValue.Value) {
			return map[string]interface{}{"class": class}, nil
		}
		sm, err := ToStringMapE(mtValue)
		sm["class"] = class
		if nil != err {
			return nil, err
		}
		return sm, nil
	})
)

func init() {
	ext.RegisterMTValueResolver("string", stringResolver)
	ext.RegisterMTValueResolver(flux.JavaLangStringClassName, stringResolver)

	ext.RegisterMTValueResolver("int", integerResolver)
	ext.RegisterMTValueResolver(flux.JavaLangIntegerClassName, integerResolver)

	ext.RegisterMTValueResolver("int64", longResolver)
	ext.RegisterMTValueResolver("long", longResolver)
	ext.RegisterMTValueResolver(flux.JavaLangLongClassName, longResolver)

	ext.RegisterMTValueResolver("float", float32Resolver)
	ext.RegisterMTValueResolver("float32", float32Resolver)
	ext.RegisterMTValueResolver(flux.JavaLangFloatClassName, float32Resolver)

	ext.RegisterMTValueResolver("float64", float64Resolver)
	ext.RegisterMTValueResolver("double", float64Resolver)
	ext.RegisterMTValueResolver(flux.JavaLangDoubleClassName, float64Resolver)

	ext.RegisterMTValueResolver("bool", booleanResolver)
	ext.RegisterMTValueResolver("boolean", booleanResolver)
	ext.RegisterMTValueResolver(flux.JavaLangBooleanClassName, booleanResolver)

	ext.RegisterMTValueResolver("map", mapResolver)
	ext.RegisterMTValueResolver(flux.JavaUtilMapClassName, mapResolver)

	ext.RegisterMTValueResolver("slice", listResolver)
	ext.RegisterMTValueResolver("list", listResolver)
	ext.RegisterMTValueResolver(flux.JavaUtilListClassName, listResolver)

	ext.RegisterMTValueResolver(ext.DefaultMTValueResolverName, complexObjectResolver)
}

// CastDecodeToString 最大努力地将值转换成String类型。
// 如果类型无法安全地转换成String或者解析异常，返回错误。
func CastDecodeMTValueToString(mtValue flux.MTValue) (string, error) {
	if isEmptyOrNil(mtValue.Value) {
		return "", nil
	}
	// 可直接转String类型：
	if str, err := cast.ToStringE(mtValue.Value); nil == err {
		return str, nil
	}
	if data, err := toByteArray0(mtValue.Value); nil == err {
		return string(data), nil
	} else if err != errCastToByteTypeNotSupported {
		return "", err
	}
	if data, err := ext.JSONMarshal(mtValue.Value); nil != err {
		return "", err
	} else {
		return string(data), nil
	}
}

// ToStringMapE 最大努力地将值转换成map[string]any类型。
// 如果类型无法安全地转换成map[string]any或者解析异常，返回错误。
func ToStringMapE(mtValue flux.MTValue) (map[string]interface{}, error) {
	if isEmptyOrNil(mtValue.Value) {
		return make(map[string]interface{}, 0), nil
	}
	switch mtValue.MediaType {
	case flux.ValueMediaTypeGoStringMap:
		return cast.ToStringMap(mtValue.Value), nil
	case flux.ValueMediaTypeGoString:
		var hashmap = map[string]interface{}{}
		if err := ext.JSONUnmarshal([]byte(mtValue.Value.(string)), &hashmap); nil != err {
			return nil, fmt.Errorf("cannot decode text to hashmap, text: %s, error:%w", mtValue.Value, err)
		} else {
			return hashmap, nil
		}
	case flux.ValueMediaTypeGoObject:
		if sm, err := cast.ToStringMapE(mtValue.Value); nil != err {
			return nil, fmt.Errorf("cannot cast object to hashmap, object: %+v, object.type:%T", mtValue.Value, mtValue.Value)
		} else {
			return sm, nil
		}
	default:
		var data []byte
		if strings.Contains(mtValue.MediaType, "application/json") {
			if bs, err := toByteArray(mtValue.Value); nil != err {
				return nil, err
			} else {
				data = bs
			}
		} else if strings.Contains(mtValue.MediaType, "application/x-www-form-urlencoded") {
			if bs, err := toByteArray(mtValue.Value); nil != err {
				return nil, err
			} else if jbs, err := JSONBytesFromQueryString(bs); nil != err {
				return nil, err
			} else {
				data = jbs
			}
		} else {
			if sm, err := cast.ToStringMapE(mtValue.Value); nil == err {
				return sm, nil
			} else {
				return nil, fmt.Errorf("unsupported mime-type to hashmap, value: %+v, value.type:%T, mime-type: %s",
					mtValue.Value, mtValue.Value, mtValue.MediaType)
			}
		}
		var hashmap = map[string]interface{}{}
		err := ext.JSONUnmarshal(data, &hashmap)
		return hashmap, err
	}
}

// ToGenericListE 最大努力地将值转换成[]any类型。
// 如果类型无法安全地转换成[]any或者解析异常，返回错误。
func ToGenericListE(generics []string, mtValue flux.MTValue) (interface{}, error) {
	if isEmptyOrNil(mtValue.Value) {
		return make([]interface{}, 0), nil
	}
	vType := reflect.TypeOf(mtValue.Value)
	// 没有指定泛型类型
	if len(generics) == 0 {
		return []interface{}{mtValue.Value}, nil
	}
	// 进行特定泛型类型转换
	generic := generics[0]
	resolver := ext.MTValueResolverByType(generic)
	kind := vType.Kind()
	if kind == reflect.Slice {
		vValue := reflect.ValueOf(mtValue.Value)
		out := make([]interface{}, vValue.Len())
		for i := 0; i < vValue.Len(); i++ {
			if v, err := resolver(flux.WrapObjectMTValue(vValue.Index(i).Interface()), generic, []string{}); nil != err {
				return nil, err
			} else {
				out[i] = v
			}
		}
		return out, nil
	}
	if v, err := resolver(mtValue, generic, []string{}); nil != err {
		return nil, err
	} else {
		return []interface{}{v}, nil
	}
}

func toByteArray(v interface{}) ([]byte, error) {
	if bs, err := toByteArray0(v); nil != err {
		return nil, fmt.Errorf("value: %+v, value.type:%T, error: %w", v, v, err)
	} else {
		return bs, nil
	}
}

func toByteArray0(v interface{}) ([]byte, error) {
	switch v.(type) {
	case []byte:
		return v.([]byte), nil
	case string:
		return []byte(v.(string)), nil
	case io.Reader:
		data, err := ioutil.ReadAll(v.(io.Reader))
		if closer, ok := v.(io.Closer); ok {
			_ = closer.Close()
		}
		if nil != err {
			return nil, err
		} else {
			return data, nil
		}
	default:
		return nil, errCastToByteTypeNotSupported
	}
}

func isEmptyOrNil(v interface{}) bool {
	if s, ok := v.(string); ok {
		return "" == s
	}
	return nil == v
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
