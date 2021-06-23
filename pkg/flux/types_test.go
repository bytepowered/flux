package flux

import (
	assert2 "github.com/stretchr/testify/assert"
	"testing"
)

func TestNamedValueSpec_ToString(t *testing.T) {
	cases := []TestCase{
		{
			Expected: "a",
			Actual:   NamedValueSpec{Value: []string{"a", "b"}}.GetString(),
			Message:  "should be [0] == a",
		},
		{
			Expected: "",
			Actual:   NamedValueSpec{}.GetString(),
			Message:  "default nil should be empty string",
		},
		{
			Expected: "",
			Actual:   NamedValueSpec{Value: []string{}}.GetString(),
			Message:  "empty array should be empty string",
		},
		{
			Expected: "",
			Actual:   NamedValueSpec{Value: nil}.GetString(),
			Message:  "nil value should be empty string",
		},
		{
			Expected: "0",
			Actual:   NamedValueSpec{Value: 0}.GetString(),
			Message:  "0 value should be '0' string",
		},
	}
	tester := assert2.New(t)
	for _, c := range cases {
		tester.Equal(c.Expected, c.Actual, c.Message)
	}
}
