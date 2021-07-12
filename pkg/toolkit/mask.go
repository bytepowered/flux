package toolkit

import "fmt"

// Mask 对str字符串屏蔽，只保留前后size字符
func Mask(str string, size int) string {
	return string(MaskBytes([]byte(str), size))
}

// MaskBytes 对字符串数据进行屏蔽，只保留前后size字符
func MaskBytes(in []byte, size int) []byte {
	l := len(in)
	if 0 == l || size >= l {
		return in
	}
	if l > size*2 {
		if l > 128 {
			return append(in[:size], append([]byte(fmt.Sprintf("**[%d]**", l)), in[(l-size):]...)...)
		}
		return append(in[:size], append([]byte("**"), in[(l-size):]...)...)
	} else {
		return append(in[:size], '*', '*')
	}
}
