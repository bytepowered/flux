package pkg

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

type ValueResolver func(typeName string, genericTypes []string, value interface{}) (interface{}, error)

type SimpleValueResolver func(value interface{}) (interface{}, error)

func (s SimpleValueResolver) ResolveFunc(_ string, _ []string, value interface{}) (interface{}, error) {
	return s(value)
}
