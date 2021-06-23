package ext

import (
	"github.com/bytepowered/fluxgo/pkg/flux"
	"strings"
)

const (
	DefaultValueObjectResolverName = "default"
)

// ValueObjectResolver 将未定类型的值，按指定类型以及泛型类型转换为实际类型
type ValueObjectResolver func(value flux.ValueObject, toClass string, toGeneric []string) (actual interface{}, err error)

// WrapValueObjectResolver 包装转换函数
type WrapValueObjectResolver func(rawValue interface{}) (actual interface{}, err error)

func (resolve WrapValueObjectResolver) ResolveTo(value flux.ValueObject, _ string, _ []string) (actual interface{}, err error) {
	return resolve(value.Value)
}

var (
	objectResolvers = make(map[string]ValueObjectResolver, 16)
)

// RegisterObjectValueResolver 添加实际值类型解析函数
func RegisterObjectValueResolver(typeName string, resolver ValueObjectResolver) {
	typeName = flux.MustNotEmpty(typeName, "typeName is empty")
	typeName = strings.ToLower(typeName)
	objectResolvers[typeName] = resolver
}

// ValueObjectResolverByType 获取值类型解析函数；如果指定类型不存在，返回默认解析函数；
func ValueObjectResolverByType(typeName string) ValueObjectResolver {
	typeName = flux.MustNotEmpty(typeName, "typeName is empty")
	typeName = strings.ToLower(typeName)
	if r, ok := objectResolvers[typeName]; ok {
		return r
	} else {
		return objectResolvers[DefaultValueObjectResolverName]
	}
}

// 构建ValueObject的工具函数

func NewNilValueObject() flux.ValueObject {
	return NewObjectValueObject(nil)
}

func NewStringValueObject(value string) flux.ValueObject {
	return flux.ValueObject{Valid: value != "", Value: value, Encoding: flux.EncodingTypeGoString}
}

func NewObjectValueObject(value interface{}) flux.ValueObject {
	return flux.ValueObject{Valid: !flux.IsNil(value), Value: value, Encoding: flux.EncodingTypeGoObject}
}

func NewMapStringValueObject(value map[string]interface{}) flux.ValueObject {
	return flux.ValueObject{Valid: !flux.IsNil(value), Value: value, Encoding: flux.EncodingTypeGoMapString}
}

func NewListStringValueObject(value []string) flux.ValueObject {
	return flux.ValueObject{Valid: !flux.IsNil(value), Value: value, Encoding: flux.EncodingTypeGoListString}
}

func NewListObjectValueObject(value []interface{}) flux.ValueObject {
	return flux.ValueObject{Valid: !flux.IsNil(value), Value: value, Encoding: flux.EncodingTypeGoListObject}
}

func NewMapStringListValueObject(value map[string][]string) flux.ValueObject {
	return flux.ValueObject{Valid: !flux.IsNil(value), Value: value, Encoding: flux.EncodingTypeMapStringList}
}
