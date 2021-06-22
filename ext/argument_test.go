package ext

import (
	"github.com/bytepowered/flux"
	assert2 "github.com/stretchr/testify/assert"
	"testing"
)

func TestArgumentDefinition(t *testing.T) {
	cases := []struct {
		definition flux.ServiceArgumentSpec
		class      string
		argType    string
		name       string
	}{
		{
			definition: NewStringArgument("str"),
			class:      flux.JavaLangStringClassName,
			argType:    flux.ServiceArgumentTypePrimitive,
			name:       "str",
		},
		{
			definition: NewIntegerArgument("int"),
			class:      flux.JavaLangIntegerClassName,
			argType:    flux.ServiceArgumentTypePrimitive,
			name:       "int",
		},
		{
			definition: NewLongArgument("long"),
			class:      flux.JavaLangLongClassName,
			argType:    flux.ServiceArgumentTypePrimitive,
			name:       "long",
		},
		{
			definition: NewBooleanArgument("boolean"),
			class:      flux.JavaLangBooleanClassName,
			argType:    flux.ServiceArgumentTypePrimitive,
			name:       "boolean",
		},
		{
			definition: NewFloatArgument("float32"),
			class:      flux.JavaLangFloatClassName,
			argType:    flux.ServiceArgumentTypePrimitive,
			name:       "float32",
		},
		{
			definition: NewDoubleArgument("float64"),
			class:      flux.JavaLangDoubleClassName,
			argType:    flux.ServiceArgumentTypePrimitive,
			name:       "float64",
		},
		{
			definition: NewSliceArrayArgument("slice-empty", flux.JavaLangStringClassName),
			class:      flux.JavaUtilListClassName,
			argType:    flux.ServiceArgumentTypePrimitive,
			name:       "slice-empty",
		},
		{
			definition: NewStringMapArgument("stringmap"),
			class:      flux.JavaUtilMapClassName,
			argType:    flux.ServiceArgumentTypeComplex,
			name:       "stringmap",
		},
		{
			definition: NewHashMapArgument("hashmap"),
			class:      flux.JavaUtilMapClassName,
			argType:    flux.ServiceArgumentTypeComplex,
			name:       "hashmap",
		},
	}
	assert := assert2.New(t)
	for _, tcase := range cases {
		assert.Equal(tcase.class, tcase.definition.ClassType, "type class", tcase.name)
		assert.Equal(tcase.argType, tcase.definition.StructType, "arg type", tcase.name)
		assert.Equal(tcase.name, tcase.definition.Name, "name")
	}
}
