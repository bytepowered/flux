package fluxpkg

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

func IsNil(v interface{}) bool {
	return v == nil || reflect.ValueOf(v).IsNil()
}

func IsNotNil(v interface{}) bool {
	return !IsNil(v)
}
