package ext

import (
	"errors"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/pkg"
)

// Default name
const (
	TypeNameSerializerDefault = "default"
	TypeNameSerializerJson    = "json"
)

var (
	_typeNamedSerializers = make(map[string]flux.Serializer, 2)
)

////

func StoreSerializer(typeName string, serializer flux.Serializer) {
	typeName = pkg.RequireNotEmpty(typeName, "typeName is empty")
	_typeNamedSerializers[typeName] = pkg.RequireNotNil(serializer, "Serializer is nil").(flux.Serializer)
}

func LoadSerializer(typeName string) flux.Serializer {
	typeName = pkg.RequireNotEmpty(typeName, "typeName is empty")
	return _typeNamedSerializers[typeName]
}

func JSONMarshal(data interface{}) ([]byte, error) {
	json := _typeNamedSerializers[TypeNameSerializerJson]
	if nil == json {
		return nil, errors.New("JSON serializer not found")
	}
	return json.Marshal(data)
}

func JSONUnmarshal(data []byte, out interface{}) error {
	json := _typeNamedSerializers[TypeNameSerializerJson]
	if nil == json {
		return errors.New("JSON serializer not found")
	}
	return json.Unmarshal(data, out)
}
