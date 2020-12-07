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
		expect flux.MIMEValue
	}{
		{scope: flux.ScopePath, key: "path", expect: flux.WrapTextMIMEValue("hahaha")},
		{scope: flux.ScopeQuery, key: "query", expect: flux.WrapTextMIMEValue("query-val")},
		{scope: flux.ScopeForm, key: "form", expect: flux.WrapTextMIMEValue("123")},
		{scope: flux.ScopeForm, key: "param", expect: flux.WrapTextMIMEValue("true")},
		{scope: flux.ScopeHeader, key: "header", expect: flux.WrapTextMIMEValue("UA")},
		{scope: flux.ScopeAuto, key: "auto", expect: flux.WrapObjectMIMEValue("auto")},
		{scope: flux.ScopeAttr, key: "attr", expect: flux.WrapObjectMIMEValue("attr")},
	}
	assert := assert2.New(t)
	for _, c := range cases {
		mtv, err := DefaultArgumentValueLookupFunc(c.scope, c.key, valctx)
		assert.NoError(err, "must no error")
		assert.Equal(c.expect, mtv)
	}
}
