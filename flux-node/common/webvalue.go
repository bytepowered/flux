package common

import (
	"github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-pkg"
	"github.com/spf13/cast"
	"strings"
)

// LookupWebValueByExpr 搜索LookupExpr表达式指定域的值。
func LookupWebValueByExpr(webc flux.ServerWebContext, expr string) string {
	if "" == expr || nil == webc {
		return ""
	}
	scope, key, ok := fluxpkg.ParseScopeExpr(expr)
	if !ok {
		return ""
	}
	return LookupWebValue(webc, scope, key)
}

// LookupWebValue 根据Scope,Key查找Http请求参数，仅支持Http参数值类型
func LookupWebValue(webc flux.ServerWebContext, scope, key string) string {
	switch strings.ToUpper(scope) {
	case flux.ScopePath:
		return webc.PathVar(key)
	case flux.ScopeQuery:
		return webc.QueryVar(key)
	case flux.ScopeForm:
		return webc.FormVar(key)
	case flux.ScopeHeader:
		return webc.HeaderVar(key)
	case flux.ScopeRequest:
		switch strings.ToUpper(key) {
		case "METHOD":
			return webc.Method()
		case "URI":
			return webc.URI()
		case "HOST":
			return webc.Host()
		case "REMOTEADDR":
			return webc.RemoteAddr()
		default:
			return ""
		}
	case flux.ScopeParam:
		v, _ := fluxpkg.LookupByProviders(key, webc.QueryVars, webc.FormVars)
		return v
	case flux.ScopeAuto:
		// Post args
		if v, ok := fluxpkg.LookupByProviders(key, webc.PathVars, webc.QueryVars, webc.FormVars); ok {
			return v
		}
		// Header: key case insensitive
		if v := webc.HeaderVar(key); v != "" {
			return v
		}
		// Variables
		return cast.ToString(webc.Variable(key))
	default:
		return ""
	}
}
