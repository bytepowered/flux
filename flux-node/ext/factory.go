package ext

import (
	flux2 "github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-pkg"
)

var (
	typedFactories = make(map[string]flux2.Factory, 16)
)

func RegisterFactory(typeName string, factory flux2.Factory) {
	typeName = fluxpkg.MustNotEmpty(typeName, "typeName is empty")
	typedFactories[typeName] = fluxpkg.MustNotNil(factory, "Factory is nil").(flux2.Factory)
}

func FactoryByType(typeName string) (flux2.Factory, bool) {
	typeName = fluxpkg.MustNotEmpty(typeName, "typeName is empty")
	f, o := typedFactories[typeName]
	return f, o
}
