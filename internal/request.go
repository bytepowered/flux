package internal

import (
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
)

type RequestWrapReader struct {
	context echo.Context
}

func (r *RequestWrapReader) ParamInQuery(name string) string {
	return r.QueryValue(name)
}

func (r *RequestWrapReader) ParamInPath(name string) string {
	return r.PathValue(name)
}

func (r *RequestWrapReader) ParamInForm(name string) string {
	return r.FormValue(name)
}

func (r *RequestWrapReader) Header(name string) string {
	return r.HeaderValue(name)
}

func (r *RequestWrapReader) QueryValue(name string) string {
	return r.context.QueryParam(name)
}

func (r *RequestWrapReader) PathValue(name string) string {
	return r.context.Param(name)
}

func (r *RequestWrapReader) FormValue(name string) string {
	return r.context.FormValue(name)
}

func (r *RequestWrapReader) HeaderValue(name string) string {
	return r.context.Request().Header.Get(name)
}

func (r *RequestWrapReader) CookieValue(name string) string {
	c, e := r.context.Cookie(name)
	if e == echo.ErrCookieNotFound {
		return ""
	} else if nil != c {
		return c.Raw
	} else {
		return ""
	}
}

func (r *RequestWrapReader) Headers() http.Header {
	return r.context.Request().Header.Clone()
}

func (r *RequestWrapReader) Cookie(name string) string {
	return r.CookieValue(name)
}

func (r *RequestWrapReader) RemoteAddress() string {
	return r.context.RealIP()
}

func (r *RequestWrapReader) HttpRequest() *http.Request {
	return r.context.Request()
}

func (r *RequestWrapReader) HttpBody() (io.ReadCloser, error) {
	return r.context.Request().GetBody()
}

func (r *RequestWrapReader) reattach(echo echo.Context) {
	r.context = echo
}

func (r *RequestWrapReader) reset() {
	r.context = nil
}

func newRequestReader() *RequestWrapReader {
	return &RequestWrapReader{}
}
