package internal

func ServerAssertF(assert func() bool, message string) {
	if !assert() {
		panic("ServerAssert: " + message)
	}
}

func ServerAssert(flag bool, message string) {
	if !flag {
		panic("ServerAssert: " + message)
	}
}
