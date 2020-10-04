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
		SetArgumentValueLookupFunc(nil)
	}, "Must panic")
}

func Test_SetGetArgumentValueLookupFunc(t *testing.T) {
	assert := assert2.New(t)
	f1 := func(scope, key string, context flux.Context) (value interface{}, err error) {
		return "value", nil
	}
	SetArgumentValueLookupFunc(f1)
	f2 := GetArgumentValueLookupFunc()
	assert.Equal(reflect.ValueOf(f1).Pointer(), reflect.ValueOf(f2).Pointer(), "Must equals func")
}

func TestNewPrimitiveArgument(t *testing.T) {
	cases := []struct {
		definition flux.Argument
		class      string
		argType    string
		name       string
		value      interface{}
	}{
		{
			definition: NewStringArgument("str", "001"),
			class:      flux.JavaLangStringClassName,
			argType:    flux.ArgumentTypePrimitive,
			name:       "str",
			value:      "001",
		},
		{
			definition: NewIntegerArgument("int", 1),
			class:      flux.JavaLangIntegerClassName,
			argType:    flux.ArgumentTypePrimitive,
			name:       "int",
			value:      1,
		},
		{
			definition: NewLongArgument("long", int64(2)),
			class:      flux.JavaLangLongClassName,
			argType:    flux.ArgumentTypePrimitive,
			name:       "long",
			value:      int64(2),
		},
		{
			definition: NewBooleanArgument("boolean", true),
			class:      flux.JavaLangBooleanClassName,
			argType:    flux.ArgumentTypePrimitive,
			name:       "boolean",
			value:      true,
		},
		{
			definition: NewFloatArgument("float32", float32(12.3456)),
			class:      flux.JavaLangFloatClassName,
			argType:    flux.ArgumentTypePrimitive,
			name:       "float32",
			value:      float32(12.3456),
		},
		{
			definition: NewDoubleArgument("float64", float64(12.345678)),
			class:      flux.JavaLangDoubleClassName,
			argType:    flux.ArgumentTypePrimitive,
			name:       "float64",
			value:      float64(12.345678),
		},
		{
			definition: NewDoubleArgument("float64", float64(12.345678)),
			class:      flux.JavaLangDoubleClassName,
			argType:    flux.ArgumentTypePrimitive,
			name:       "float64",
			value:      float64(12.345678),
		},
		{
			definition: NewStringMapArgument("stringmap", map[string]interface{}{"key": "value"}),
			class:      flux.JavaUtilMapClassName,
			argType:    flux.ArgumentTypeComplex,
			name:       "stringmap",
			value:      map[string]interface{}{"key": "value"},
		},
		{
			definition: NewHashMapArgument("hashmap", map[string]interface{}{"hash": "map"}),
			class:      flux.JavaUtilMapClassName,
			argType:    flux.ArgumentTypeComplex,
			name:       "hashmap",
			value:      map[string]interface{}{"hash": "map"},
		},
		{
			definition: NewHashMapArgument("nil-hashmap", nil),
			class:      flux.JavaUtilMapClassName,
			argType:    flux.ArgumentTypeComplex,
			name:       "nil-hashmap",
			value:      nil,
		},
		{
			definition: NewSliceArrayArgument("slice-empty", []string{}),
			class:      flux.JavaUtilListClassName,
			argType:    flux.ArgumentTypeComplex,
			name:       "slice-empty",
			value:      []string{},
		},
		{
			definition: NewSliceArrayArgument("slice", []string{"a", "b"}),
			class:      flux.JavaUtilListClassName,
			argType:    flux.ArgumentTypeComplex,
			name:       "slice",
			value:      []string{"a", "b"},
		},
		{
			definition: NewSliceArrayArgument("slice-nil", nil),
			class:      flux.JavaUtilListClassName,
			argType:    flux.ArgumentTypeComplex,
			name:       "slice-nil",
			value:      nil,
		},
		{
			definition: NewSliceArrayArgument("array", [2]string{"a", "b"}),
			class:      flux.JavaUtilListClassName,
			argType:    flux.ArgumentTypeComplex,
			name:       "array",
			value:      [2]string{"a", "b"},
		},
	}
	for _, tcase := range cases {
		assert := assert2.New(t)
		assert.Equal(tcase.class, tcase.definition.TypeClass, "type class")
		assert.Equal(tcase.argType, tcase.definition.Type, "arg type")
		assert.Equal(tcase.name, tcase.definition.Name, "name")
		assert.Equal(tcase.value, tcase.definition.HttpValue.Value(), "name")
	}
}
