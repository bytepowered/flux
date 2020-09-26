package ext

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/pkg"
)

var (
	_typeNamedFactories = make(map[string]flux.Factory, 16)
)

func GetFactory(typeName string) (flux.Factory, bool) {
	f, o := _typeNamedFactories[typeName]
	return f, o
}

func SetFactory(typeName string, factory flux.Factory) {
	_typeNamedFactories[typeName] = pkg.RequireNotNil(factory, "Factory is nil").(flux.Factory)
}
