package common

import (
	"github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-node/internal"
	"github.com/labstack/echo/v4"
	"net/http/httptest"
)

var mock = echo.New()

func MockWebContext(id string) flux.ServerWebContext {
	mr := httptest.NewRequest("GET", "http://mockctx/"+id, nil)
	mw := httptest.NewRecorder()
	return internal.NewServeWebContext(mock.NewContext(mr, mw), id, nil)
}

func MockContext(id string) *flux.Context {
	ctx := flux.NewContext()
	ctx.Reset(MockWebContext(id), &flux.Endpoint{})
	return ctx
}
