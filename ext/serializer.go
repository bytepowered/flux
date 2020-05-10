package ext

import (
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

func SetSerializer(typeName string, serializer flux.Serializer) {
	_typeNamedSerializers[typeName] = pkg.RequireNotNil(serializer, "Serializer is nil").(flux.Serializer)
}

func GetSerializer(typeName string) flux.Serializer {
	return _typeNamedSerializers[typeName]
}
