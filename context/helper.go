package context

import (
	"errors"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/backend"
	"github.com/bytepowered/flux/pkg"
)

// LookupContextByExpr 搜索LookupExpr表达式指定域的值。
func LookupContextByExpr(lookupExpr string, ctx flux.Context) (interface{}, error) {
	if "" == lookupExpr || nil == ctx {
		return nil, errors.New("empty lookup expr or context")
	}
	scope, key, ok := pkg.LookupParseExpr(lookupExpr)
	if !ok {
		return "", errors.New("illegal lookup expr: " + lookupExpr)
	}
	mtv, err := backend.DefaultArgumentValueLookupFunc(scope, key, ctx)
	if nil != err {
		return "", err
	}
	return mtv.Value, nil
}
