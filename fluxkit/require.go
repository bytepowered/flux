package fluxkit

import "reflect"

// 检查非Nil值。当为Nil时，Panic报错
func MustNotNil(v interface{}, msg string) interface{} {
	Assert(v != nil, msg)
	return v
}

// MustNotEmpty 检查并返回非空字符串。Panic报错
func MustNotEmpty(str string, msg string) string {
	Assert("" != str, msg)
	return str
}

func IsNil(i interface{}) bool {
	if nil == i {
		return true
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return reflect.ValueOf(i).IsNil()
	}
	return false
}

func IsNotNil(v interface{}) bool {
	return !IsNil(v)
}
