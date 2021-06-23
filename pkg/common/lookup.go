package common

import (
	"errors"
	"github.com/bytepowered/fluxgo/pkg/ext"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"
)

import (
	"github.com/bytepowered/fluxgo/pkg/flux"
	"github.com/bytepowered/fluxgo/pkg/toolkit"
)

// LookupValueByExpr 搜索LookupExpr表达式指定域的值。
func LookupValueByExpr(ctx *flux.Context, expr string) (interface{}, error) {
	if expr == "" || nil == ctx {
		return nil, errors.New("empty lookup expr, or context is nil")
	}
	scope, key, ok := toolkit.ParseScopeExpr(expr)
	if !ok {
		return "", errors.New("illegal lookup expr: " + expr)
	}
	mtv, err := LookupValueByScoped(ctx, scope, key)
	if nil != err {
		return "", err
	}
	return mtv.Value, nil
}

// LookupValueByScoped 根据Scope,Key从Context中查找参数；支持复杂参数类型
func LookupValueByScoped(ctx *flux.Context, scope, key string) (value flux.ValueObject, err error) {
	if scope == "" || key == "" {
		return ext.NewNilValueObject(), errors.New("lookup empty scope or key, scope: " + scope + ", key: " + key)
	}
	if nil == ctx {
		return ext.NewNilValueObject(), errors.New("lookup nil context")
	}
	switch strings.ToUpper(scope) {
	case flux.ScopePath:
		return lookupValues(ctx.PathVars(), key), nil
	case flux.ScopePathMap:
		return ext.NewMapStringListValueObject(ctx.PathVars()), nil
	case flux.ScopeQuery:
		return lookupValues(ctx.QueryVars(), key), nil
	case flux.ScopeQueryMulti:
		return ext.NewListStringValueObject(ctx.QueryVars()[key]), nil
	case flux.ScopeQueryMap:
		return ext.NewMapStringListValueObject(ctx.QueryVars()), nil
	case flux.ScopeForm:
		return lookupValues(ctx.FormVars(), key), nil
	case flux.ScopeFormMap:
		return ext.NewMapStringListValueObject(ctx.FormVars()), nil
	case flux.ScopeFormMulti:
		return ext.NewListStringValueObject(ctx.FormVars()[key]), nil
	case flux.ScopeHeader:
		return lookupValues(ctx.HeaderVars(), key), nil
	case flux.ScopeHeaderMap:
		return ext.NewMapStringListValueObject(ctx.HeaderVars()), nil
	case flux.ScopeAttr:
		if v, ok := ctx.AttributeEx(key); ok {
			return ToMTValue(v), nil
		}
		return ext.NewNilValueObject(), nil
	case flux.ScopeAttrs:
		return ext.NewMapStringValueObject(ctx.Attributes()), nil
	case flux.ScopeBody:
		reader, err := ctx.BodyReader()
		hct := ctx.HeaderVar(flux.HeaderContentType)
		return flux.ValueObject{Valid: err == nil, Value: reader, Encoding: flux.EncodingType(hct)}, err
	case flux.ScopeParam:
		v, _ := LookupValues(key, ctx.QueryVars, ctx.FormVars)
		return ext.NewStringValueObject(v), nil
	case flux.ScopeRequest:
		switch strings.ToUpper(key) {
		case "METHOD":
			return ext.NewStringValueObject(ctx.Method()), nil
		case "URI":
			return ext.NewStringValueObject(ctx.URI()), nil
		case "HOST":
			return ext.NewStringValueObject(ctx.Host()), nil
		case "REMOTEADDR":
			return ext.NewStringValueObject(ctx.RemoteAddr()), nil
		default:
			return ext.NewNilValueObject(), nil
		}
	default:
		if v, ok := LookupValues(key, ctx.PathVars, ctx.QueryVars, ctx.FormVars); ok {
			return ext.NewStringValueObject(v), nil
		}
		if mtv := lookupValues(ctx.HeaderVars(), key); mtv.Valid {
			return mtv, nil
		}
		if v, ok := ctx.AttributeEx(key); ok {
			return ToMTValue(v), nil
		}
		return ext.NewNilValueObject(), nil
	}
}

func ToMTValue(v interface{}) flux.ValueObject {
	switch v.(type) {
	case string:
		return ext.NewStringValueObject(v.(string))
	case map[string]interface{}:
		return ext.NewMapStringValueObject(v.(map[string]interface{}))
	case map[string][]string:
		return ext.NewMapStringListValueObject(v.(map[string][]string))
	case []string:
		return ext.NewListStringValueObject(v.([]string))
	case []interface{}:
		return ext.NewListObjectValueObject(v.([]interface{}))
	default:
		return ext.NewObjectValueObject(v)
	}
}

func LookupValues(key string, providers ...func() url.Values) (string, bool) {
	for _, fun := range providers {
		values := fun()
		if v, ok := values[key]; ok {
			return v[0], true
		}
	}
	return "", false
}

func lookupValues(values interface{}, key string) flux.ValueObject {
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
			return ext.NewStringValueObject(value[0])
		case len(value) > 1:
			copied := make([]string, len(value))
			copy(copied, value)
			return ext.NewListStringValueObject(copied)
		default:
			return ext.NewStringValueObject("")
		}
	} else {
		return ext.NewNilValueObject()
	}
}
