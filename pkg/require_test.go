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
