package internal

import (
	"github.com/bytepowered/flux"
	jsoniter "github.com/json-iterator/go"
)

// 默认JSON序列化实现
type JsonSerializer struct {
	json jsoniter.API
}

func (s *JsonSerializer) Marshal(v interface{}) ([]byte, error) {
	return s.json.Marshal(v)
}

func (s *JsonSerializer) Unmarshal(d []byte, v interface{}) error {
	return s.json.Unmarshal(d, v)
}

func NewJsonSerializer() flux.Serializer {
	return &JsonSerializer{json: jsoniter.ConfigCompatibleWithStandardLibrary}
}
