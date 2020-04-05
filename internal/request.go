package internal

import (
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
)

type FxRequest struct {
	ctx echo.Context
}

func (r *FxRequest) ParamInQuery(name string) string {
	return r.ctx.QueryParam(name)
}

func (r *FxRequest) ParamInPath(name string) string {
	return r.ctx.Param(name)
}

func (r *FxRequest) ParamInForm(name string) string {
	return r.ctx.FormValue(name)
}

func (r *FxRequest) Header(name string) string {
	return r.ctx.Request().Header.Get(name)
}

func (r *FxRequest) Headers() http.Header {
	return r.ctx.Request().Header.Clone()
}

func (r *FxRequest) Cookie(name string) string {
	c, e := r.ctx.Cookie(name)
	if e == echo.ErrCookieNotFound {
		return ""
	} else {
		return c.Raw
	}
}

func (r *FxRequest) RemoteAddress() string {
	return r.ctx.RealIP()
}

func (r *FxRequest) HttpRequest() *http.Request {
	return r.ctx.Request()
}

func (r *FxRequest) HttpBody() (io.ReadCloser, error) {
	return r.ctx.Request().GetBody()
}

func (r *FxRequest) attach(echo echo.Context) {
	r.ctx = echo
}

func newRequestReader() *FxRequest {
	return &FxRequest{}
}
