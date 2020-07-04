package internal

import (
	"github.com/bytepowered/flux"
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
)

// RequestWrapReader Request请求读取接口的实现
type RequestWrapReader struct {
	context echo.Context
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

// ResponseWrapWriter 定义响应数据
type ResponseWrapWriter struct {
	status  int
	headers http.Header
	body    interface{}
}

func (r *ResponseWrapWriter) StatusCode() int {
	return r.status
}

func (r *ResponseWrapWriter) Headers() http.Header {
	return r.headers
}

func (r *ResponseWrapWriter) Body() interface{} {
	return r.body
}

func (r *ResponseWrapWriter) SetStatusCode(status int) flux.ResponseWriter {
	r.status = status
	return r
}

func (r *ResponseWrapWriter) AddHeader(name, value string) flux.ResponseWriter {
	r.headers.Add(name, value)
	return r
}

func (r *ResponseWrapWriter) SetHeaders(headers http.Header) flux.ResponseWriter {
	r.headers = headers
	return r
}

func (r *ResponseWrapWriter) SetBody(body interface{}) flux.ResponseWriter {
	r.body = body
	return r
}

func (r *ResponseWrapWriter) reset() {
	r.status = flux.StatusOK
	r.body = nil
	r.headers = http.Header{}
}

func newResponseWriter() *ResponseWrapWriter {
	return &ResponseWrapWriter{
		status:  flux.StatusOK,
		headers: http.Header{},
		body:    nil,
	}
}
