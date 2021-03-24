package fluxpkg

import "fmt"

const (
	assertMessagePrefix = "SERVER:CRITICAL:ASSERT:"
)

func AssertT(tester func() bool, message string) {
	if !tester() {
		panic(assertMessagePrefix + message)
	}
}

func AssertL(true bool, lazyMessage func() string) {
	if !true {
		panic(assertMessagePrefix + lazyMessage())
	}
}

func Assert(true bool, message string) {
	if !true {
		panic(assertMessagePrefix + message)
	}
}

func AssertNil(v interface{}, message string, args ...interface{}) {
	if IsNotNil(v) {
		panic(assertMessagePrefix + fmt.Sprintf(message, args...))
	}
}

func AssertNotNil(v interface{}, message string, args ...interface{}) {
	if IsNil(v) {
		panic(assertMessagePrefix + fmt.Sprintf(message, args...))
	}
}
