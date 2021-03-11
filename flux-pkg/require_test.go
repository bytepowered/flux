package fluxpkg

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMustNotNil(t *testing.T) {
	asserter := assert.New(t)
	asserter.Panics(func() {
		MustNotNil(nil, "should nil, panic")
	})
}

func TestMustNotNilFunc(t *testing.T) {
	asserter := assert.New(t)
	var nilFunc func() = nil
	asserter.Panics(func() {
		MustNotNil(nilFunc, "should nil, panic")
	})
}
