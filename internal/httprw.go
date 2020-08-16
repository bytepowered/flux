package internal

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/webx"
	"net/http"
)

var _ flux.RequestReader = new(RequestWrappedReader)

// RequestWrappedReader Request请求读取接口的实现
type RequestWrappedReader struct {
	webx.WebContext
}

func (r *RequestWrappedReader) QueryValue(name string) string {
	return r.WebContext.QueryValues().Get(name)
}

func (r *RequestWrappedReader) PathValue(name string) string {
	return r.WebContext.PathValues().Get(name)
}

func (r *RequestWrappedReader) FormValue(name string) string {
	form, err := r.WebContext.FormValues()
	if nil != err {
		logger.Panicw("parse form value", "error", err)
		return ""
	} else {
		return form.Get(name)
	}
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

func (r *ResponseWrappedWriter) SetStatusCode(status int) flux.ResponseWriter {
	r.status = status
	return r
}

func (r *ResponseWrappedWriter) AddHeader(name, value string) flux.ResponseWriter {
	r.headers.Add(name, value)
	return r
}

func (r *ResponseWrappedWriter) SetHeaders(headers http.Header) flux.ResponseWriter {
	r.headers = headers
	return r
}

func (r *ResponseWrappedWriter) SetBody(body interface{}) flux.ResponseWriter {
	r.body = body
	return r
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
