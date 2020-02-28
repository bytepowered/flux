package extension

import "github.com/bytepowered/flux"

var (
	_typeNamedFactories = make(map[string]flux.Factory)
)

func GetFactory(typeName string) (flux.Factory, bool) {
	f, o := _typeNamedFactories[typeName]
	return f, o
}

func SetFactory(typeName string, f flux.Factory) {
	_typeNamedFactories[typeName] = f
}
