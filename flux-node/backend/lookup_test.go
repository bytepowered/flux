package backend

import (
	"github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-node/context"
	assert2 "github.com/stretchr/testify/assert"
	"net/http"
	"net/url"
	"testing"
)

func TestNilContext(t *testing.T) {
	assert := assert2.New(t)
	// Scope & key
	_, err0 := DefaultArgumentLookupFunc("", "", context.NewEmpty())
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
	valctx := context.NewMockWith("@rid", values)
	cases := []struct {
		scope  string
		key    string
		expect flux.MTValue
	}{
		{scope: flux.ScopePath, key: "path", expect: flux.WrapStringMTValue("hahaha")},
		{scope: flux.ScopeQuery, key: "query", expect: flux.WrapStringMTValue("query-val")},
		{scope: flux.ScopeForm, key: "form", expect: flux.WrapStringMTValue("123")},
		{scope: flux.ScopeForm, key: "param", expect: flux.WrapStringMTValue("true")},
		{scope: flux.ScopeHeader, key: "header", expect: flux.WrapStringMTValue("UA")},
		{scope: flux.ScopeAuto, key: "auto", expect: flux.WrapObjectMTValue("auto")},
		{scope: flux.ScopeAttr, key: "attr", expect: flux.WrapObjectMTValue("attr")},
		{scope: flux.ScopePathMap, key: "path-values", expect: flux.WrapStrValuesMapMTValue(url.Values{})},
		{scope: flux.ScopeQueryMap, key: "query-values", expect: flux.WrapStrValuesMapMTValue(url.Values{})},
		{scope: flux.ScopeFormMap, key: "form-query", expect: flux.WrapStrValuesMapMTValue(url.Values{})},
		{scope: flux.ScopeHeaderMap, key: "header-map", expect: flux.WrapStrValuesMapMTValue(url.Values{})},
	}
	assert := assert2.New(t)
	for _, c := range cases {
		mtv, err := DefaultArgumentLookupFunc(c.scope, c.key, valctx)
		assert.NoError(err, "must no error")
		assert.Equal(c.expect, mtv)
	}
}
