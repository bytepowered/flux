package support

import (
	"github.com/bytepowered/flux"
	"github.com/spf13/cast"
	"net/url"
	"strings"
)

// LookupWebContextByExpr 搜索LookupExpr表达式指定域的值。
func LookupWebContextByExpr(lookupExpr string, webc flux.WebContext) string {
	if "" == lookupExpr || nil == webc {
		return ""
	}
	scope, key, ok := LookupParseExpr(lookupExpr)
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
		v, _ := SearchValueProviders(key, webc.QueryValues, webc.FormValues)
		return v
	case flux.ScopeAuto:
		if v, ok := SearchValueProviders(key, webc.PathValues, webc.QueryValues, webc.FormValues, makeHeaderProvider(webc)); ok {
			return v
		}
		return cast.ToString(webc.GetValue(key))
	default:
		return ""
	}
}

// LookupByExprContext 搜索LookupExpr表达式指定域的值。
func LookupByExprContext(lookupExpr string, ctx flux.Context) interface{} {
	if "" == lookupExpr || nil == ctx {
		return nil
	}
	scope, key, ok := LookupParseExpr(lookupExpr)
	if !ok {
		return ""
	}
	return LookupValueContext(scope, key, ctx)
}

// LookupValueContext 搜索Lookup指定域的值。支持：
func LookupValueContext(scope, key string, ctx flux.Context) interface{} {
	req := ctx.Request()
	switch strings.ToUpper(scope) {
	case flux.ScopePath:
		return req.PathValue(key)
	case flux.ScopeQuery:
		return req.QueryValue(key)
	case flux.ScopeForm:
		return req.FormValue(key)
	case flux.ScopeHeader:
		return req.HeaderValue(key)
	case flux.ScopeAttr:
		v, _ := ctx.GetAttribute(key)
		return v
	case flux.ScopeAttrs:
		return ctx.Attributes()
	case flux.ScopeRequest:
		switch strings.ToLower(key) {
		case "method":
			return ctx.Method()
		case "uri":
			return ctx.RequestURI()
		}
		return ctx.Method()
	case flux.ScopeParam:
		v, _ := SearchValueProviders(key, req.QueryValues, req.FormValues)
		return v
	case flux.ScopeAuto:
		if v, ok := SearchValueProviders(key, req.PathValues, req.QueryValues, req.FormValues, makeHeaderProvider(ctx.Request())); ok {
			return v
		}
		av, _ := ctx.GetAttribute(key)
		return av
	default:
		return nil
	}
}

func SearchValueProviders(key string, providers ...func() url.Values) (string, bool) {
	for _, fun := range providers {
		values := fun()
		if v, ok := values[key]; ok {
			return v[0], true
		}
	}
	return "", false
}

// LookupParseExpr 解析Lookup键值对
func LookupParseExpr(lookupExpr string) (scope, key string, ok bool) {
	if "" == lookupExpr {
		return
	}
	kv := strings.Split(lookupExpr, ":")
	if len(kv) < 2 || ("" == kv[0] || "" == kv[1]) {
		return
	}
	return strings.ToUpper(kv[0]), kv[1], true
}

func makeHeaderProvider(req flux.RequestReader) func() url.Values {
	return func() url.Values {
		h, _ := req.HeaderValues()
		return url.Values(h)
	}
}
