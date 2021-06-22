package transporter

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
)

// Resolve 解析Argument参数值
func Resolve(ctx *flux.Context, a *flux.ServiceArgumentSpec) (interface{}, error) {
	switch a.StructType {
	case flux.ServiceArgumentTypeComplex:
		return resolvepojo(ctx, a)

	case flux.ServiceArgumentTypePrimitive:
		return resolvefield(ctx, a)

	default:
		return resolveobj(ctx, a)
	}
}

// resolvepojo 解析POJO，按字段声明的复杂对象
func resolvepojo(ctx *flux.Context, a *flux.ServiceArgumentSpec) (interface{}, error) {
	sm := make(map[string]interface{}, len(a.Fields))
	sm["class"] = a.ClassType
	for _, field := range a.Fields {
		if fv, err := Resolve(ctx, &field); nil != err {
			return nil, err
		} else {
			sm[field.Name] = fv
		}
	}
	return sm, nil
}

// resolvefield 解析字段值
func resolvefield(ctx *flux.Context, a *flux.ServiceArgumentSpec) (interface{}, error) {
	resolver := ext.MTValueResolverByType(a.ClassType)
	// 1: By Value loader
	if loader, ok := ext.GetArgumentValueLoader(a); ok && nil != loader {
		mtv := loader()
		return resolver(mtv, a.ClassType, a.GenericTypes)
	}
	// 2: Lookup and resolve
	mtv, err := ext.LookupScopedValueFunc()(ctx, a.HttpScope, a.HttpName)
	if nil != err {
		return nil, err
	}
	// 3: Default value
	if !mtv.Valid && a.Annotations != nil {
		if anno, ok := a.Annotations.AnnotationEx(flux.ServiceArgumentAnnotationTagDefault); ok {
			mtv = flux.NewStringMTValue(anno.ToString())
		}
	}
	return resolver(mtv, a.ClassType, a.GenericTypes)
}

// resolveobj 解析Map对象
func resolveobj(ctx *flux.Context, a *flux.ServiceArgumentSpec) (interface{}, error) {
	mtv, err := ext.LookupScopedValueFunc()(ctx, a.HttpScope, a.HttpName)
	if nil != err {
		return nil, err
	}
	// 使用Default解析器
	resolver := ext.MTValueResolverByType(ext.DefaultMTValueResolverName)
	return resolver(mtv, a.ClassType, a.GenericTypes)
}
