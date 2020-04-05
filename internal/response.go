package internal

import (
	"github.com/bytepowered/flux"
	"net/http"
)

// FxResponse 定义响应数据
type FxResponse struct {
	status  int
	headers http.Header
	body    interface{}
}

func (r *FxResponse) StatusCode() int {
	return r.status
}

func (r *FxResponse) Headers() http.Header {
	return r.headers
}

func (r *FxResponse) Body() interface{} {
	return r.body
}

func (r *FxResponse) SetStatusCode(status int) flux.ResponseWriter {
	r.status = status
	return r
}

func (r *FxResponse) AddHeader(name, value string) flux.ResponseWriter {
	r.headers.Add(name, value)
	return r
}

func (r *FxResponse) SetHeaders(headers http.Header) flux.ResponseWriter {
	r.headers = headers
	return r
}

func (r *FxResponse) SetBody(body interface{}) flux.ResponseWriter {
	r.body = body
	return r
}

func (r *FxResponse) reset() {
	r.status = flux.StatusOK
	r.body = nil
	r.headers = http.Header{}
}

func newResponseWriter() *FxResponse {
	return &FxResponse{
		status:  flux.StatusOK,
		headers: http.Header{},
		body:    nil,
	}
}
