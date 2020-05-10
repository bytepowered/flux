package pkg

import (
	"fmt"
	"reflect"
	"strconv"
)

type ValueError struct {
	Message string
	Value   interface{}
}

func (e *ValueError) Error() string {
	return fmt.Sprintf("value resolve: %s, value: %v", e.Message, e.Value)
}

func ToString(v interface{}) string {
	return ToStringWith(v, func(i interface{}) string {
		return fmt.Sprintf("%s", v)
	})
}

func ToStringWith(v interface{}, strFunc func(interface{}) string) string {
	if nil == v {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	if b, ok := v.([]byte); ok {
		return string(b)
	}
	if ser, ok := v.(fmt.Stringer); ok {
		return ser.String()
	} else {
		return strFunc(v)
	}
}

func ToInt(v interface{}) (int, error) {
	i, e := ToInt64(v)
	return int(i), e
}

func ToInt64(v interface{}) (int64, error) {
	if nil == v {
		return 0, nil
	}
	rvalue := reflect.ValueOf(v)
	switch rvalue.Kind() {
	case reflect.Int:
		fallthrough
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		return rvalue.Int(), nil
	case reflect.Uint:
		fallthrough
	case reflect.Uint8:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		return int64(rvalue.Uint()), nil
	case reflect.Float32:
		fallthrough
	case reflect.Float64:
		return int64(rvalue.Float()), nil
	case reflect.String:
		if s := rvalue.String(); "" == s {
			return 0, nil
		} else {
			i, e := strconv.ParseFloat(s, 64)
			return int64(i), e
		}
	default:
		return 0, &ValueError{Message: "cannot be converted to int64", Value: v}

	}
}

func ToFloat32(v interface{}) (float32, error) {
	f, e := ToFloat64(v)
	return float32(f), e
}

func ToFloat64(v interface{}) (float64, error) {
	if nil == v {
		return 0, nil
	}
	rvalue := reflect.ValueOf(v)
	switch rvalue.Kind() {
	case reflect.Int:
		fallthrough
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		return float64(rvalue.Int()), nil
	case reflect.Uint:
		fallthrough
	case reflect.Uint8:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		return float64(rvalue.Uint()), nil
	case reflect.Float32:
		fallthrough
	case reflect.Float64:
		return rvalue.Float(), nil
	case reflect.String:
		if s := rvalue.String(); "" == s {
			return 0, nil
		} else {
			f, e := strconv.ParseFloat(s, 64)
			return f, e
		}
	default:
		return 0, &ValueError{Message: "cannot be converted to float64", Value: v}

	}
}

func ToBool(value interface{}) bool {
	if v, ok := value.(bool); ok {
		return v
	} else {
		b, _ := strconv.ParseBool(fmt.Sprintf("%s", value))
		return b
	}
}
