package common

import (
	"fmt"
	"github.com/bytepowered/flux/flux-node/ext"
	"io"
	"io/ioutil"
)

func SerializeObject(body interface{}) ([]byte, error) {
	if bytes, ok := body.([]byte); ok {
		return bytes, nil
	} else if str, ok := body.(string); ok {
		return []byte(str), nil
	} else if r, ok := body.(io.Reader); ok {
		if c, ok := r.(io.Closer); ok {
			defer c.Close()
		}
		if bytes, err := ioutil.ReadAll(r); nil != err {
			return nil, fmt.Errorf("SERVER:SERIALIZE/READER: %w", err)
		} else {
			return bytes, nil
		}
	} else {
		if bytes, err := ext.JSONMarshal(body); nil != err {
			return nil, fmt.Errorf("SERVER:SERIALIZE/JSON: %w", err)
		} else {
			return bytes, nil
		}
	}
}
