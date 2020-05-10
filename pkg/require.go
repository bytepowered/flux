package pkg

// 检查非Nil值。当为Nil时，Panic报错
func RequireNotNil(v interface{}, msg string) interface{} {
	if nil == v {
		panic(msg)
	}
	return v
}
