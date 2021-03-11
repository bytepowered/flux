package backend

import (
	"github.com/bytepowered/flux"
	flux2 "github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-node/context"
	assert2 "github.com/stretchr/testify/assert"
	"net/http"
	"net/url"
	"testing"
)

func TestNilContext(t *testing.T) {
	assert := assert2.New(t)
	// Scope & key
	_, err0 := DefaultArgumentLookupFunc("", "", context.NewEmptyContext())
	assert.Error(err0, "must error")
	// Nil context
	_, err1 := DefaultArgumentLookupFunc("a", "b", nil)
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
	valctx := context.NewMockContext(values)
	cases := []struct {
		scope  string
		key    string
		expect flux2.MTValue
	}{
		{scope: flux2.ScopePath, key: "path", expect: flux2.WrapStringMTValue("hahaha")},
		{scope: flux2.ScopeQuery, key: "query", expect: flux2.WrapStringMTValue("query-val")},
		{scope: flux2.ScopeForm, key: "form", expect: flux2.WrapStringMTValue("123")},
		{scope: flux2.ScopeForm, key: "param", expect: flux2.WrapStringMTValue("true")},
		{scope: flux2.ScopeHeader, key: "header", expect: flux2.WrapStringMTValue("UA")},
		{scope: flux2.ScopeAuto, key: "auto", expect: flux2.WrapObjectMTValue("auto")},
		{scope: flux2.ScopeAttr, key: "attr", expect: flux2.WrapObjectMTValue("attr")},
		{scope: flux.ScopeValue, key: "value", expect: flux2.WrapObjectMTValue("value")},
		{scope: flux2.ScopePathMap, key: "path-values", expect: flux2.WrapStrValuesMapMTValue(url.Values{})},
		{scope: flux2.ScopeQueryMap, key: "query-values", expect: flux2.WrapStrValuesMapMTValue(url.Values{})},
		{scope: flux2.ScopeFormMap, key: "form-query", expect: flux2.WrapStrValuesMapMTValue(url.Values{})},
		{scope: flux2.ScopeHeaderMap, key: "header-map", expect: flux2.WrapStrValuesMapMTValue(url.Values{})},
	}
	assert := assert2.New(t)
	for _, c := range cases {
		mtv, err := DefaultArgumentLookupFunc(c.scope, c.key, valctx)
		assert.NoError(err, "must no error")
		assert.Equal(c.expect, mtv)
	}
}
