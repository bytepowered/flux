package support

import (
	"github.com/bytepowered/flux"
	assert2 "github.com/stretchr/testify/assert"
	"testing"
)

func TestDefaultArgumentValueLookupFunc(t *testing.T) {
	values := map[string]interface{}{
		"path":   "hahaha",
		"query":  "query-val",
		"form":   123,
		"param":  true,
		"header": "UA",
		"attr":   "attr",
		"auto":   "auto",
	}
	valctx := NewValuesContext(values)
	cases := []struct {
		scope  string
		key    string
		expect flux.MTValue
	}{
		{scope: flux.ScopePath, key: "path", expect: flux.WrapTextMTValue("hahaha")},
		{scope: flux.ScopeQuery, key: "query", expect: flux.WrapTextMTValue("query-val")},
		{scope: flux.ScopeForm, key: "form", expect: flux.WrapTextMTValue("123")},
		{scope: flux.ScopeForm, key: "param", expect: flux.WrapTextMTValue("true")},
		{scope: flux.ScopeHeader, key: "header", expect: flux.WrapTextMTValue("UA")},
		{scope: flux.ScopeAuto, key: "auto", expect: flux.WrapObjectMTValue("auto")},
		{scope: flux.ScopeAttr, key: "attr", expect: flux.WrapObjectMTValue("attr")},
	}
	assert := assert2.New(t)
	for _, c := range cases {
		mtv, err := DefaultArgumentValueLookupFunc(c.scope, c.key, valctx)
		assert.NoError(err, "must no error")
		assert.Equal(c.expect, mtv)
	}
}
