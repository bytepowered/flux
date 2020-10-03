package ext

import (
	"github.com/bytepowered/flux"
)

const (
	DefaultTypedValueResolverName = "default"
)

var (
	_typedValueResolvers = make(map[string]flux.TypedValueResolver, 16)
)

// SetTypedValueResolver 添加值类型解析函数
func SetTypedValueResolver(typeName string, resolver flux.TypedValueResolver) {
	_typedValueResolvers[typeName] = resolver
}

// GetTypedValueResolver 获取值类型解析函数
func GetTypedValueResolver(valueTypeName string) flux.TypedValueResolver {
	return _typedValueResolvers[valueTypeName]
}

// GetDefaultTypedValueResolver 获取默认的值类型解析函数
func GetDefaultTypedValueResolver() flux.TypedValueResolver {
	return _typedValueResolvers[DefaultTypedValueResolverName]
}
