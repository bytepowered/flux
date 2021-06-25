package flux

import (
	"fmt"
	"os"
	"reflect"
)

const (
	assertPrefix = "SERVER:CRITICAL:ASSERT:"
)

var (
	_AssertEnabled = func() bool {
		// 允许通过环境变量禁用断言
		v, ok := os.LookupEnv("FLUX_ASSERT_DISABLED")
		return !ok || "true" != v
	}()
)

// AssertT 对provider函数返回值进行断言，期望函数返回值为True；如果函数返回值为False，断言将触发panic，抛出错误消息。
func AssertT(provider func() bool, errMessage string) {
	if _AssertEnabled && !provider() {
		panic(fmt.Errorf(assertPrefix+"%s", errMessage))
	}
}

// AssertM 对输入bool值进行断言，期望为True；如果输入值为False，断言将触发panic，抛出的错误消息由函数延迟加载。
func AssertM(shouldTrue bool, errMessageLoader func() string) {
	if _AssertEnabled && !shouldTrue {
		panic(fmt.Errorf(assertPrefix+"%s", errMessageLoader()))
	}
}

// Assert 对输入bool值进行断言，期望为True；如果输入值为False，断言将触发panic，抛出错误消息。
func Assert(shouldTrue bool, errMessage string) {
	AssertTrue(shouldTrue, errMessage)
}

// AssertTrue 对输入bool值进行断言，期望为True；如果输入值为False，断言将触发panic，抛出错误消息。
func AssertTrue(shouldTrue bool, errMessage string) {
	if _AssertEnabled && !shouldTrue {
		panic(fmt.Errorf(assertPrefix+"%s", errMessage))
	}
}

// AssertNil 对输入值进行断言，期望为Nil(包含nil和值nil情况)；如果输入值为非Nil，断言将触发panic，抛出错误消息（消息模板）。
func AssertNil(v interface{}, message string, args ...interface{}) {
	if _AssertEnabled && NotNil(v) {
		panic(fmt.Errorf(assertPrefix+"%s", fmt.Sprintf(message, args...)))
	}
}

// AssertNotNil 对输入值进行断言，期望为非Nil(非nil并且值非nil的情况)；如果输入值为Nil，断言将触发panic，抛出错误消息（消息模板）。
func AssertNotNil(v interface{}, message string, args ...interface{}) {
	if _AssertEnabled && IsNil(v) {
		panic(fmt.Errorf(assertPrefix+"%s", fmt.Sprintf(message, args...)))
	}
}

// AssertIsEmpty 对输入字符串进行断言，期望为空字符；如果输入值为非空字符串，断言将触发panic，抛出错误消息（消息模板）。
func AssertIsEmpty(v string, message string, args ...interface{}) {
	if _AssertEnabled && v != "" {
		panic(fmt.Errorf(assertPrefix+"%s", fmt.Sprintf(message, args...)))
	}
}

// AssertNotEmpty 对输入字符串进行断言，期望为非空字符；如果输入值为空字符串，断言将触发panic，抛出错误消息（消息模板）。
func AssertNotEmpty(v string, message string, args ...interface{}) {
	if _AssertEnabled && v == "" {
		panic(fmt.Errorf(assertPrefix+"%s", fmt.Sprintf(message, args...)))
	}
}

// MustNotNil 对输入值断言，期望为非Nil值；断言成功时返回原值。当值为Nil时，触发Panic断言，抛出错误消息。
func MustNotNil(v interface{}, msg string) interface{} {
	AssertNotNil(v, msg)
	return v
}

// MustNotEmpty 对输入字符串断言，期望为非空字符串；断言成功时返回原值。当值为空字符串时，触发Panic断言，抛出错误消息。
func MustNotEmpty(str string, msg string) string {
	AssertNotEmpty(str, msg)
	return str
}

// IsNil 判断输入值是否为Nil值（包括：nil、类型非Nil但值为Nil），用于检查类型值是否为Nil。
// 只针对引用类型判断有效，任何数值类型、结构体非指针类型等均为非Nil值。
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

// NotNil 判断输入值是否为非Nil值（包括：nil、类型非Nil但值为Nil），用于检查类型值是否为Nil。
// 只针对引用类型判断有效，任何数值类型、结构体非指针类型等均为非Nil值。
func NotNil(v interface{}) bool {
	return !IsNil(v)
}
