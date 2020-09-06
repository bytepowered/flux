package flux

import jsoniter "github.com/json-iterator/go"

// 序列化接口
type Serializer interface {
	// Marshal 将对象序列化为字节数组
	Marshal(any interface{}) (bytes []byte, err error)

	// Unmarshal 将字节数组反序列化为对象；
	// 注意：对象为指针类型；
	Unmarshal(bytes []byte, obj interface{}) error
}

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

func NewJsonSerializer() Serializer {
	return &JsonSerializer{json: jsoniter.ConfigCompatibleWithStandardLibrary}
}
