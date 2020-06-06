package pkg

import "io"

func CloseSilently(c io.Closer) {
	_ = c.Close()
}

func Silently(_ interface{}) {

}
