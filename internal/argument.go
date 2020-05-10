package internal

import "github.com/bytepowered/flux"

// 默认实现：查找Argument的值函数
func DefaultArgumentValueLookupFunc(arg flux.Argument, ctx flux.Context) (interface{}, error) {
	request := ctx.RequestReader()
	switch arg.HttpScope {
	case flux.ScopeQuery:
		return request.QueryValue(arg.HttpKey), nil
	case flux.ScopePath:
		return request.PathValue(arg.HttpKey), nil
	case flux.ScopeParam:
		if v := request.QueryValue(arg.HttpKey); "" == v {
			return request.FormValue(arg.HttpKey), nil
		} else {
			return v, nil
		}
	case flux.ScopeHeader:
		return request.HeaderValue(arg.HttpKey), nil
	case flux.ScopeForm:
		return request.FormValue(arg.HttpKey), nil
	case flux.ScopeAttrs:
		return ctx.AttrValues(), nil
	case flux.ScopeAttr:
		value, _ := ctx.AttrValue(arg.HttpKey)
		return value, nil
	case flux.ScopeAuto:
		fallthrough
	default:
		if v := request.PathValue(arg.HttpKey); "" != v {
			return v, nil
		} else if v := request.QueryValue(arg.HttpKey); "" != v {
			return v, nil
		} else if v := request.FormValue(arg.HttpKey); "" != v {
			return v, nil
		} else if v := request.HeaderValue(arg.HttpKey); "" != v {
			return v, nil
		} else if v, _ := ctx.AttrValue(arg.HttpKey); "" != v {
			return v, nil
		} else {
			return nil, nil
		}
	}
}
