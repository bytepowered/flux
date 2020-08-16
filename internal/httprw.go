package internal

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/webex"
	"io"
	"net/http"
)

var _ flux.RequestReader = new(RequestWrapReader)

// RequestWrapReader Request请求读取接口的实现
type RequestWrapReader struct {
	webc webex.WebContext
}

func (r *RequestWrapReader) QueryValue(name string) string {
	return r.webc.QueryValue(name)
}

func (r *RequestWrapReader) PathValue(name string) string {
	return r.webc.PathValue(name)
}

func (r *RequestWrapReader) FormValue(name string) string {
	return r.webc.FormValue(name)
}

func (r *RequestWrapReader) HeaderValue(name string) string {
	return r.webc.RequestHeader().Get(name)
}

func (r *RequestWrapReader) CookieValue(name string) string {
	return r.webc.CookieValue(name)
}

func (r *RequestWrapReader) Headers() http.Header {
	return r.webc.RequestHeader().Clone()
}

func (r *RequestWrapReader) RemoteAddress() string {
	return r.webc.RequestRemoteAddr()
}

func (r *RequestWrapReader) HttpRequest() *http.Request {
	return r.webc.Request()
}

func (r *RequestWrapReader) HttpBody() (io.ReadCloser, error) {
	return r.webc.Request().GetBody()
}

func (r *RequestWrapReader) reattach(webex webex.WebContext) {
	r.webc = webex
}

func (r *RequestWrapReader) reset() {
	r.webc = nil
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
