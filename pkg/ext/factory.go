package ext

import (
	"github.com/bytepowered/fluxgo/pkg/flux"
	"strings"
)

var (
	typedFactories = make(map[string]flux.Factory, 16)
)

func RegisterFactory(typeName string, factory flux.Factory) {
	typeName = flux.MustNotEmpty(typeName, "typeName is empty")
	typeName = strings.ToLower(typeName)
	typedFactories[typeName] = flux.MustNotNil(factory, "Factory is nil").(flux.Factory)
}

func FactoryByType(typeName string) (flux.Factory, bool) {
	typeName = flux.MustNotEmpty(typeName, "typeName is empty")
	typeName = strings.ToLower(typeName)
	f, o := typedFactories[typeName]
	return f, o
}
