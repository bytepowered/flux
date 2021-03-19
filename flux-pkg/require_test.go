package fluxpkg

import (
	"github.com/stretchr/testify/assert"
	"net/http"
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

func TestIsNil(t *testing.T) {
	cases := []struct {
		isnil bool
		value interface{}
	}{
		{isnil: true},
		{isnil: true, value: nil},
		{isnil: true, value: error(nil)},
		{isnil: true, value: http.Handler(nil)},
		{isnil: true, value: map[string]interface{}(nil)},
		{isnil: true, value: chan int(nil)},
		{isnil: true, value: []int(nil)},
		{isnil: true, value: (*struct{})(nil)},
		{isnil: false, value: ""},
		{isnil: false, value: 0},
		{isnil: false, value: 0.0},
		{isnil: false, value: 0.0},
		{isnil: false, value: struct{}{}},
		{isnil: false, value: &struct{}{}},
	}
	for _, c := range cases {
		if c.isnil != IsNil(c.value) {
			t.Fatalf("should nil, but not, was: %+v", c.value)
		}
	}
}
