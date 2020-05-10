package filter

import (
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/pkg"
)

func NewArgumentValueLookupFilter() flux.Filter {
	return new(ArgumentValueLookupFilter)
}

// 参数值查找Filter
type ArgumentValueLookupFilter int

func (ArgumentValueLookupFilter) Invoke(next flux.FilterInvoker) flux.FilterInvoker {
	lookup := ext.GetArgumentLookupFunc()
	return func(ctx flux.Context) *flux.InvokeError {
		if err := lookupResolve(lookup, ctx.Endpoint().Arguments, ctx); nil != err {
			return &flux.InvokeError{
				StatusCode: flux.StatusBadRequest,
				Message:    "PARAMETERS:LOOKUP",
				Internal:   err,
			}
		}
		return next(ctx)
	}
}

func (*ArgumentValueLookupFilter) TypeId() string {
	return "ArgumentValueLookupFilter"
}

func lookupResolve(lookup ext.ArgumentLookupFunc, arguments []flux.Argument, ctx flux.Context) error {
	for _, p := range arguments {
		if flux.ArgumentTypePrimitive == p.Type {
			raw := lookup(p, ctx)
			if v, err := _resolve(p.TypeClass, p.TypeGeneric, raw); nil != err {
				logger.Warnf("解析参数值错误, class: %s, generic: %s, value: %+v, err: ", p.TypeClass, p.TypeGeneric, raw, err)
				return fmt.Errorf("endpoint argument resolve: arg.http=%s, class=[%s], generic=[%+v], error=%s",
					p.HttpKey, p.TypeClass, p.TypeGeneric, err)
			} else {
				p.HttpValue.SetValue(v)
			}
		} else if flux.ArgumentTypeComplex == p.Type {
			if err := lookupResolve(lookup, p.Fields, ctx); nil != err {
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
