package support

import (
	"errors"
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
	scope, key, ok := ParseLookupExpr(lookupExpr)
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
		if v, ok := SearchValueProviders(key, webc.PathValues, webc.QueryValues, webc.FormValues, WrapHeaderProviderFunc(webc)); ok {
			return v
		}
		return cast.ToString(webc.GetValue(key))
	default:
		return ""
	}
}

// LookupContextByExpr 搜索LookupExpr表达式指定域的值。
func LookupContextByExpr(lookupExpr string, ctx flux.Context) (interface{}, error) {
	if "" == lookupExpr || nil == ctx {
		return nil, errors.New("empty lookup expr or context")
	}
	scope, key, ok := ParseLookupExpr(lookupExpr)
	if !ok {
		return "", errors.New("illegal lookup expr: " + lookupExpr)
	}
	mtv, err := DefaultArgumentValueLookupFunc(scope, key, ctx)
	if nil != err {
		return "", err
	}
	return mtv.Value, nil
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

// ParseLookupExpr 解析Lookup键值对
func ParseLookupExpr(lookupExpr string) (scope, key string, ok bool) {
	if "" == lookupExpr {
		return
	}
	kv := strings.Split(lookupExpr, ":")
	if len(kv) < 2 || ("" == kv[0] || "" == kv[1]) {
		return
	}
	return strings.ToUpper(kv[0]), kv[1], true
}

func WrapHeaderProviderFunc(req flux.RequestReader) func() url.Values {
	return func() url.Values {
		h, _ := req.HeaderValues()
		return url.Values(h)
	}
}
