package fluxkit

import (
	"net/url"
	"strings"
)

// ParseDefineExpr 解析键值定义的表达式（KEY=VALUE），返回(Key,Value)元组；
func ParseDefineExpr(expr string) (key, value string, ok bool) {
	return ParseExprBySep(expr, "=")
}

// ParseScopeExpr 解析指定查找域的表达式（SCOPE:KEY），返回(Scope,key)元组；
func ParseScopeExpr(expr string) (scope, key string, ok bool) {
	return ParseExprBySep(expr, ":")
}

func ParseExprBySep(expr, sep string) (scope, key string, ok bool) {
	if "" == expr {
		return
	}
	kv := strings.Split(expr, sep)
	if len(kv) == 1 {
		return expr, "", false
	}
	return kv[0], kv[1], true
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
