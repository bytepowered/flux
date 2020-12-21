package flux

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

const (
	ValueMediaTypeGoObject          = "go:object"
	ValueMediaTypeGoString          = "go:string"
	ValueMediaTypeGoStringList      = "go:string-list"
	ValueMediaTypeGoStringMap       = "go:string-map"
	ValueMediaTypeGoStringValuesMap = "go:string-list-map"
)

// MTValue 包含指示值的媒体类型和Value结构
type MTValue struct {
	// 原始值类型
	Value interface{}
	// 媒体类型
	MediaType string
}

func WrapStringMTValue(value string) MTValue {
	return MTValue{Value: value, MediaType: ValueMediaTypeGoString}
}

func WrapObjectMTValue(value interface{}) MTValue {
	return MTValue{Value: value, MediaType: ValueMediaTypeGoObject}
}

func WrapStrMapMTValue(value map[string]interface{}) MTValue {
	return MTValue{Value: value, MediaType: ValueMediaTypeGoStringMap}
}

func WrapStrListMTValue(value []string) MTValue {
	return MTValue{Value: value, MediaType: ValueMediaTypeGoStringList}
}

func WrapStrValuesMapMTValue(value map[string][]string) MTValue {
	return MTValue{Value: value, MediaType: ValueMediaTypeGoStringValuesMap}
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
