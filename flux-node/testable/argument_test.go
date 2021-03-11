package testable

import (
	flux2 "github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-node/backend"
	"github.com/bytepowered/flux/flux-node/context"
	"github.com/bytepowered/flux/flux-node/ext"
	assert2 "github.com/stretchr/testify/assert"
	"testing"
)

func TestPrimitiveArgumentLookupResolve(t *testing.T) {
	ext.SetArgumentLookupFunc(backend.DefaultArgumentLookupFunc)
	cases := []struct {
		definition flux2.Argument
		class      string
		argType    string
		name       string
		expected   interface{}
	}{
		{
			definition: ext.NewStringArgument("str"),
			class:      flux2.JavaLangStringClassName,
			argType:    flux2.ArgumentTypePrimitive,
			name:       "str",
			expected:   "value:str",
		},
		{
			definition: ext.NewIntegerArgument("int"),
			class:      flux2.JavaLangIntegerClassName,
			argType:    flux2.ArgumentTypePrimitive,
			name:       "int",
			expected:   int(12345),
		},
		{
			definition: ext.NewLongArgument("long"),
			class:      flux2.JavaLangLongClassName,
			argType:    flux2.ArgumentTypePrimitive,
			name:       "long",
			expected:   int64(1234567890),
		},
		{
			definition: ext.NewBooleanArgument("boolean"),
			class:      flux2.JavaLangBooleanClassName,
			argType:    flux2.ArgumentTypePrimitive,
			name:       "boolean",
			expected:   true,
		},
		{
			definition: ext.NewFloatArgument("float32"),
			class:      flux2.JavaLangFloatClassName,
			argType:    flux2.ArgumentTypePrimitive,
			name:       "float32",
			expected:   float32(12345.678),
		},
		{
			definition: ext.NewFloatArgument("float"),
			class:      flux2.JavaLangFloatClassName,
			argType:    flux2.ArgumentTypePrimitive,
			name:       "float",
			expected:   float32(87654.321),
		},
		{
			definition: ext.NewDoubleArgument("float64"),
			class:      flux2.JavaLangDoubleClassName,
			argType:    flux2.ArgumentTypePrimitive,
			name:       "float64",
			expected:   12345.678,
		},
		{
			definition: ext.NewDoubleArgument("double"),
			class:      flux2.JavaLangDoubleClassName,
			argType:    flux2.ArgumentTypePrimitive,
			name:       "double",
			expected:   98765.4321,
		},
		{
			definition: ext.NewSliceArrayArgument("list", flux2.JavaLangStringClassName),
			class:      flux2.JavaUtilListClassName,
			argType:    flux2.ArgumentTypePrimitive,
			name:       "list",
			expected:   []interface{}{"abc", "def"},
		},
		{
			definition: ext.NewSliceArrayArgument("tostrlist", flux2.JavaLangStringClassName),
			class:      flux2.JavaUtilListClassName,
			argType:    flux2.ArgumentTypePrimitive,
			name:       "tostrlist",
			expected:   []interface{}{"123", "456"},
		},
		{
			definition: ext.NewSliceArrayArgument("intlist", flux2.JavaLangIntegerClassName),
			class:      flux2.JavaUtilListClassName,
			argType:    flux2.ArgumentTypePrimitive,
			name:       "intlist",
			expected:   []interface{}{12345, 56789},
		},
	}
	assert := assert2.New(t)
	ctx := context.NewMockContext(map[string]interface{}{
		"str":       "value:str",
		"int":       12345,
		"long":      int64(1234567890),
		"boolean":   true,
		"float32":   float32(12345.678),
		"float":     float32(87654.321),
		"float64":   12345.678,
		"double":    98765.4321,
		"list":      []string{"abc", "def"},
		"tostrlist": []int{123, 456},
		"intlist":   []string{"12345", "56789"},
	})
	for _, tcase := range cases {
		assert.Equal(tcase.class, tcase.definition.Class, "type class")
		assert.Equal(tcase.argType, tcase.definition.Type, "arg type")
		assert.Equal(tcase.name, tcase.definition.Name, "name")
		// check resolve
		v, err := tcase.definition.Resolve(ctx)
		assert.Nil(err)
		assert.Equal(tcase.expected, v, "value match")
	}
}

