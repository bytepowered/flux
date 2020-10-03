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

// TypedValueResolver 将未定类型的值，按指定类型以及泛型类型转换
type TypedValueResolver func(typeClassName string, genericTypes []string, value interface{}) (interface{}, error)

// TypedValueResolveWrapper 包装转换函数
type TypedValueResolveWrapper func(value interface{}) (interface{}, error)

func (s TypedValueResolveWrapper) ResolveFunc(_ string, _ []string, value interface{}) (interface{}, error) {
	return s(value)
}
