package ext

import (
	"errors"
	"github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-pkg"
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
	typeName = fluxpkg.MustNotEmpty(typeName, "typeName is empty")
	typedSerializers[typeName] = fluxpkg.MustNotNil(serializer, "Serializer is nil").(flux.Serializer)
}

func SerializerByType(typeName string) flux.Serializer {
	typeName = fluxpkg.MustNotEmpty(typeName, "typeName is empty")
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
