package toolkit

import (
	"fmt"
	"io"
	"io/ioutil"
)

// ReadReaderBytesE 读取HttpBody的数据，返回字节数组
func ReadReaderBytesE(reader io.ReadCloser, err error) ([]byte, error) {
	if err != nil {
		return nil, err
	}
	if data, err := ioutil.ReadAll(reader); err != nil {
		return nil, err
	} else {
		return data, nil
	}
}

// ReadReaderBytes 读取HttpBody的数据，返回字节数组
func ReadReaderBytes(reader io.ReadCloser, err error) []byte {
	data, err := ReadReaderBytesE(reader, err)
	if err != nil {
		return []byte(fmt.Sprintf("<error> %s", err))
	}
	return data
}
