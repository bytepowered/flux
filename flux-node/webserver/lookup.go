package webserver

import (
	"github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-pkg"
	"github.com/spf13/cast"
	"net/url"
	"strings"
)

// LookupValueByExpr 搜索LookupExpr表达式指定域的值。
func LookupValueByExpr(lookupExpr string, webex flux.WebExchange) string {
	if "" == lookupExpr || nil == webex {
		return ""
	}
	scope, key, ok := fluxpkg.LookupParseExpr(lookupExpr)
	if !ok {
		return ""
	}
	return LookupValue(scope, key, webex)
}

func LookupValue(scope, key string, webex flux.WebExchange) string {
	switch strings.ToUpper(scope) {
	case flux.ScopePath:
		return webex.PathVar(key)
	case flux.ScopeQuery:
		return webex.QueryVar(key)
	case flux.ScopeForm:
		return webex.FormVar(key)
	case flux.ScopeHeader:
		return webex.HeaderVar(key)
	case flux.ScopeRequest:
		switch strings.ToLower(key) {
		case "method":
			return webex.Method()
		case "uri":
			return webex.URI()
		}
		return webex.Method()
	case flux.ScopeParam:
		v, _ := fluxpkg.LookupByProviders(key, webex.QueryVars, webex.FormVars)
		return v
	case flux.ScopeAuto:
		if v, ok := fluxpkg.LookupByProviders(key, webex.PathVars, webex.QueryVars, webex.FormVars, func() url.Values {
			return url.Values(webex.HeaderVars())
		}); ok {
			return v
		}
		return cast.ToString(webex.Variable(key))
	default:
		return ""
	}
}
