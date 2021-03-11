package context

import (
	"errors"
	"github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-node/backend"
	"github.com/bytepowered/flux/flux-pkg"
)

// LookupExpr 搜索LookupExpr表达式指定域的值。
func LookupExpr(expr string, ctx flux.Context) (interface{}, error) {
	if "" == expr || nil == ctx {
		return nil, errors.New("empty lookup expr, or context is nil")
	}
	scope, key, ok := fluxpkg.LookupParseExpr(expr)
	if !ok {
		return "", errors.New("illegal lookup expr: " + expr)
	}
	mtv, err := backend.DefaultArgumentLookupFunc(scope, key, ctx)
	if nil != err {
		return "", err
	}
	return mtv.Value, nil
}