func TestComplexArgumentLookupResolve(t *testing.T) {
	ext.SetArgumentLookupFunc(backend.DefaultArgumentLookupFunc)
	cases := []struct {
		definition flux2.Argument
		class      string
		argType    string
		name       string
		expected   interface{}
	}{
		{
			definition: ext.NewStringMapArgument("stringmap"),
			class:      flux2.JavaUtilMapClassName,
			argType:    flux2.ArgumentTypeComplex,
			name:       "stringmap",
			expected: map[string]interface{}{
				"key": "value",
				"int": 123,
			},
		},
		{
			definition: ext.NewHashMapArgument("hashmap"),
			class:      flux2.JavaUtilMapClassName,
			argType:    flux2.ArgumentTypeComplex,
			name:       "hashmap",
			expected: map[string]interface{}{
				"key": "value",
				"int": 123,
			},
		},
		{
			definition: ext.NewComplexArgument("net.bytepowreed.test.UserVO", "user"),
			class:      "net.bytepowreed.test.UserVO",
			argType:    flux2.ArgumentTypeComplex,
			name:       "user",
			expected: map[string]interface{}{
				"class":    "net.bytepowreed.test.UserVO",
				"username": "yongjiachen",
				"year":     2020,
			},
		},
		{
			definition: func() flux2.Argument {
				arg := ext.NewComplexArgument("net.bytepowreed.test.POJO", "pojo")
				arg.Fields = []flux2.Argument{
					ext.NewStringArgument("username"),
					ext.NewIntegerArgument("year"),
					ext.NewHashMapArgument("hashmap"),
				}
				return arg
			}(),
			class:   "net.bytepowreed.test.POJO",
			argType: flux2.ArgumentTypeComplex,
			name:    "pojo",
			expected: map[string]interface{}{
				"class":    "net.bytepowreed.test.POJO",
				"username": "yongjiachen",
				"year":     2020,
				"hashmap": map[string]interface{}{
					"key": "value",
					"int": 123,
				},
			},
		},
	}
	assert := assert2.New(t)
	ctx := context.NewMockContext(map[string]interface{}{
		"stringmap": map[string]interface{}{
			"key": "value",
			"int": 123,
		},
		"hashmap": map[string]interface{}{
			"key": "value",
			"int": 123,
		},
		"user": map[string]interface{}{
			"username": "yongjiachen",
			"year":     2020,
		},
		"username": "yongjiachen",
		"year":     2020,
	})
	for _, tcase := range cases {
		assert.Equal(tcase.class, tcase.definition.Class, "type class")
		assert.Equal(tcase.argType, tcase.definition.Type, "arg type")
		assert.Equal(tcase.name, tcase.definition.Name, "name")
		// check resolve
		v, err := tcase.definition.Resolve(ctx)
		assert.Nil(err)
		assert.Equal(tcase.expected, v, "value match")
	}
}

func TestComplexArgumentValueLoader(t *testing.T) {
	cases := []struct {
		definition flux2.Argument
		name       string
		expected   interface{}
	}{
		{
			definition: ext.NewStringArgumentWith("xx", "my-value"),
			name:       "string",
			expected:   "my-value",
		},
		{
			definition: ext.NewIntegerArgumentWith("xx", 1234),
			name:       "int",
			expected:   1234,
		},
	}
	assert := assert2.New(t)
	ctx := context.NewEmptyContext()
	for _, tcase := range cases {
		// check resolve
		v, err := tcase.definition.Resolve(ctx)
		assert.Nil(err)
		assert.Equal(tcase.expected, v, "value match")
	}
}
