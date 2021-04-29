package fluxkit

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStringSliceContains(t *testing.T) {
	cases := []struct {
		elements []string
		ele      string
		has      bool
	}{
		{elements: []string{}, ele: "", has: false},
		{elements: []string{""}, ele: "", has: true},
		{elements: []string{"", "a"}, ele: "", has: true},
		{elements: []string{"", "a"}, ele: "a", has: true},
		{elements: []string{"", "a"}, ele: "b", has: false},
	}
	assert := assert.New(t)
	for _, tcase := range cases {
		has := StringSliceContains(tcase.elements, tcase.ele)
		assert.Equal(tcase.has, has, "has: not match")
	}
}
