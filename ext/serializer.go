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

func StoreSerializer(typeName string, serializer flux.Serializer) {
	typeName = pkg.RequireNotEmpty(typeName, "typeName is empty")
	_typeNamedSerializers[typeName] = pkg.RequireNotNil(serializer, "Serializer is nil").(flux.Serializer)
}

func LoadSerializer(typeName string) flux.Serializer {
	typeName = pkg.RequireNotEmpty(typeName, "typeName is empty")
	return _typeNamedSerializers[typeName]
}
