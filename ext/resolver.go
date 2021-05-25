package ext

import (
	"github.com/bytepowered/flux"
	"strings"
)

const (
	DefaultMTValueResolverName = "default"
)

var (
	mediaTypeValueResolvers = make(map[string]flux.MTValueResolver, 16)
)

// RegisterMTValueResolver 添加实际值类型解析函数
func RegisterMTValueResolver(typeName string, resolver flux.MTValueResolver) {
	typeName = flux.MustNotEmpty(typeName, "typeName is empty")
	typeName = strings.ToLower(typeName)
	mediaTypeValueResolvers[typeName] = resolver
}

// MTValueResolverByType 获取值类型解析函数；如果指定类型不存在，返回默认解析函数；
func MTValueResolverByType(typeName string) flux.MTValueResolver {
	typeName = flux.MustNotEmpty(typeName, "typeName is empty")
	typeName = strings.ToLower(typeName)
	if r, ok := mediaTypeValueResolvers[typeName]; ok {
		return r
	} else {
		return mediaTypeValueResolvers[DefaultMTValueResolverName]
	}
}
