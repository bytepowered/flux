package ext

import (
	"github.com/bytepowered/flux"
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

func SetSerializer(typeName string, serializer flux.Serializer) {
	if nil == serializer {
		GetLogger().Panic("Serialize is nil")
	}
	_typeNamedSerializers[typeName] = serializer
}

func GetSerializer(typeName string) flux.Serializer {
	return _typeNamedSerializers[typeName]
}
