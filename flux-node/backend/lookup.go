package backend

import (
	"errors"
	flux2 "github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-pkg"
	"net/url"
	"strings"
)

// 默认实现：查找Argument的值函数
func DefaultArgumentLookupFunc(scope, key string, ctx flux2.Context) (value flux2.MTValue, err error) {
	if "" == scope || "" == key {
		return flux2.WrapObjectMTValue(nil), errors.New("lookup empty scope or key, scope: " + scope + ", key: " + key)
	}
	if nil == ctx {
		return flux2.WrapObjectMTValue(nil), errors.New("lookup nil context")
	}
	req := ctx.Request()
	switch strings.ToUpper(scope) {
	case flux2.ScopePath:
		return flux2.WrapStringMTValue(req.PathVar(key)), nil
	case flux2.ScopePathMap:
		return flux2.WrapStrValuesMapMTValue(req.PathVars()), nil
	case flux2.ScopeQuery:
		return flux2.WrapStringMTValue(req.QueryVar(key)), nil
	case flux2.ScopeQueryMulti:
		return flux2.WrapStrListMTValue(req.QueryVars()[key]), nil
	case flux2.ScopeQueryMap:
		return flux2.WrapStrValuesMapMTValue(req.QueryVars()), nil
	case flux2.ScopeForm:
		return flux2.WrapStringMTValue(req.FormVar(key)), nil
	case flux2.ScopeFormMap:
		return flux2.WrapStrValuesMapMTValue(req.FormVars()), nil
	case flux2.ScopeFormMulti:
		return flux2.WrapStrListMTValue(req.FormVars()[key]), nil
	case flux2.ScopeHeader:
		return flux2.WrapStringMTValue(req.HeaderVar(key)), nil
	case flux2.ScopeHeaderMap:
		return flux2.WrapStrValuesMapMTValue(req.HeaderVars()), nil
	case flux2.ScopeAttr:
		v, _ := ctx.GetAttribute(key)
		return flux2.WrapObjectMTValue(v), nil
	case flux2.ScopeAttrs:
		return flux2.WrapStrMapMTValue(ctx.Attributes()), nil
	case flux2.ScopeBody:
		reader, err := req.BodyReader()
		return flux2.MTValue{Value: reader, MediaType: req.HeaderVar(flux2.HeaderContentType)}, err
	case flux2.ScopeParam:
		v, _ := fluxpkg.LookupByProviders(key, req.QueryVars, req.FormVars)
		return flux2.WrapStringMTValue(v), nil
	case flux2.ScopeRequest:
		switch strings.ToLower(key) {
		case "method":
			return flux2.WrapStringMTValue(ctx.Method()), nil
		case "uri":
			return flux2.WrapStringMTValue(ctx.URI()), nil
		default:
			return flux2.WrapStringMTValue(""), nil
		}
	case flux2.ScopeAuto:
		fallthrough
	default:
		if v, ok := fluxpkg.LookupByProviders(key, req.PathVars, req.QueryVars, req.FormVars, func() url.Values {
			return url.Values(req.HeaderVars())
		}); ok {
			return flux2.WrapStringMTValue(v), nil
		}
		if v, ok := ctx.GetAttribute(key); ok {
			return flux2.WrapObjectMTValue(v), nil
		}
		return flux2.WrapObjectMTValue(nil), nil
	}
}
