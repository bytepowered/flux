package backend

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/support"
	assert2 "github.com/stretchr/testify/assert"
	"testing"
)

func TestLookupResolveWith(t *testing.T) {
	context := support.NewValuesContext(map[string]interface{}{
		"accessId": "aid123",
		"userId":   123,
		"enabled":  true,
		"profile":  map[string]interface{}{"key": "value"},
	})
	cases := []struct {
		argument flux.Argument
		expect   interface{}
	}{
		{argument: ext.NewStringArgument("accessId"), expect: "aid123"},
		{argument: ext.NewIntegerArgument("userId"), expect: 123},
		{argument: ext.NewBooleanArgument("enabled"), expect: true},
		{argument: ext.NewStringMapArgument("profile"), expect: map[string]interface{}{"key": "value"}},
	}
	assert := assert2.New(t)
	for _, c := range cases {
		value, err := LookupResolveWith(
			c.argument,
			support.DefaultArgumentValueLookupFunc, support.DefaultArgumentValueResolveFunc,
			context)
		assert.NoError(err, "must no error")
		assert.Equal(c.expect, value)
	}
}
