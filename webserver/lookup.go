package webserver

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/pkg"
	"github.com/spf13/cast"
	"net/url"
	"strings"
)

// LookupValueByExpr 搜索LookupExpr表达式指定域的值。
func LookupValueByExpr(lookupExpr string, webc flux.WebContext) string {
	if "" == lookupExpr || nil == webc {
		return ""
	}
	scope, key, ok := pkg.LookupParseExpr(lookupExpr)
	if !ok {
		return ""
	}
	return LookupValue(scope, key, webc)
}

func LookupValue(scope, key string, webc flux.WebContext) string {
	switch strings.ToUpper(scope) {
	case flux.ScopePath:
		return webc.PathVar(key)
	case flux.ScopeQuery:
		return webc.QueryVar(key)
	case flux.ScopeForm:
		return webc.FormVar(key)
	case flux.ScopeHeader:
		return webc.HeaderVar(key)
	case flux.ScopeValue:
		return cast.ToString(webc.GetValue(key))
	case flux.ScopeRequest:
		switch strings.ToLower(key) {
		case "method":
			return webc.Method()
		case "uri":
			return webc.URI()
		}
		return webc.Method()
	case flux.ScopeParam:
		v, _ := pkg.LookupByProviders(key, webc.QueryVars, webc.FormVars)
		return v
	case flux.ScopeAuto:
		if v, ok := pkg.LookupByProviders(key, webc.PathVars, webc.QueryVars, webc.FormVars, func() url.Values {
			return url.Values(webc.HeaderVars())
		}); ok {
			return v
		}
		return cast.ToString(webc.GetValue(key))
	default:
		return ""
	}
}
