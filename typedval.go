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

func NewTextTypedValue(value string) MIMEValue {
	return MIMEValue{Value: value, MIMEType: ValueMIMETypeGoText}
}

func NewObjectTypedValue(value interface{}) MIMEValue {
	return MIMEValue{Value: value, MIMEType: ValueMIMETypeGoObject}
}

func NewStrMapTypedValue(value map[string]interface{}) MIMEValue {
	return MIMEValue{Value: value, MIMEType: ValueMIMETypeGoStringMap}
}

// TypedValueResolver 将未定类型的值，按指定类型以及泛型类型转换
// @param typeClass 值类型
// @param typeGeneric 值泛型类型
// @param value Http请求的值
type TypedValueResolver func(typeClass string, typeGenerics []string, value MIMEValue) (typedValue interface{}, err error)

// TypedValueResolveWrapper 包装转换函数
type TypedValueResolveWrapper func(value interface{}) (typedValue interface{}, err error)

func (f TypedValueResolveWrapper) ResolveFunc(_ string, _ []string, value MIMEValue) (typedValue interface{}, err error) {
	return f(value.Value)
}
