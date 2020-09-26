package support

import (
	"github.com/bytepowered/flux"
	"github.com/spf13/cast"
	"strings"
)

// ScopeLookupWebContextValue 搜索Lookup指定域的值。支持：
// 1. query:<name>
// 2. form:<name>
// 3. path:<name>
// 4. header:<name>
// 5. attr:<name>
func ScopeLookupWebContextValue(lookup string, ctx flux.WebContext) string {
	if "" == lookup || nil == ctx {
		return ""
	}
	scope, key, ok := ScopeLookupParsePair(lookup)
	if !ok {
		return ""
	}
	switch strings.ToUpper(scope) {
	case flux.ScopeQuery:
		return ctx.QueryValue(key)
	case flux.ScopeForm:
		return ctx.FormValue(key)
	case flux.ScopePath:
		return ctx.PathValue(key)
	case flux.ScopeHeader:
		return ctx.HeaderValue(key)
	case flux.ScopeAttr:
		return cast.ToString(ctx.GetValue(key))
	default:
		return ""
	}
}

// ScopeLookupContextValue 搜索Lookup指定域的值。支持：
// 1. query:<name>
// 2. form:<name>
// 3. path:<name>
// 4. header:<name>
// 5. attr:<name>
func ScopeLookupContextValue(lookup string, ctx flux.Context) interface{} {
	if "" == lookup || nil == ctx {
		return nil
	}
	scope, key, ok := ScopeLookupParsePair(lookup)
	if !ok {
		return ""
	}
	req := ctx.Request()
	switch scope {
	case flux.ScopeQuery:
		return req.QueryValue(key)
	case flux.ScopeForm:
		return req.FormValue(key)
	case flux.ScopePath:
		return req.PathValue(key)
	case flux.ScopeHeader:
		return req.HeaderValue(key)
	case flux.ScopeAttr:
		v, _ := ctx.GetAttachment(key)
		return v
	default:
		return nil
	}
}

// ScopeLookupParsePair 解析Lookup键值对
func ScopeLookupParsePair(lookup string) (scope, key string, ok bool) {
	if "" == lookup {
		return
	}
	kv := strings.Split(lookup, ":")
	if len(kv) < 2 || ("" == kv[0] || "" == kv[1]) {
		return
	}
	return strings.ToUpper(kv[0]), kv[1], true
}
