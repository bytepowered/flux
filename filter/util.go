package filter

import (
	"github.com/bytepowered/flux"
	"strings"
)

// LookupValue 搜索Lookup指定域的值。支持：
// 1. query:<name>
// 2. form:<name>
// 3. path:<name>
// 4. header:<name>
// 5. attr:<name>
// 默认取Header的值
func LookupValue(lookup string, ctx flux.Context) interface{} {
	req := ctx.RequestReader()
	parts := strings.Split(lookup, ":")
	if len(parts) == 1 {
		return req.Header(parts[0])
	}
	switch parts[0] {
	case "query":
		return req.QueryValue(parts[1])
	case "form":
		return req.FormValue(parts[1])
	case "path":
		return req.PathValue(parts[1])
	case "header":
		return req.HeaderValue(parts[1])
	case "attr":
		v, _ := ctx.AttrValue(parts[1])
		return v
	default:
		return nil
	}
}
