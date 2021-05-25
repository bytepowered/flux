package flux

import (
	"fmt"
	"reflect"
)

const (
	assertPrefix = "SERVER:CRITICAL:ASSERT:"
)

func AssertT(tester func() bool, message string) {
	if !tester() {
		panic(assertPrefix + message)
	}
}

func AssertM(true bool, lazyMessage func() string) {
	if !true {
		panic(assertPrefix + lazyMessage())
	}
}

func Assert(true bool, message string) {
	AssertTrue(true, message)
}

func AssertTrue(true bool, message string) {
	if !true {
		panic(assertPrefix + message)
	}
}

func AssertNil(v interface{}, message string, args ...interface{}) {
	if NotNil(v) {
		panic(assertPrefix + fmt.Sprintf(message, args...))
	}
}

func AssertNotNil(v interface{}, message string, args ...interface{}) {
	if IsNil(v) {
		panic(assertPrefix + fmt.Sprintf(message, args...))
	}
}

func AssertIsEmpty(v string, message string, args ...interface{}) {
	if v != "" {
		panic(assertPrefix + fmt.Sprintf(message, args...))
	}
}

func AssertNotEmpty(v string, message string, args ...interface{}) {
	if v == "" {
		panic(assertPrefix + fmt.Sprintf(message, args...))
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
	value := reflect.ValueOf(i)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map,
		reflect.Interface, reflect.Slice,
		reflect.Ptr, reflect.UnsafePointer:
		return value.IsNil()
	}
	return false
}

func NotNil(v interface{}) bool {
	return !IsNil(v)
}
