package filter

import (
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/pkg"
	"net/http"
)

func NewArgumentValueLookupFilter() flux.Filter {
	return new(ArgumentValueLookupFilter)
}

// 参数值查找Filter
type ArgumentValueLookupFilter int

func (ArgumentValueLookupFilter) Invoke(next flux.FilterInvoker) flux.FilterInvoker {
	lookupFunc := ext.GetArgumentLookupFunc()
	return func(ctx flux.Context) *flux.InvokeError {
		// HEAD, OPTIONS 不需要解析参数
		method := ctx.RequestMethod()
		if http.MethodHead == method || http.MethodOptions == method {
			return next(ctx)
		}
		args := ctx.Endpoint().Arguments
		if 0 == len(args) {
			return next(ctx)
		}
		// 解析参数值
		if err := resolve(lookupFunc, ctx.Endpoint().Arguments, ctx); nil != err {
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

func resolve(lookupFunc ext.ArgumentLookupFunc, arguments []flux.Argument, ctx flux.Context) error {
	for _, p := range arguments {
		if flux.ArgumentTypePrimitive == p.Type {
			value, err := lookupFunc(p, ctx)
			if nil != err {
				return fmt.Errorf("argument lookup error: http.key=%s, class=%s[%s], error=%s", p.HttpKey, p.TypeClass, p.TypeGeneric, err)
			}
			if v, err := _resolve(p.TypeClass, p.TypeGeneric, value); nil != err {
				logger.Trace(ctx.RequestId()).Warnw("Failed to resolve argument",
					"class", p.TypeClass, "generic", p.TypeGeneric, "value", value, "error", err)
				return fmt.Errorf("argument resolve error: http.key=%s, class=%s[%s], error=%s", p.HttpKey, p.TypeClass, p.TypeGeneric, err)
			} else {
				p.HttpValue.SetValue(v)
			}
		} else if flux.ArgumentTypeComplex == p.Type {
			if err := resolve(lookupFunc, p.Fields, ctx); nil != err {
				return err
			}
		} else {
			logger.Trace(ctx.RequestId()).Warnw("Unsupported argument type",
				"class", p.TypeClass, "generic", p.TypeGeneric, "type", p.Type)
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
