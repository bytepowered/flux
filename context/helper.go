package context

import (
	"errors"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/backend"
	"github.com/bytepowered/flux/pkg"
)

// LookupExpr 搜索LookupExpr表达式指定域的值。
func LookupExpr(expr string, ctx flux.Context) (interface{}, error) {
	if "" == expr || nil == ctx {
		return nil, errors.New("empty lookup expr, or context is nil")
	}
	scope, key, ok := pkg.LookupParseExpr(expr)
	if !ok {
		return "", errors.New("illegal lookup expr: " + expr)
	}
	mtv, err := backend.DefaultArgumentLookupFunc(scope, key, ctx)
	if nil != err {
		return "", err
	}
	return mtv.Value, nil
}
