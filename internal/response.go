package internal

import (
	"github.com/bytepowered/flux"
	"net/http"
)

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
