package pkg

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRequireNotNil(t *testing.T) {
	asserter := assert.New(t)
	asserter.Panics(func() {
		RequireNotNil(nil, "should nil, panic")
	})
}

func TestRequireNotNilFunc(t *testing.T) {
	asserter := assert.New(t)
	var nilFunc func() = nil
	asserter.Panics(func() {
		RequireNotNil(nilFunc, "should nil, panic")
	})
}
