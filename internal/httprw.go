package internal

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/webx"
	"net/http"
)

var _ flux.RequestReader = new(RequestWrappedReader)

// RequestWrappedReader Request请求读取接口的实现
type RequestWrappedReader struct {
	webx.WebContext
}

func (r *RequestWrappedReader) QueryValue(name string) string {
	return r.WebContext.QueryValue(name)
}

func (r *RequestWrappedReader) PathValue(name string) string {
	return r.WebContext.PathValue(name)
}

func (r *RequestWrappedReader) FormValue(name string) string {
	return r.WebContext.FormValue(name)
}

func (r *RequestWrappedReader) HeaderValue(name string) string {
	return r.WebContext.GetRequestHeader(name)
}

func (r *RequestWrappedReader) CookieValue(name string) string {
	for _, cookie := range r.WebContext.CookieValues() {
		if name == cookie.Name {
			return cookie.Value
		}
	}
	return ""
}

func (r *RequestWrappedReader) reattach(webex webx.WebContext) {
	r.WebContext = webex
}

func (r *RequestWrappedReader) reset() {
	r.WebContext = nil
}

func newRequestWrappedReader() *RequestWrappedReader {
	return &RequestWrappedReader{}
}

////

var _ flux.ResponseWriter = new(ResponseWrappedWriter)

// ResponseWrappedWriter 定义响应数据
type ResponseWrappedWriter struct {
	status  int
	headers http.Header
	body    interface{}
}

func (r *ResponseWrappedWriter) StatusCode() int {
	return r.status
}

func (r *ResponseWrappedWriter) Headers() http.Header {
	return r.headers
}

func (r *ResponseWrappedWriter) Body() interface{} {
	return r.body
}

func (r *ResponseWrappedWriter) SetStatusCode(status int) {
	r.status = status
}

func (r *ResponseWrappedWriter) AddHeader(name, value string) {
	r.headers.Add(name, value)
}

func (r *ResponseWrappedWriter) SetHeader(name, value string) {
	r.headers.Set(name, value)
}

func (r *ResponseWrappedWriter) SetHeaders(headers http.Header) {
	r.headers = headers
}

func (r *ResponseWrappedWriter) SetBody(body interface{}) {
	r.body = body
}

func (r *ResponseWrappedWriter) reset() {
	r.status = flux.StatusOK
	r.body = nil
	r.headers = http.Header{}
}

func newResponseWrappedWriter() *ResponseWrappedWriter {
	return &ResponseWrappedWriter{
		status:  flux.StatusOK,
		headers: http.Header{},
		body:    nil,
	}
}
