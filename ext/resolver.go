package ext

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/pkg"
	"strings"
)

const (
	DefaultMTValueResolverName = "default"
)

var (
	mediaTypeValueResolvers = make(map[string]flux.MTValueResolver, 16)
)

// RegisterMTValueResolver 添加实际值类型解析函数
func RegisterMTValueResolver(actualTypeName string, resolver flux.MTValueResolver) {
	actualTypeName = pkg.RequireNotEmpty(actualTypeName, "actualTypeName is empty")
	actualTypeName = strings.ToLower(actualTypeName)
	mediaTypeValueResolvers[actualTypeName] = resolver
}

// LoadMTValueResolver 获取值类型解析函数
func LoadMTValueResolver(actualTypeName string) flux.MTValueResolver {
	actualTypeName = pkg.RequireNotEmpty(actualTypeName, "actualTypeName is empty")
	actualTypeName = strings.ToLower(actualTypeName)
	return mediaTypeValueResolvers[actualTypeName]
}

// LoadMTValueDefaultResolver 获取默认的值类型解析函数
func LoadMTValueDefaultResolver() flux.MTValueResolver {
	return mediaTypeValueResolvers[DefaultMTValueResolverName]
}
