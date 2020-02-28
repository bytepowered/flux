package internal

import (
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
)

type request struct {
	echoCtx echo.Context
}

func (r *request) ParamInQuery(name string) string {
	return r.echoCtx.QueryParam(name)
}

func (r *request) ParamInPath(name string) string {
	return r.echoCtx.Param(name)
}

func (r *request) ParamInForm(name string) string {
	return r.echoCtx.FormValue(name)
}

func (r *request) Header(name string) string {
	return r.echoCtx.Request().Header.Get(name)
}

func (r *request) Headers() http.Header {
	return r.echoCtx.Request().Header.Clone()
}

func (r *request) Cookie(name string) string {
	c, e := r.echoCtx.Cookie(name)
	if e == echo.ErrCookieNotFound {
		return ""
	} else {
		return c.Raw
	}
}

func (r *request) RemoteAddress() string {
	return r.echoCtx.RealIP()
}

func (r *request) HttpRequest() *http.Request {
	return r.echoCtx.Request()
}

func (r *request) HttpBody() (io.ReadCloser, error) {
	return r.echoCtx.Request().GetBody()
}

func (r *request) attach(echo echo.Context) {
	r.echoCtx = echo
}

func newRequestReader() *request {
	return &request{}
}
