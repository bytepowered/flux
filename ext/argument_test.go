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
	f1 := func(scope, key string, context flux.Context) (value flux.MIMEValue, err error) {
		return flux.NewTextTypedValue("value"), nil
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
			definition: NewStringArgument("str"),
			class:      flux.JavaLangStringClassName,
			argType:    flux.ArgumentTypePrimitive,
			name:       "str",
			value:      "001",
		},
		{
			definition: NewIntegerArgument("int"),
			class:      flux.JavaLangIntegerClassName,
			argType:    flux.ArgumentTypePrimitive,
			name:       "int",
			value:      1,
		},
		{
			definition: NewLongArgument("long"),
			class:      flux.JavaLangLongClassName,
			argType:    flux.ArgumentTypePrimitive,
			name:       "long",
			value:      int64(2),
		},
		{
			definition: NewBooleanArgument("boolean"),
			class:      flux.JavaLangBooleanClassName,
			argType:    flux.ArgumentTypePrimitive,
			name:       "boolean",
			value:      true,
		},
		{
			definition: NewFloatArgument("float32"),
			class:      flux.JavaLangFloatClassName,
			argType:    flux.ArgumentTypePrimitive,
			name:       "float32",
			value:      float32(12.3456),
		},
		{
			definition: NewDoubleArgument("float64"),
			class:      flux.JavaLangDoubleClassName,
			argType:    flux.ArgumentTypePrimitive,
			name:       "float64",
			value:      float64(12.345678),
		},
		{
			definition: NewDoubleArgument("float64"),
			class:      flux.JavaLangDoubleClassName,
			argType:    flux.ArgumentTypePrimitive,
			name:       "float64",
			value:      float64(12.345678),
		},
		{
			definition: NewStringMapArgument("stringmap"),
			class:      flux.JavaUtilMapClassName,
			argType:    flux.ArgumentTypeComplex,
			name:       "stringmap",
			value:      map[string]interface{}{"key": "value"},
		},
		{
			definition: NewHashMapArgument("hashmap"),
			class:      flux.JavaUtilMapClassName,
			argType:    flux.ArgumentTypeComplex,
			name:       "hashmap",
			value:      map[string]interface{}{"hash": "map"},
		},
		{
			definition: NewHashMapArgument("nil-hashmap"),
			class:      flux.JavaUtilMapClassName,
			argType:    flux.ArgumentTypeComplex,
			name:       "nil-hashmap",
			value:      nil,
		},
		{
			definition: NewSliceArrayArgument("slice-empty"),
			class:      flux.JavaUtilListClassName,
			argType:    flux.ArgumentTypeComplex,
			name:       "slice-empty",
			value:      []string{},
		},
		{
			definition: NewSliceArrayArgument("slice"),
			class:      flux.JavaUtilListClassName,
			argType:    flux.ArgumentTypeComplex,
			name:       "slice",
			value:      []string{"a", "b"},
		},
		{
			definition: NewSliceArrayArgument("slice-nil"),
			class:      flux.JavaUtilListClassName,
			argType:    flux.ArgumentTypeComplex,
			name:       "slice-nil",
			value:      nil,
		},
		{
			definition: NewSliceArrayArgument("array"),
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
	}
}
