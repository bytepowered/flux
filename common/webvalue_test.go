package common

import (
	"fmt"
	"github.com/bytepowered/flux/toolkit"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestScopeLookupParsePair(t *testing.T) {
	cases := []struct {
		lookup string
		scope  string
		key    string
		ok     bool
	}{
		{lookup: "", scope: "", key: "", ok: false},
		{lookup: ":", scope: "", key: "", ok: false},
		{lookup: "scope:", scope: "", key: "", ok: false},
		{lookup: ":key", scope: "", key: "", ok: false},
		{lookup: "scope:key", scope: "scope", key: "key", ok: true},
		{lookup: "SCOPE:key", scope: "SCOPE", key: "key", ok: true},
		{lookup: "Scope:key", scope: "Scope", key: "key", ok: true},
		{lookup: "scope:key:", scope: "", key: "", ok: false},
		{lookup: "scope:key:key2", scope: "", key: "", ok: false},
	}
	assert := assert.New(t)
	for idx, tcase := range cases {
		scope, key, ok := toolkit.ParseScopeExpr(tcase.lookup)
		assert.Equal(tcase.ok, ok, fmt.Sprintf("ok: not match, case.idx: %d", idx))
		assert.Equal(tcase.scope, scope, fmt.Sprintf("scope: not match, case.idx: %d", idx))
		assert.Equal(tcase.key, key, fmt.Sprintf("key: not match, case.idx: %d", idx))
	}
}
