package ext

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/pkg"
)

const (
	DefaultTypedValueResolverName = "default"
)

var (
	_typedValueResolvers = make(map[string]flux.TypedValueResolver, 16)
)

// StoreTypedValueResolver 添加值类型解析函数
func StoreTypedValueResolver(typeName string, resolver flux.TypedValueResolver) {
	typeName = pkg.RequireNotEmpty(typeName, "typeName is empty")
	_typedValueResolvers[typeName] = resolver
}

// LoadTypedValueResolver 获取值类型解析函数
func LoadTypedValueResolver(typeName string) flux.TypedValueResolver {
	typeName = pkg.RequireNotEmpty(typeName, "typeName is empty")
	return _typedValueResolvers[typeName]
}

// LoadDefaultTypedValueResolver 获取默认的值类型解析函数
func LoadDefaultTypedValueResolver() flux.TypedValueResolver {
	return _typedValueResolvers[DefaultTypedValueResolverName]
}
