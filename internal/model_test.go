package internal

import (
	"github.com/bytepowered/flux"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVerifyAnnotationsPasses(t *testing.T) {
	allpass := flux.Annotations{
		"abc":         0,
		"abc.def":     0,
		"flux.go/abc": 0,
		"flux.go/123": 0,
	}
	tester := assert.New(t)
	tester.Nil(VerifyAnnotations(allpass))
}

func TestVerifyAnnotationsFails(t *testing.T) {
	fails := flux.Annotations{
		"-abc":          0,
		"+abc.def":      0,
		"//flux.go/abc": 0,
		"**flux.go/123": 0,
	}
	tester := assert.New(t)
	for key := range fails {
		tester.NotNil(VerifyAnnotationKeySpec(key))
	}
}
