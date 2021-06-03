package transporter

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
)

// Resolve 解析Argument参数值
func Resolve(ctx *flux.Context, a *flux.Argument) (interface{}, error) {
	switch a.Type {
	case flux.ArgumentTypeComplex:
		return resolvepojo(ctx, a)

	case flux.ArgumentTypePrimitive:
		return resolvefield(ctx, a)

	default:
		return resolveobj(ctx, a)
	}
}

// resolvepojo 解析POJO，按字段声明的复杂对象
func resolvepojo(ctx *flux.Context, a *flux.Argument) (interface{}, error) {
	sm := make(map[string]interface{}, len(a.Fields))
	sm["class"] = a.Class
	for _, field := range a.Fields {
		if fv, err := Resolve(ctx, a); nil != err {
			return nil, err
		} else {
			sm[field.Name] = fv
		}
	}
	return sm, nil
}

// resolvefield 解析字段值
func resolvefield(ctx *flux.Context, a *flux.Argument) (interface{}, error) {
	resolver := ext.MTValueResolverByType(a.Class)
	// 1: By Value loader
	if nil != a.ValueLoader {
		mtv := a.ValueLoader()
		return resolver(mtv, a.Class, a.Generic)
	}
	// 2: Lookup and resolve
	mtv, err := ext.LookupFunc()(ctx, a.HttpScope, a.HttpName)
	if nil != err {
		return nil, err
	}
	// 3: Default value
	if !mtv.Valid && a.Attributes != nil {
		if attr, ok := a.Attributes.SingleEx(flux.ArgumentAttributeTagDefault); ok {
			mtv = flux.NewStringMTValue(attr.ToString())
		}
	}
	return resolver(mtv, a.Class, a.Generic)
}

// resolveobj 解析Map对象
func resolveobj(ctx *flux.Context, a *flux.Argument) (interface{}, error) {
	mtv, err := ext.LookupFunc()(ctx, a.HttpScope, a.HttpName)
	if nil != err {
		return nil, err
	}
	// 使用Default解析器
	resolver := ext.MTValueResolverByType(ext.DefaultMTValueResolverName)
	return resolver(mtv, a.Class, a.Generic)
}
