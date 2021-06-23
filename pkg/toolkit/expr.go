package toolkit

import (
	"strings"
)

const (
	SepKeyValue = "="
	SepScopeKey = ":"
)

// ParseDefineExpr 解析键值定义的表达式（KEY=VALUE），返回(Key,Value)元组；
func ParseDefineExpr(expr string) (key, value string, ok bool) {
	return ParseExprBySep(expr, SepKeyValue)
}

// ParseScopeExpr 解析指定查找域的表达式（SCOPE:KEY），返回(Scope,key)元组；
func ParseScopeExpr(expr string) (scope, key string, ok bool) {
	return ParseExprBySep(expr, SepScopeKey)
}

func ParseExprBySep(expr, sep string) (first, second string, ok bool) {
	if "" == expr || len(expr) < len("a:b") {
		return
	}
	pair := strings.Split(expr, sep)
	if len(pair) != 2 {
		return "", "", false
	}
	f, s := strings.TrimSpace(pair[0]), strings.TrimSpace(pair[1])
	if f != "" && s != "" {
		return f, s, true
	}
	return "", "", false
}
