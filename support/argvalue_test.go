package support

import (
	"github.com/bytepowered/flux"
	"github.com/spf13/cast"
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
		"attrs":  map[string]string{"key": "value"},
		"auto":   "auto",
	}
	valctx := NewValuesContext(values)
	assert := assert2.New(t)
	cases := []struct {
		scope  string
		key    string
		expect interface{}
	}{
		{scope: flux.ScopePath, key: "path", expect: "hahaha"},
		{scope: flux.ScopeQuery, key: "query", expect: "query-val"},
		{scope: flux.ScopeForm, key: "form", expect: 123},
		{scope: flux.ScopeForm, key: "param", expect: true},
		{scope: flux.ScopeHeader, key: "header", expect: "UA"},
		{scope: flux.ScopeAttr, key: "attr", expect: "attr"},
		{scope: flux.ScopeAuto, key: "auto", expect: "auto"},
	}
	for _, c := range cases {
		mtv, err := DefaultArgumentValueLookupFunc(c.scope, c.key, valctx)
		assert.NoError(err, "must no error")
		assert.Equal(cast.ToString(c.expect), mtv.Value)
	}
}
