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

// SetMTValueResolver 添加实际值类型解析函数
func SetMTValueResolver(typeName string, resolver flux.MTValueResolver) {
	typeName = pkg.RequireNotEmpty(typeName, "typeName is empty")
	typeName = strings.ToLower(typeName)
	mediaTypeValueResolvers[typeName] = resolver
}

// GetMTValueResolver 获取值类型解析函数；如果指定类型不存在，返回默认解析函数；
func GetMTValueResolver(typeName string) flux.MTValueResolver {
	typeName = pkg.RequireNotEmpty(typeName, "typeName is empty")
	typeName = strings.ToLower(typeName)
	if r, ok := mediaTypeValueResolvers[typeName]; ok {
		return r
	} else {
		return mediaTypeValueResolvers[DefaultMTValueResolverName]
	}
}
