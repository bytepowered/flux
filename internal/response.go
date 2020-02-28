package internal

import (
	"github.com/bytepowered/flux"
	"net/http"
)

// Response 定义响应数据
type response struct {
	status  int
	headers http.Header
	body    interface{}
}

func (r *response) StatusCode() int {
	return r.status
}

func (r *response) Headers() http.Header {
	return r.headers
}

func (r *response) Body() interface{} {
	return r.body
}

func (r *response) SetStatusCode(status int) flux.ResponseWriter {
	r.status = status
	return r
}

func (r *response) AddHeader(name, value string) flux.ResponseWriter {
	r.headers.Add(name, value)
	return r
}

func (r *response) SetHeaders(headers http.Header) flux.ResponseWriter {
	r.headers = headers
	return r
}

func (r *response) SetBody(body interface{}) flux.ResponseWriter {
	r.body = body
	return r
}

func (r *response) reset() {
	r.status = flux.StatusOK
	r.body = nil
	r.headers = http.Header{}
}

func newResponseWriter() *response {
	return &response{
		status:  flux.StatusOK,
		headers: http.Header{},
		body:    nil,
	}
}
