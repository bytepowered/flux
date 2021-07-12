package common

import (
	assert2 "github.com/stretchr/testify/assert"
	"testing"
)

func TestToEncodeValue(t *testing.T) {
	cases := []struct {
		expect interface{}
		valid  bool
		value  interface{}
	}{
		{
			expect: float64(0),
			valid:  true,
			value:  0,
		},
		{
			expect: float64(0.0),
			valid:  true,
			value:  float32(0.0),
		},
		{
			expect: float64(0.99),
			valid:  true,
			value:  float64(0.99),
		},
		{
			expect: "abc",
			valid:  true,
			value:  "abc",
		},
	}
	assert := assert2.New(t)
	for _, c := range cases {
		ev := ToEncodeValue(c.value)
		assert.Equal(c.valid, ev.IsValid())
		if c.valid {
			assert.Equal(c.expect, ev.Value)
		}
	}
}
