package ext

import (
	"github.com/bytepowered/flux"
	assert2 "github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func Test_SetNilArgumentValueLookupFunc(t *testing.T) {
	assert := assert2.New(t)
	assert.Panicsf(func() {
		StoreArgumentValueLookupFunc(nil)
	}, "Must panic")
}

func Test_SetGetArgumentValueLookupFunc(t *testing.T) {
	assert := assert2.New(t)
	f1 := func(scope, key string, context flux.Context) (value flux.MTValue, err error) {
		return flux.WrapStringMTValue("value"), nil
	}
	StoreArgumentValueLookupFunc(f1)
	f2 := LoadArgumentValueLookupFunc()
	assert.Equal(reflect.ValueOf(f1).Pointer(), reflect.ValueOf(f2).Pointer(), "Must equals func")
}

func TestNewPrimitiveArgument(t *testing.T) {
	cases := []struct {
		definition flux.Argument
		class      string
		argType    string
		name       string
	}{
		{
			definition: NewStringArgument("str"),
			class:      flux.JavaLangStringClassName,
			argType:    flux.ArgumentTypePrimitive,
			name:       "str",
		},
		{
			definition: NewIntegerArgument("int"),
			class:      flux.JavaLangIntegerClassName,
			argType:    flux.ArgumentTypePrimitive,
			name:       "int",
		},
		{
			definition: NewLongArgument("long"),
			class:      flux.JavaLangLongClassName,
			argType:    flux.ArgumentTypePrimitive,
			name:       "long",
		},
		{
			definition: NewBooleanArgument("boolean"),
			class:      flux.JavaLangBooleanClassName,
			argType:    flux.ArgumentTypePrimitive,
			name:       "boolean",
		},
		{
			definition: NewFloatArgument("float32"),
			class:      flux.JavaLangFloatClassName,
			argType:    flux.ArgumentTypePrimitive,
			name:       "float32",
		},
		{
			definition: NewDoubleArgument("float64"),
			class:      flux.JavaLangDoubleClassName,
			argType:    flux.ArgumentTypePrimitive,
			name:       "float64",
		},
		{
			definition: NewDoubleArgument("float64"),
			class:      flux.JavaLangDoubleClassName,
			argType:    flux.ArgumentTypePrimitive,
			name:       "float64",
		},
		{
			definition: NewStringMapArgument("stringmap"),
			class:      flux.JavaUtilMapClassName,
			argType:    flux.ArgumentTypeComplex,
			name:       "stringmap",
		},
		{
			definition: NewHashMapArgument("hashmap"),
			class:      flux.JavaUtilMapClassName,
			argType:    flux.ArgumentTypeComplex,
			name:       "hashmap",
		},
		{
			definition: NewSliceArrayArgument("slice-empty"),
			class:      flux.JavaUtilListClassName,
			argType:    flux.ArgumentTypeComplex,
			name:       "slice-empty",
		},
	}
	for _, tcase := range cases {
		assert := assert2.New(t)
		assert.Equal(tcase.class, tcase.definition.Class, "type class")
		assert.Equal(tcase.argType, tcase.definition.Type, "arg type")
		assert.Equal(tcase.name, tcase.definition.Name, "name")
	}
}
