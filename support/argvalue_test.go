package support

import (
	"github.com/bytepowered/flux"
	assert2 "github.com/stretchr/testify/assert"
	"net/http"
	"net/url"
	"testing"
)

func TestNilContext(t *testing.T) {
	assert := assert2.New(t)
	// Scope & key
	_, err0 := DefaultArgumentValueLookupFunc("", "", NewEmptyContext())
	assert.Error(err0, "must error")
	// Nil context
	_, err1 := DefaultArgumentValueLookupFunc("a", "b", nil)
	assert.Error(err1, "must error")
}

func TestDefaultArgumentValueLookupFunc(t *testing.T) {
	values := map[string]interface{}{
		"path":          "hahaha",
		"path-values":   url.Values{},
		"query":         "query-val",
		"query-values":  url.Values{},
		"form":          123,
		"form-values":   url.Values{},
		"param":         true,
		"header":        "UA",
		"header-values": http.Header{},
		"attr":          "attr",
		"auto":          "auto",
		"value":         "value",
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
		{scope: flux.ScopeValue, key: "value", expect: flux.WrapObjectMTValue("value")},
		{scope: flux.ScopePathMap, key: "path-values", expect: flux.WrapStrValuesMapMTValue(url.Values{})},
		{scope: flux.ScopeQueryMap, key: "query-values", expect: flux.WrapStrValuesMapMTValue(url.Values{})},
		{scope: flux.ScopeFormMap, key: "form-query", expect: flux.WrapStrValuesMapMTValue(url.Values{})},
		{scope: flux.ScopeHeaderMap, key: "header-map", expect: flux.WrapStrValuesMapMTValue(url.Values{})},
	}
	assert := assert2.New(t)
	for _, c := range cases {
		mtv, err := DefaultArgumentValueLookupFunc(c.scope, c.key, valctx)
		assert.NoError(err, "must no error")
		assert.Equal(c.expect, mtv)
	}
}
