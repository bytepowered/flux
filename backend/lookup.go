package backend

import (
	"errors"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/pkg"
	"net/url"
	"strings"
)

// 默认实现：查找Argument的值函数
func DefaultArgumentLookupFunc(scope, key string, ctx flux.Context) (value flux.MTValue, err error) {
	if "" == scope || "" == key {
		return flux.WrapObjectMTValue(nil), errors.New("lookup empty scope or key, scope: " + scope + ", key: " + key)
	}
	if nil == ctx {
		return flux.WrapObjectMTValue(nil), errors.New("lookup nil context")
	}
	req := ctx.Request()
	switch strings.ToUpper(scope) {
	case flux.ScopePath:
		return flux.WrapStringMTValue(req.PathValue(key)), nil
	case flux.ScopePathMap:
		return flux.WrapStrValuesMapMTValue(req.PathValues()), nil
	case flux.ScopeQuery:
		return flux.WrapStringMTValue(req.QueryValue(key)), nil
	case flux.ScopeQueryMulti:
		return flux.WrapStrListMTValue(req.QueryValues()[key]), nil
	case flux.ScopeQueryMap:
		return flux.WrapStrValuesMapMTValue(req.QueryValues()), nil
	case flux.ScopeForm:
		return flux.WrapStringMTValue(req.FormValue(key)), nil
	case flux.ScopeFormMap:
		return flux.WrapStrValuesMapMTValue(req.FormValues()), nil
	case flux.ScopeFormMulti:
		return flux.WrapStrListMTValue(req.FormValues()[key]), nil
	case flux.ScopeHeader:
		return flux.WrapStringMTValue(req.HeaderValue(key)), nil
	case flux.ScopeHeaderMap:
		header, _ := req.HeaderValues()
		return flux.WrapStrValuesMapMTValue(header), nil
	case flux.ScopeAttr:
		v, _ := ctx.GetAttribute(key)
		return flux.WrapObjectMTValue(v), nil
	case flux.ScopeAttrs:
		return flux.WrapStrMapMTValue(ctx.Attributes()), nil
	case flux.ScopeBody:
		reader, err := req.RequestBodyReader()
		return flux.MTValue{Value: reader, MediaType: req.HeaderValue(flux.HeaderContentType)}, err
	case flux.ScopeParam:
		v, _ := pkg.LookupByProviders(key, req.QueryValues, req.FormValues)
		return flux.WrapStringMTValue(v), nil
	case flux.ScopeValue:
		v, _ := ctx.GetValue(key)
		return flux.WrapObjectMTValue(v), nil
	case flux.ScopeRequest:
		switch strings.ToLower(key) {
		case "method":
			return flux.WrapStringMTValue(ctx.Method()), nil
		case "uri":
			return flux.WrapStringMTValue(ctx.RequestURI()), nil
		default:
			return flux.WrapStringMTValue(""), nil
		}
	case flux.ScopeAuto:
		fallthrough
	default:
		if v, ok := pkg.LookupByProviders(key, req.PathValues, req.QueryValues, req.FormValues, func() url.Values {
			h, _ := req.HeaderValues()
			return url.Values(h)
		}); ok {
			return flux.WrapStringMTValue(v), nil
		}
		if v, ok := ctx.GetAttribute(key); ok {
			return flux.WrapObjectMTValue(v), nil
		}
		if v, ok := ctx.GetValue(key); ok {
			return flux.WrapObjectMTValue(v), nil
		}
		return flux.WrapObjectMTValue(nil), nil
	}
}
