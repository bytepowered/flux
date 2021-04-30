package common

import (
	"github.com/bytepowered/flux/fluxkit"
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
		{lookup: "scope:key", scope: "SCOPE", key: "key", ok: true},
		{lookup: "SCOPE:key", scope: "SCOPE", key: "key", ok: true},
		{lookup: "scope:key:", scope: "SCOPE", key: "key", ok: true},
		{lookup: "scope:key:key2", scope: "SCOPE", key: "key", ok: true},
		{lookup: "Scope:key", scope: "SCOPE", key: "key", ok: true},
	}
	assert := assert.New(t)
	for _, tcase := range cases {
		scope, key, ok := fluxkit.ParseScopeExpr(tcase.lookup)
		assert.Equal(tcase.ok, ok, "ok: not match")
		assert.Equal(tcase.scope, scope, "scope: not match")
		assert.Equal(tcase.key, key, "key: not match")
	}
}
