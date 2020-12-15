package flux

import (
	"errors"
	assert2 "github.com/stretchr/testify/assert"
	"testing"
)

func TestStateError_ErrorNil(t *testing.T) {
	err := &ServeError{
		StatusCode: 500,
		ErrorCode:  "SERVER_ERROR",
		Message:    "Server internal error",
	}
	emsg := err.Error()
	assert := assert2.New(t)
	assert.Equal("ServeError: StatusCode=500, ErrorCode=SERVER_ERROR, Message=Server internal error", emsg)
}

func TestStateError_ErrorPresent(t *testing.T) {
	err := &ServeError{
		StatusCode: 500,
		ErrorCode:  "SERVER_ERROR",
		Message:    "Server internal error",
		Internal:   errors.New("error"),
	}
	emsg := err.Error()
	assert := assert2.New(t)
	assert.Equal("ServeError: StatusCode=500, ErrorCode=SERVER_ERROR, Message=Server internal error, Error=error", emsg)
}
