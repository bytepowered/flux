package filter

import (
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/pkg"
)

func NewParameterParsingFilter() flux.Filter {
	return new(ParameterParsingFilter)
}

type ParameterParsingFilter int

func (ParameterParsingFilter) Invoke(next flux.FilterInvoker) flux.FilterInvoker {
	return func(ctx flux.Context) *flux.InvokeError {
		if err := resolve(ctx.Endpoint().Arguments, ctx); nil != err {
			return &flux.InvokeError{
				StatusCode: flux.StatusBadRequest,
				Message:    "PARAMETERS:RESOLVE",
				Internal:   err,
			}
		}
		return next(ctx)
	}
}

func (*ParameterParsingFilter) TypeId() string {
	return "ParameterParsingFilter"
}

////

func resolve(arguments []flux.Argument, ctx flux.Context) error {
	for _, p := range arguments {
		if flux.ArgumentTypePrimitive == p.Type {
			raw := _lookup(p, ctx)
			if v, err := _resolve(p.TypeClass, p.TypeGeneric, raw); nil != err {
				logger.Warnf("解析参数错误, class: %s, generic: %s, value: %+v, err: ", p.TypeClass, p.TypeGeneric, raw, err)
				return fmt.Errorf("endpoint argument resolve: arg.http=%s, class=[%s], generic=[%+v], error=%s",
					p.HttpKey, p.TypeClass, p.TypeGeneric, err)
			} else {
				p.HttpValue.SetValue(v)
			}
		} else if flux.ArgumentTypeComplex == p.Type {
			if err := resolve(p.Fields, ctx); nil != err {
				return err
			}
		} else {
			logger.Warnf("未支持的参数类型, class: %s, generic: %s, type: %s",
				p.TypeClass, p.TypeGeneric, p.Type)
		}
	}
	return nil
}

func _resolve(classType string, genericTypes []string, val interface{}) (interface{}, error) {
	resolver := pkg.GetValueResolver(classType)
	if nil == resolver {
		resolver = pkg.GetDefaultResolver()
	}
	return resolver(classType, genericTypes, val)
}

func _lookup(arg flux.Argument, ctx flux.Context) interface{} {
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
