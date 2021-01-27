package webserver

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/pkg"
	"github.com/spf13/cast"
	"net/url"
	"strings"
)

// LookupWebContextByExpr 搜索LookupExpr表达式指定域的值。
func LookupWebContextByExpr(lookupExpr string, webc flux.WebContext) string {
	if "" == lookupExpr || nil == webc {
		return ""
	}
	scope, key, ok := pkg.LookupParseExpr(lookupExpr)
	if !ok {
		return ""
	}
	return LookupWebContext(scope, key, webc)
}

func LookupWebContext(scope, key string, webc flux.WebContext) string {
	switch strings.ToUpper(scope) {
	case flux.ScopePath:
		return webc.PathValue(key)
	case flux.ScopeQuery:
		return webc.QueryValue(key)
	case flux.ScopeForm:
		return webc.FormValue(key)
	case flux.ScopeHeader:
		return webc.HeaderValue(key)
	case flux.ScopeAttr:
		return cast.ToString(webc.GetValue(key))
	case flux.ScopeRequest:
		switch strings.ToLower(key) {
		case "method":
			return webc.Method()
		case "uri":
			return webc.RequestURI()
		}
		return webc.Method()
	case flux.ScopeParam:
		v, _ := pkg.LookupByProviders(key, webc.QueryValues, webc.FormValues)
		return v
	case flux.ScopeAuto:
		if v, ok := pkg.LookupByProviders(key, webc.PathValues, webc.QueryValues, webc.FormValues, func() url.Values {
			h, _ := webc.HeaderValues()
			return url.Values(h)
		}); ok {
			return v
		}
		return cast.ToString(webc.GetValue(key))
	default:
		return ""
	}
}
