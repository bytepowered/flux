package common

import (
	"errors"
	"github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-pkg"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"
)

// LookupMTValueByExpr 搜索LookupExpr表达式指定域的值。
func LookupMTValueByExpr(expr string, ctx *flux.Context) (interface{}, error) {
	if expr == "" || nil == ctx {
		return nil, errors.New("empty lookup expr, or context is nil")
	}
	scope, key, ok := fluxpkg.ParseScopeExpr(expr)
	if !ok {
		return "", errors.New("illegal lookup expr: " + expr)
	}
	mtv, err := LookupMTValue(scope, key, ctx)
	if nil != err {
		return "", err
	}
	return mtv.Value, nil
}

// LookupMTValue 根据指定 Scope/Key 查找参数值；返回MTValue对象，包含值对象的MediaType类型；
func LookupMTValue(scope, key string, ctx *flux.Context) (value flux.MTValue, err error) {
	if scope == "" || key == "" {
		return flux.NewInvalidMTValue(), errors.New("lookup empty scope or key, scope: " + scope + ", key: " + key)
	}
	if nil == ctx {
		return flux.NewInvalidMTValue(), errors.New("lookup nil context")
	}
	switch strings.ToUpper(scope) {
	case flux.ScopePath:
		return lookupValues(ctx.PathVars(), key), nil
	case flux.ScopePathMap:
		return flux.NewMapStringListMTValue(ctx.PathVars()), nil
	case flux.ScopeQuery:
		return lookupValues(ctx.QueryVars(), key), nil
	case flux.ScopeQueryMulti:
		return flux.NewListStringMTValue(ctx.QueryVars()[key]), nil
	case flux.ScopeQueryMap:
		return flux.NewMapStringListMTValue(ctx.QueryVars()), nil
	case flux.ScopeForm:
		return lookupValues(ctx.FormVars(), key), nil
	case flux.ScopeFormMap:
		return flux.NewMapStringListMTValue(ctx.FormVars()), nil
	case flux.ScopeFormMulti:
		return flux.NewListStringMTValue(ctx.FormVars()[key]), nil
	case flux.ScopeHeader:
		return lookupValues(ctx.HeaderVars(), key), nil
	case flux.ScopeHeaderMap:
		return flux.NewMapStringListMTValue(ctx.HeaderVars()), nil
	case flux.ScopeAttr:
		if v, ok := ctx.GetAttribute(key); ok {
			return ToMTValue(v), nil
		}
		return flux.NewInvalidMTValue(), nil
	case flux.ScopeAttrs:
		return flux.NewMapStringMTValue(ctx.Attributes()), nil
	case flux.ScopeBody:
		reader, err := ctx.BodyReader()
		hct := ctx.HeaderVar(flux.HeaderContentType)
		return flux.MTValue{Valid: err == nil, Value: reader, MediaType: flux.MediaType(hct)}, err
	case flux.ScopeParam:
		v, _ := fluxpkg.LookupByProviders(key, ctx.QueryVars, ctx.FormVars)
		return flux.NewStringMTValue(v), nil
	case flux.ScopeRequest:
		switch strings.ToUpper(key) {
		case "METHOD":
			return flux.NewStringMTValue(ctx.Method()), nil
		case "URI":
			return flux.NewStringMTValue(ctx.URI()), nil
		case "HOST":
			return flux.NewStringMTValue(ctx.Host()), nil
		case "REMOTEADDR":
			return flux.NewStringMTValue(ctx.RemoteAddr()), nil
		default:
			return flux.NewInvalidMTValue(), nil
		}
	default:
		if v, ok := fluxpkg.LookupByProviders(key, ctx.PathVars, ctx.QueryVars, ctx.FormVars); ok {
			return flux.NewStringMTValue(v), nil
		}
		if mtv := lookupValues(ctx.HeaderVars(), key); mtv.Valid {
			return mtv, nil
		}
		if v, ok := ctx.GetAttribute(key); ok {
			return ToMTValue(v), nil
		}
		return flux.NewInvalidMTValue(), nil
	}
}

func ToMTValue(v interface{}) flux.MTValue {
	switch v.(type) {
	case string:
		return flux.NewStringMTValue(v.(string))
	case map[string]interface{}:
		return flux.NewMapStringMTValue(v.(map[string]interface{}))
	case map[string][]string:
		return flux.NewMapStringListMTValue(v.(map[string][]string))
	case []string:
		return flux.NewListStringMTValue(v.([]string))
	default:
		return flux.NewObjectMTValue(v)
	}
}

func lookupValues(values interface{}, key string) flux.MTValue {
	var value []string
	var ok bool
	switch values.(type) {
	case url.Values:
		value, ok = values.(url.Values)[key]
	case http.Header:
		// Header: key case insensitive
		value, ok = values.(http.Header)[textproto.CanonicalMIMEHeaderKey(key)]
	case map[string][]string:
		value, ok = values.(map[string][]string)[key]
	default:
		ok = false
	}
	if ok {
		switch {
		case len(value) == 1:
			return flux.NewStringMTValue(value[0])
		case len(value) > 1:
			copied := make([]string, len(value))
			copy(copied, value)
			return flux.NewListStringMTValue(copied)
		default:
			return flux.NewStringMTValue("")
		}
	} else {
		return flux.NewInvalidMTValue()
	}
}
