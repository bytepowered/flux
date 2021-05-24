package common

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/listener"
	"github.com/labstack/echo/v4"
	"net/http/httptest"
)

var mock = echo.New()

func MockWebContext(id string) flux.ServerWebContext {
	mr := httptest.NewRequest("GET", "http://mockctx/"+id, nil)
	mw := httptest.NewRecorder()
	return listener.NewServeWebContext(mock.NewContext(mr, mw), id, nil)
}

func MockContext(id string) *flux.Context {
	return MockContextVars(id, map[string]interface{}{
		"is.mock.ctx": true,
	})
}

func MockContextVars(id string, vars map[string]interface{}) *flux.Context {
	ctx := flux.NewContext()
	ctx.Reset(MockWebContext(id), &flux.Endpoint{})
	for k, v := range vars {
		ctx.SetVariable(k, v)
	}
	return ctx
}
