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
	ValueMIMETypeGoText      = "go:text"
	ValueMIMETypeGoObject    = "go:object"
	ValueMIMETypeGoStringMap = "go:string-map"
)

// MIMEValue 包含值类型信息的Value包装结构
type MIMEValue struct {
	Value    interface{}
	MIMEType string
}

func WrapTextMIMEValue(value string) MIMEValue {
	return MIMEValue{Value: value, MIMEType: ValueMIMETypeGoText}
}

func WrapObjectMIMEValue(value interface{}) MIMEValue {
	return MIMEValue{Value: value, MIMEType: ValueMIMETypeGoObject}
}

func WrapStrMapMIMEValue(value map[string]interface{}) MIMEValue {
	return MIMEValue{Value: value, MIMEType: ValueMIMETypeGoStringMap}
}

// TypedValueResolver 将未定类型的值，按指定类型以及泛型类型转换
// @param typeClass 值类型
// @param typeGeneric 值泛型类型
// @param mimeValue Http请求带类型的值
type TypedValueResolver func(class string, generics []string, mimeValue MIMEValue) (value interface{}, err error)

// TypedValueResolveWrapper 包装转换函数
type TypedValueResolveWrapper func(in interface{}) (value interface{}, err error)

func (f TypedValueResolveWrapper) ResolveMIME(_ string, _ []string, mimeValue MIMEValue) (value interface{}, err error) {
	return f(mimeValue.Value)
}
