package internal

import "github.com/bytepowered/flux"

// 默认实现：查找Argument的值函数
func DefaultArgumentValueLookupFunc(arg flux.Argument, ctx flux.Context) interface{} {
	request := ctx.RequestReader()
	switch arg.HttpScope {
	case flux.ScopeQuery:
		return request.QueryValue(arg.HttpKey)
	case flux.ScopePath:
		return request.PathValue(arg.HttpKey)
	case flux.ScopeParam:
		if v := request.QueryValue(arg.HttpKey); "" == v {
			return request.FormValue(arg.HttpKey)
		} else {
			return v
		}
	case flux.ScopeHeader:
		return request.HeaderValue(arg.HttpKey)
	case flux.ScopeForm:
		return request.FormValue(arg.HttpKey)
	case flux.ScopeAttrs:
		return ctx.AttrValues()
	case flux.ScopeAttr:
		value, _ := ctx.AttrValue(arg.HttpKey)
		return value
	case flux.ScopeAuto:
		fallthrough
	default:
		if v := request.PathValue(arg.HttpKey); "" != v {
			return v
		} else if v := request.QueryValue(arg.HttpKey); "" != v {
			return v
		} else if v := request.FormValue(arg.HttpKey); "" != v {
			return v
		} else if v := request.HeaderValue(arg.HttpKey); "" != v {
			return v
		} else if v, _ := ctx.AttrValue(arg.HttpKey); "" != v {
			return v
		} else {
			return nil
		}
	}
}
