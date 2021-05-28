package common

import (
	"context"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/listener"
	"github.com/labstack/echo/v4"
	"net/http/httptest"
)

var _slim = echo.New()

func NewSlimContextTODO(id string) *flux.Context {
	return NewSlimContext(context.TODO(), id)
}

func NewSlimContext(ctx context.Context, id string, vars ...map[string]interface{}) *flux.Context {
	fxctx := flux.NewContext()
	fxctx.Reset(newSlimWithID(ctx, id), &flux.Endpoint{Application: "slim"})
	fxctx.SetVariable("is.slim.ctx", true)
	if len(vars) > 0 {
		for k, v := range vars[0] {
			fxctx.SetVariable(k, v)
		}
	}
	return fxctx
}

func newSlimWithID(ctx context.Context, id string) flux.ServerWebContext {
	req := httptest.NewRequest("GET", "http://slimctx/"+id, nil)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	return listener.NewServeWebContext(_slim.NewContext(req, rec), id, nil)
}
