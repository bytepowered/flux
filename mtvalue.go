package flux

import "strings"

const (
	JavaLangStringClassName  = "java.lang.String"
	JavaLangIntegerClassName = "java.lang.Integer"
	JavaLangLongClassName    = "java.lang.Long"
	JavaLangFloatClassName   = "java.lang.Float"
	JavaLangDoubleClassName  = "java.lang.Double"
	JavaLangBooleanClassName = "java.lang.Boolean"
	JavaUtilMapClassName     = "java.util.Map"
	JavaUtilListClassName    = "java.util.List"
)

// Golang内置参数类型
type MediaType string

func (m MediaType) Contains(s string) bool {
	return strings.Contains(string(m), s)
}

const (
	MediaTypeGoObject        = MediaType("go:object")
	MediaTypeGoString        = MediaType("go:string")
	MediaTypeGoListString    = MediaType("go:[]string")
	MediaTypeGoMapString     = MediaType("go:map[string]object")
	MediaTypeGoMapStringList = MediaType("go:map[string][]string")
)

// MTValue 包含指示值的媒体类型和Value结构
type MTValue struct {
	Valid     bool        // 是否有效
	Value     interface{} // 原始值类型
	MediaType MediaType   // 数据媒体类型
}

func NewInvalidMTValue() MTValue {
	return NewObjectMTValue(nil)
}

func NewStringMTValue(value string) MTValue {
	return MTValue{Valid: value != "", Value: value, MediaType: MediaTypeGoString}
}

func NewObjectMTValue(value interface{}) MTValue {
	return MTValue{Valid: value != nil, Value: value, MediaType: MediaTypeGoObject}
}

func NewMapStringMTValue(value map[string]interface{}) MTValue {
	return MTValue{Valid: value != nil, Value: value, MediaType: MediaTypeGoMapString}
}

func NewListStringMTValue(value []string) MTValue {
	return MTValue{Valid: value != nil, Value: value, MediaType: MediaTypeGoListString}
}

func NewMapStringListMTValue(value map[string][]string) MTValue {
	return MTValue{Valid: value != nil, Value: value, MediaType: MediaTypeGoMapStringList}
}

// MTValueResolver 将未定类型的值，按指定类型以及泛型类型转换为实际类型
// @param mtValue Http请求指示媒体类型的值
// @param toClass 目标值类型
// @param toGeneric 目标值泛型类型
type MTValueResolver func(mtValue MTValue, toClass string, toGeneric []string) (actualValue interface{}, err error)

// WrapMTValueResolver 包装转换函数
type WrapMTValueResolver func(rawValue interface{}) (actualValue interface{}, err error)

func (resolve WrapMTValueResolver) ResolveMT(mtValue MTValue, _ string, _ []string) (actualValue interface{}, err error) {
	return resolve(mtValue.Value)
}
