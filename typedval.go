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
	ValueMIMETypeLangText      = "go:text"
	ValueMIMETypeLangObject    = "go:object"
	ValueMIMETypeLangStringMap = "go:string-map"
)

// MIMETypeValue
type MIMETypeValue struct {
	Value    interface{}
	MIMEType string
}

// TypedValueResolver 将未定类型的值，按指定类型以及泛型类型转换
// @param typeClass 值类型
// @param typeGeneric 值泛型类型
// @param value Http请求的值
type TypedValueResolver func(typeClass string, typeGenerics []string, value MIMETypeValue) (typedValue interface{}, err error)

// TypedValueResolveWrapper 包装转换函数
type TypedValueResolveWrapper func(value interface{}) (typedValue interface{}, err error)

func (s TypedValueResolveWrapper) ResolveFunc(_ string, _ []string, value MIMETypeValue) (typedValue interface{}, err error) {
	return s(value.Value)
}
