package ext

import (
	flux2 "github.com/bytepowered/flux/flux-node"
	assert2 "github.com/stretchr/testify/assert"
	"testing"
)

func TestArgumentDefinition(t *testing.T) {
	cases := []struct {
		definition flux2.Argument
		class      string
		argType    string
		name       string
	}{
		{
			definition: NewStringArgument("str"),
			class:      flux2.JavaLangStringClassName,
			argType:    flux2.ArgumentTypePrimitive,
			name:       "str",
		},
		{
			definition: NewIntegerArgument("int"),
			class:      flux2.JavaLangIntegerClassName,
			argType:    flux2.ArgumentTypePrimitive,
			name:       "int",
		},
		{
			definition: NewLongArgument("long"),
			class:      flux2.JavaLangLongClassName,
			argType:    flux2.ArgumentTypePrimitive,
			name:       "long",
		},
		{
			definition: NewBooleanArgument("boolean"),
			class:      flux2.JavaLangBooleanClassName,
			argType:    flux2.ArgumentTypePrimitive,
			name:       "boolean",
		},
		{
			definition: NewFloatArgument("float32"),
			class:      flux2.JavaLangFloatClassName,
			argType:    flux2.ArgumentTypePrimitive,
			name:       "float32",
		},
		{
			definition: NewDoubleArgument("float64"),
			class:      flux2.JavaLangDoubleClassName,
			argType:    flux2.ArgumentTypePrimitive,
			name:       "float64",
		},
		{
			definition: NewSliceArrayArgument("slice-empty", flux2.JavaLangStringClassName),
			class:      flux2.JavaUtilListClassName,
			argType:    flux2.ArgumentTypePrimitive,
			name:       "slice-empty",
		},
		{
			definition: NewStringMapArgument("stringmap"),
			class:      flux2.JavaUtilMapClassName,
			argType:    flux2.ArgumentTypeComplex,
			name:       "stringmap",
		},
		{
			definition: NewHashMapArgument("hashmap"),
			class:      flux2.JavaUtilMapClassName,
			argType:    flux2.ArgumentTypeComplex,
			name:       "hashmap",
		},
	}
	assert := assert2.New(t)
	for _, tcase := range cases {
		assert.Equal(tcase.class, tcase.definition.Class, "type class", tcase.name)
		assert.Equal(tcase.argType, tcase.definition.Type, "arg type", tcase.name)
		assert.Equal(tcase.name, tcase.definition.Name, "name")
	}
}
