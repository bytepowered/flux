package ext

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/pkg"
)

var (
	_typeNamedFactories = make(map[string]flux.Factory, 16)
)

func StoreTypedFactory(typeName string, factory flux.Factory) {
	typeName = pkg.RequireNotEmpty(typeName, "typeName is empty")
	_typeNamedFactories[typeName] = pkg.RequireNotNil(factory, "Factory is nil").(flux.Factory)
}

func LoadTypedFactory(typeName string) (flux.Factory, bool) {
	typeName = pkg.RequireNotEmpty(typeName, "typeName is empty")
	f, o := _typeNamedFactories[typeName]
	return f, o
}
