package pkg

import "reflect"

// 检查非Nil值。当为Nil时，Panic报错
func RequireNotNil(v interface{}, msg string) interface{} {
	if nil == v || reflect.ValueOf(v).IsNil() {
		panic(msg)
	}
	return v
}
