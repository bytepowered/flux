package common

import (
	"github.com/bytepowered/fluxgo/pkg/flux"
	"github.com/bytepowered/fluxgo/pkg/internal"
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
			class:      internal.JavaLangStringClassName,
			argType:    flux.ServiceArgumentTypePrimitive,
			name:       "str",
		},
		{
			definition: NewIntegerArgument("int"),
			class:      internal.JavaLangIntegerClassName,
			argType:    flux.ServiceArgumentTypePrimitive,
			name:       "int",
		},
		{
			definition: NewLongArgument("long"),
			class:      internal.JavaLangLongClassName,
			argType:    flux.ServiceArgumentTypePrimitive,
			name:       "long",
		},
		{
			definition: NewBooleanArgument("boolean"),
			class:      internal.JavaLangBooleanClassName,
			argType:    flux.ServiceArgumentTypePrimitive,
			name:       "boolean",
		},
		{
			definition: NewFloatArgument("float32"),
			class:      internal.JavaLangFloatClassName,
			argType:    flux.ServiceArgumentTypePrimitive,
			name:       "float32",
		},
		{
			definition: NewDoubleArgument("float64"),
			class:      internal.JavaLangDoubleClassName,
			argType:    flux.ServiceArgumentTypePrimitive,
			name:       "float64",
		},
		{
			definition: NewSliceArrayArgument("slice-empty", internal.JavaLangStringClassName),
			class:      internal.JavaUtilListClassName,
			argType:    flux.ServiceArgumentTypePrimitive,
			name:       "slice-empty",
		},
		{
			definition: NewStringMapArgument("stringmap"),
			class:      internal.JavaUtilMapClassName,
			argType:    flux.ServiceArgumentTypeComplex,
			name:       "stringmap",
		},
		{
			definition: NewHashMapArgument("hashmap"),
			class:      internal.JavaUtilMapClassName,
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
