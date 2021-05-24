package toolkit

import (
	"fmt"
	"reflect"
)

const (
	AssertMessagePrefix = "SERVER:CRITICAL:ASSERT:"
)

func AssertT(tester func() bool, message string) {
	if !tester() {
		panic(AssertMessagePrefix + message)
	}
}

func AssertL(true bool, lazyMessage func() string) {
	if !true {
		panic(AssertMessagePrefix + lazyMessage())
	}
}

func Assert(true bool, message string) {
	if !true {
		panic(AssertMessagePrefix + message)
	}
}

func AssertNil(v interface{}, message string, args ...interface{}) {
	if IsNotNil(v) {
		panic(AssertMessagePrefix + fmt.Sprintf(message, args...))
	}
}

func AssertNotNil(v interface{}, message string, args ...interface{}) {
	if IsNil(v) {
		panic(AssertMessagePrefix + fmt.Sprintf(message, args...))
	}
}

func AssertEmpty(v string, message string, args ...interface{}) {
	if v != "" {
		panic(AssertMessagePrefix + fmt.Sprintf(message, args...))
	}
}

func AssertNotEmpty(v string, message string, args ...interface{}) {
	if v == "" {
		panic(AssertMessagePrefix + fmt.Sprintf(message, args...))
	}
}

// 检查非Nil值。当为Nil时，Panic报错
func MustNotNil(v interface{}, msg string) interface{} {
	AssertNotNil(v, msg)
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
	case reflect.Chan, reflect.Func, reflect.Map,
		reflect.Interface, reflect.Slice,
		reflect.Ptr, reflect.UnsafePointer:
		return reflect.ValueOf(i).IsNil()
	}
	return false
}

func IsNotNil(v interface{}) bool {
	return !IsNil(v)
}
