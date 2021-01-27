package ext

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/pkg"
)

var (
	typedFactories = make(map[string]flux.Factory, 16)
)

func SetFactory(typeName string, factory flux.Factory) {
	typeName = pkg.RequireNotEmpty(typeName, "typeName is empty")
	typedFactories[typeName] = pkg.RequireNotNil(factory, "Factory is nil").(flux.Factory)
}

func GetFactory(typeName string) (flux.Factory, bool) {
	typeName = pkg.RequireNotEmpty(typeName, "typeName is empty")
	f, o := typedFactories[typeName]
	return f, o
}
