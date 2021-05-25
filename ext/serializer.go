package ext

import (
	"errors"
	"fmt"
	"github.com/bytepowered/flux"
	"io"
	"io/ioutil"
	"strings"
)

// Default name
const (
	TypeNameSerializerDefault = "default"
	TypeNameSerializerJson    = "json"
)

var (
	typedSerializers = make(map[string]flux.Serializer, 2)
)

////

func RegisterSerializer(typeName string, serializer flux.Serializer) {
	typeName = flux.MustNotEmpty(typeName, "typeName is empty")
	typeName = strings.ToLower(typeName)
	typedSerializers[typeName] = flux.MustNotNil(serializer, "Serializer is nil").(flux.Serializer)
}

func SerializerByType(typeName string) flux.Serializer {
	typeName = flux.MustNotEmpty(typeName, "typeName is empty")
	typeName = strings.ToLower(typeName)
	return typedSerializers[typeName]
}

func JSONMarshal(data interface{}) ([]byte, error) {
	json := typedSerializers[TypeNameSerializerJson]
	if nil == json {
		return nil, errors.New("JSON serializer not found")
	}
	return json.Marshal(data)
}

func JSONUnmarshal(data []byte, out interface{}) error {
	json := typedSerializers[TypeNameSerializerJson]
	if nil == json {
		return errors.New("JSON serializer not found")
	}
	return json.Unmarshal(data, out)
}

func JSONMarshalObject(body interface{}) ([]byte, error) {
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
		if bytes, err := JSONMarshal(body); nil != err {
			return nil, fmt.Errorf("SERVER:SERIALIZE/JSON: %w", err)
		} else {
			return bytes, nil
		}
	}
}
