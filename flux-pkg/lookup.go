package fluxpkg

import (
	"net/url"
	"strings"
)

// LookupParseExpr 解析Lookup键值对
func LookupParseExpr(expr string) (scope, key string, ok bool) {
	if "" == expr {
		return
	}
	kv := strings.Split(expr, ":")
	if len(kv) < 2 || ("" == kv[0] || "" == kv[1]) {
		return
	}
	return strings.ToUpper(kv[0]), kv[1], true
}

func LookupByProviders(key string, providers ...func() url.Values) (string, bool) {
	for _, fun := range providers {
		values := fun()
		if v, ok := values[key]; ok {
			return v[0], true
		}
	}
	return "", false
}
