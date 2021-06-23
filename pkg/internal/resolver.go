package internal

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/spf13/cast"
	"io"
	"io/ioutil"
	"net/url"
	"reflect"
	"strings"
)

import (
	"github.com/bytepowered/fluxgo/pkg/ext"
	"github.com/bytepowered/fluxgo/pkg/flux"
)

const (
	JavaLangObjectClassName  = "java.lang.Object"
	JavaLangStringClassName  = "java.lang.String"
	JavaLangIntegerClassName = "java.lang.Integer"
	JavaLangLongClassName    = "java.lang.Long"
	JavaLangFloatClassName   = "java.lang.Float"
	JavaLangDoubleClassName  = "java.lang.Double"
	JavaLangBooleanClassName = "java.lang.Boolean"
	JavaUtilMapClassName     = "java.util.Map"
	JavaUtilListClassName    = "java.util.List"
	JavaIOSerializable       = "java.io.Serializable"
)

var (
	errCastToByteTypeNotSupported = errors.New("cannot convert value to []byte")
)

var (
	objectResolver = ext.ValueObjectResolver(func(mtValue flux.ValueObject, _ string, genericTypes []string) (interface{}, error) {
		return mtValue.Value, nil
	})
	stringResolver = ext.ValueObjectResolver(func(mtValue flux.ValueObject, _ string, genericTypes []string) (interface{}, error) {
		return CastDecodeMTValueToString(mtValue)
	})
	integerResolver = ext.WrapValueObjectResolver(func(value interface{}) (interface{}, error) {
		if isEmptyOrNil(value) {
			return int(0), nil
		}
		return cast.ToIntE(value)
	}).ResolveTo
	longResolver = ext.WrapValueObjectResolver(func(value interface{}) (interface{}, error) {
		if isEmptyOrNil(value) {
			return int64(0), nil
		}
		return cast.ToInt64E(value)
	}).ResolveTo
	float32Resolver = ext.WrapValueObjectResolver(func(value interface{}) (interface{}, error) {
		if isEmptyOrNil(value) {
			return float32(0), nil
		}
		return cast.ToFloat32E(value)
	}).ResolveTo
	float64Resolver = ext.WrapValueObjectResolver(func(value interface{}) (interface{}, error) {
		if isEmptyOrNil(value) {
			return float64(0), nil
		}
		return cast.ToFloat64E(value)
	}).ResolveTo
	booleanResolver = ext.WrapValueObjectResolver(func(value interface{}) (interface{}, error) {
		if isEmptyOrNil(value) {
			return false, nil
		}
		return cast.ToBoolE(value)
	}).ResolveTo
	mapResolver = ext.ValueObjectResolver(func(value flux.ValueObject, _ string, genericTypes []string) (interface{}, error) {
		return ToStringMapE(value)
	})
	listResolver = ext.ValueObjectResolver(func(value flux.ValueObject, _ string, genericTypes []string) (interface{}, error) {
		return ToGenericListE(genericTypes, value)
	})
	complexObjectResolver = ext.ValueObjectResolver(func(mtValue flux.ValueObject, class string, generic []string) (interface{}, error) {
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
	ext.RegisterObjectValueResolver("string", stringResolver)
	ext.RegisterObjectValueResolver(JavaLangStringClassName, stringResolver)

	ext.RegisterObjectValueResolver("int", integerResolver)
	ext.RegisterObjectValueResolver(JavaLangIntegerClassName, integerResolver)

	ext.RegisterObjectValueResolver("int64", longResolver)
	ext.RegisterObjectValueResolver("long", longResolver)
	ext.RegisterObjectValueResolver(JavaLangLongClassName, longResolver)

	ext.RegisterObjectValueResolver("float", float32Resolver)
	ext.RegisterObjectValueResolver("float32", float32Resolver)
	ext.RegisterObjectValueResolver(JavaLangFloatClassName, float32Resolver)

	ext.RegisterObjectValueResolver("float64", float64Resolver)
	ext.RegisterObjectValueResolver("double", float64Resolver)
	ext.RegisterObjectValueResolver(JavaLangDoubleClassName, float64Resolver)

	ext.RegisterObjectValueResolver("bool", booleanResolver)
	ext.RegisterObjectValueResolver("boolean", booleanResolver)
	ext.RegisterObjectValueResolver(JavaLangBooleanClassName, booleanResolver)

	ext.RegisterObjectValueResolver("map", mapResolver)
	ext.RegisterObjectValueResolver(JavaUtilMapClassName, mapResolver)

	ext.RegisterObjectValueResolver("slice", listResolver)
	ext.RegisterObjectValueResolver("list", listResolver)
	ext.RegisterObjectValueResolver(JavaUtilListClassName, listResolver)

	ext.RegisterObjectValueResolver(JavaIOSerializable, objectResolver)
	ext.RegisterObjectValueResolver(JavaLangObjectClassName, objectResolver)

	ext.RegisterObjectValueResolver(ext.DefaultValueObjectResolverName, complexObjectResolver)
}

// CastDecodeMTValueToString 最大努力地将值转换成String类型。
// 如果类型无法安全地转换成String或者解析异常，返回错误。
func CastDecodeMTValueToString(mtValue flux.ValueObject) (string, error) {
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
func ToStringMapE(mtValue flux.ValueObject) (map[string]interface{}, error) {
	if isEmptyOrNil(mtValue.Value) || !mtValue.Valid {
		return make(map[string]interface{}, 0), nil
	}
	switch mtValue.Encoding {
	case flux.EncodingTypeMapStringList:
		orimap, ok := mtValue.Value.(map[string][]string)
		flux.AssertM(ok, func() string {
			return fmt.Sprintf("mt-value(define:%s) is not map[string][]string, mt-value:%+v", mtValue.Encoding, mtValue.Value)
		})
		var hashmap = make(map[string]interface{}, len(orimap))
		for k, v := range orimap {
			hashmap[k] = v
		}
		return hashmap, nil
	case flux.EncodingTypeGoMapString:
		return cast.ToStringMap(mtValue.Value), nil
	case flux.EncodingTypeGoString:
		oristr, ok := mtValue.Value.(string)
		flux.AssertM(ok, func() string {
			return fmt.Sprintf("mt-value(define:%s) is not go:string, mt-value:%+v", mtValue.Encoding, mtValue.Value)
		})
		var hashmap = map[string]interface{}{}
		if err := ext.JSONUnmarshal([]byte(oristr), &hashmap); nil != err {
			return nil, fmt.Errorf("cannot decode text to hashmap, text: %s, error:%w", mtValue.Value, err)
		} else {
			return hashmap, nil
		}
	case flux.EncodingTypeGoObject:
		if sm, err := cast.ToStringMapE(mtValue.Value); nil != err {
			return nil, fmt.Errorf("cannot cast object to hashmap, object: %+v, object.type:%T", mtValue.Value, mtValue.Value)
		} else {
			return sm, nil
		}
	default:
		var data []byte
		if mtValue.Encoding.Contains("application/json") {
			if bs, err := toByteArray(mtValue.Value); nil != err {
				return nil, err
			} else {
				data = bs
			}
		} else if mtValue.Encoding.Contains("application/x-www-form-urlencoded") {
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
					mtValue.Value, mtValue.Value, mtValue.Encoding)
			}
		}
		var hashmap = map[string]interface{}{}
		err := ext.JSONUnmarshal(data, &hashmap)
		return hashmap, err
	}
}

// ToGenericListE 最大努力地将值转换成[]any类型。
// 如果类型无法安全地转换成[]any或者解析异常，返回错误。
func ToGenericListE(generics []string, mtValue flux.ValueObject) (interface{}, error) {
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
	resolver := ext.ValueObjectResolverByType(generic)
	kind := vType.Kind()
	if kind == reflect.Slice {
		vValue := reflect.ValueOf(mtValue.Value)
		out := make([]interface{}, vValue.Len())
		for i := 0; i < vValue.Len(); i++ {
			if v, err := resolver(ext.NewObjectValueObject(vValue.Index(i).Interface()), generic, []string{}); nil != err {
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
