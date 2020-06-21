package pkg

import "io"

func SilentlyCloseFunc(c io.Closer) func() {
	return func() {
		_ = c.Close()
	}
}

func Silently(_ interface{}) {

}
