package common

import (
	"github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-node/internal"
	"github.com/labstack/echo/v4"
	"net/http/httptest"
)

var mock = echo.New()

func MockWebContext(id string) flux.ServerWebContext {
	mr := httptest.NewRequest("GET", "http://mocking/"+id, nil)
	mw := httptest.NewRecorder()
	return internal.NewServeWebContext(id, mock.NewContext(mr, mw))
}

func MockContext(id string) *flux.Context {
	return flux.NewContext(MockWebContext(id), &flux.Endpoint{})
}
