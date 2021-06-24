package ext

import (
	"github.com/bytepowered/fluxgo/pkg/flux"
	"strings"
)

const (
	DefaultEncodeValueResolverName = "default"
)

// EncodeValueResolver 将未定类型的值，按指定类型以及泛型类型转换为实际类型
type EncodeValueResolver func(value flux.EncodeValue, toClass string, toGeneric []string) (actual interface{}, err error)

// WrapEncodeValueResolver 包装转换函数
type WrapEncodeValueResolver func(rawValue interface{}) (actual interface{}, err error)

func (resolve WrapEncodeValueResolver) ResolveTo(value flux.EncodeValue, _ string, _ []string) (actual interface{}, err error) {
	return resolve(value.Value)
}

var (
	objectResolvers = make(map[string]EncodeValueResolver, 16)
)

// RegisterEncodeValueResolver 添加实际值类型解析函数
func RegisterEncodeValueResolver(typeName string, resolver EncodeValueResolver) {
	typeName = flux.MustNotEmpty(typeName, "typeName is empty")
	typeName = strings.ToLower(typeName)
	objectResolvers[typeName] = resolver
}

// EncodeValueResolverByType 获取值类型解析函数；如果指定类型不存在，返回默认解析函数；
func EncodeValueResolverByType(typeName string) EncodeValueResolver {
	typeName = flux.MustNotEmpty(typeName, "typeName is empty")
	typeName = strings.ToLower(typeName)
	if r, ok := objectResolvers[typeName]; ok {
		return r
	} else {
		return objectResolvers[DefaultEncodeValueResolverName]
	}
}

// 构建EncodeValue的工具函数

func NewNilEncodeValue() flux.EncodeValue {
	return NewObjectEncodeValue(nil)
}

func NewNumberEncodeValueOf(value float64, valid bool) flux.EncodeValue {
	return flux.NewEncodeValueWith(value, flux.EncodingTypeGoNumber, func() bool {
		return valid
	})
}

func NewNumberEncodeValue(value float64) flux.EncodeValue {
	return NewNumberEncodeValueOf(value, true)
}

func NewStringEncodeValue(value string) flux.EncodeValue {
	return flux.NewEncodeValueWith(value, flux.EncodingTypeGoString, func() bool {
		return "" != value
	})
}

func NewObjectEncodeValue(value interface{}) flux.EncodeValue {
	return flux.NewEncodeValueWith(value, flux.EncodingTypeGoObject, func() bool {
		return !flux.IsNil(value)
	})
}

func NewMapStringEncodeValue(value map[string]interface{}) flux.EncodeValue {
	return flux.NewEncodeValueWith(value, flux.EncodingTypeGoMapString, func() bool {
		return !flux.IsNil(value)
	})
}

func NewListStringEncodeValue(value []string) flux.EncodeValue {
	return flux.NewEncodeValueWith(value, flux.EncodingTypeGoListString, func() bool {
		return !flux.IsNil(value)
	})
}

func NewListObjectEncodeValue(value []interface{}) flux.EncodeValue {
	return flux.NewEncodeValueWith(value, flux.EncodingTypeGoListObject, func() bool {
		return !flux.IsNil(value)
	})
}

func NewMapStringListEncodeValue(value map[string][]string) flux.EncodeValue {
	return flux.NewEncodeValueWith(value, flux.EncodingTypeMapStringList, func() bool {
		return !flux.IsNil(value)
	})
}
