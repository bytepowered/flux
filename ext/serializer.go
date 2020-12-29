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
	typedSerializers = make(map[string]flux.Serializer, 2)
)

////

func StoreSerializer(typeName string, serializer flux.Serializer) {
	typeName = pkg.RequireNotEmpty(typeName, "typeName is empty")
	typedSerializers[typeName] = pkg.RequireNotNil(serializer, "Serializer is nil").(flux.Serializer)
}

func LoadSerializer(typeName string) flux.Serializer {
	typeName = pkg.RequireNotEmpty(typeName, "typeName is empty")
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
