package pkg

import "reflect"

// 检查非Nil值。当为Nil时，Panic报错
func RequireNotNil(v interface{}, msg string) interface{} {
	if IsNil(v) {
		panic(msg)
	}
	return v
}

// RequireNotEmpty 检查并返回非空字符串。Panic报错
func RequireNotEmpty(str string, msg string) string {
	if "" == str {
		panic(msg)
	}
	return str
}

func IsNil(v interface{}) bool {
	return v == nil || reflect.ValueOf(v).IsNil()
}

func IsNotNil(v interface{}) bool {
	return !IsNil(v)
}
