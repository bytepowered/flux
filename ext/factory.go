package ext

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/toolkit"
	"strings"
)

var (
	typedFactories = make(map[string]flux.Factory, 16)
)

func RegisterFactory(typeName string, factory flux.Factory) {
	typeName = toolkit.MustNotEmpty(typeName, "typeName is empty")
	typeName = strings.ToLower(typeName)
	typedFactories[typeName] = toolkit.MustNotNil(factory, "Factory is nil").(flux.Factory)
}

func FactoryByType(typeName string) (flux.Factory, bool) {
	typeName = toolkit.MustNotEmpty(typeName, "typeName is empty")
	typeName = strings.ToLower(typeName)
	f, o := typedFactories[typeName]
	return f, o
}
