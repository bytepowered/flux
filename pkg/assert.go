package pkg

import "fmt"

func ServerAssertT(assert func() bool, message string) {
	if !assert() {
		panic("ServerAssert: " + message)
	}
}

func ServerAssert(flag bool, message string) {
	if !flag {
		panic("ServerAssert: " + message)
	}
}

func ServerAssertF(flag bool, message string, args ...interface{}) {
	if !flag {
		panic("ServerAssert: " + fmt.Sprintf(message, args...))
	}
}
