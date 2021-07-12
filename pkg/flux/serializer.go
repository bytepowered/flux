package flux

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/json-iterator/go/extra"
)

// Serializer 序列化接口
type Serializer interface {
	// Marshal 将对象序列化为字节数组
	Marshal(any interface{}) (bytes []byte, err error)

	// Unmarshal 将字节数组反序列化为对象；
	// 注意：对象为指针类型；
	Unmarshal(bytes []byte, obj interface{}) error
}

// JSONSerializer 默认JSON序列化实现
type JSONSerializer struct {
	json jsoniter.API
}

func (s *JSONSerializer) Marshal(v interface{}) ([]byte, error) {
	return s.json.Marshal(v)
}

func (s *JSONSerializer) Unmarshal(d []byte, v interface{}) error {
	return s.json.Unmarshal(d, v)
}

func NewJsonSerializer() Serializer {
	// 容忍字符串和数字互转
	extra.RegisterFuzzyDecoders()
	return &JSONSerializer{json: jsoniter.ConfigCompatibleWithStandardLibrary}
}
