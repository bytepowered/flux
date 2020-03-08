package internal

import (
	"github.com/bytepowered/flux"
	jsoniter "github.com/json-iterator/go"
)

// 默认JSON序列化实现
type jsonSerializer struct {
	json jsoniter.API
}

func (s *jsonSerializer) Marshal(v interface{}) ([]byte, error) {
	return s.json.Marshal(v)
}

func (s *jsonSerializer) Unmarshal(d []byte, v interface{}) error {
	return s.json.Unmarshal(d, v)
}

func NewJsonSerializer() flux.Serializer {
	return &jsonSerializer{json: jsoniter.ConfigCompatibleWithStandardLibrary}
}
