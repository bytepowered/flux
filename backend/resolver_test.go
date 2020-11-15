package backend

import (
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/support"
	assert2 "github.com/stretchr/testify/assert"
	"testing"
)

func TestLookupResolveWith(t *testing.T) {
	values := support.NewValuesContext(map[string]interface{}{
		"accessId": "aid123",
	})
	value, err := LookupResolveWith(
		ext.NewStringArgument("accessId"),
		support.DefaultArgumentValueLookupFunc, support.DefaultArgumentValueResolveFunc,
		values)
	assert := assert2.New(t)
	assert.NoError(err, "must no error")
	assert.Equal("aid123", value)
}
