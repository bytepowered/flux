package pkg

import "fmt"

func AssertT(assert func() bool, message string) {
	if !assert() {
		panic("ServerAssert: " + message)
	}
}

func Assert(true bool, message string) {
	if !true {
		panic("ServerAssert: " + message)
	}
}

func AssertNil(v interface{}, message string, args ...interface{}) {
	if nil != v {
		panic("ServerAssert: " + fmt.Sprintf(message, args...))
	}
}

func AssertNotNil(v interface{}, message string, args ...interface{}) {
	if nil == v {
		panic("ServerAssert: " + fmt.Sprintf(message, args...))
	}
}
