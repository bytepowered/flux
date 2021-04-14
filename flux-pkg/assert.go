package fluxpkg

import "fmt"

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
