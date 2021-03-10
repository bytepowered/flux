package context

import (
	"github.com/bytepowered/flux"
	"io"
	"net/http"
	"net/url"
)

var (
	_ flux.Request = new(AttacheRequest)
)

// AttacheRequest Request请求读取接口的实现
type AttacheRequest struct {
	flux.WebExchange
}

func (r *AttacheRequest) Method() string {
	return r.WebExchange.Method()
}

func (r *AttacheRequest) Host() string {
	return r.WebExchange.Host()
}

func (r *AttacheRequest) UserAgent() string {
	return r.WebExchange.UserAgent()
}

func (r *AttacheRequest) URI() string {
	return r.WebExchange.URI()
}

func (r *AttacheRequest) URL() *url.URL {
	return r.WebExchange.URL()
}

func (r *AttacheRequest) BodyReader() (io.ReadCloser, error) {
	return r.WebExchange.BodyReader()
}

func (r *AttacheRequest) Rewrite(method string, path string) {
	r.WebExchange.Rewrite(method, path)
}

func (r *AttacheRequest) HeaderVars() http.Header {
	return r.WebExchange.HeaderVars()
}

func (r *AttacheRequest) QueryVars() url.Values {
	return r.WebExchange.QueryVars()
}

func (r *AttacheRequest) PathVars() url.Values {
	return r.WebExchange.PathVars()
}

func (r *AttacheRequest) FormVars() url.Values {
	return r.WebExchange.FormVars()
}

func (r *AttacheRequest) CookieVars() []*http.Cookie {
	return r.WebExchange.CookieVars()
}

func (r *AttacheRequest) CookieVar(name string) *http.Cookie {
	return r.WebExchange.CookieVar(name)
}

func (r *AttacheRequest) QueryVar(name string) string {
	return r.WebExchange.QueryVar(name)
}

func (r *AttacheRequest) PathVar(name string) string {
	return r.WebExchange.PathVar(name)
}

func (r *AttacheRequest) FormVar(name string) string {
	return r.WebExchange.FormVar(name)
}

func (r *AttacheRequest) HeaderVar(name string) string {
	return r.WebExchange.HeaderVar(name)
}

func (r *AttacheRequest) reset(webex flux.WebExchange) {
	r.WebExchange = webex
}

func (r *AttacheRequest) release() {
	r.WebExchange = nil
}

func NewAttacheRequest() *AttacheRequest {
	return &AttacheRequest{}
}

////

var _ flux.Response = new(AttacheResponse)

// AttacheResponse 定义响应数据
type AttacheResponse struct {
	status  int
	headers http.Header
	payload interface{}
}

func (r *AttacheResponse) StatusCode() int {
	return r.status
}

func (r *AttacheResponse) HeaderVars() http.Header {
	return r.headers
}

func (r *AttacheResponse) SetStatusCode(status int) {
	r.status = status
}

func (r *AttacheResponse) AddHeader(name, value string) {
	r.headers.Add(name, value)
}

func (r *AttacheResponse) SetHeader(name, value string) {
	r.headers.Set(name, value)
}

func (r *AttacheResponse) SetPayload(payload interface{}) {
	r.payload = payload
}

func (r *AttacheResponse) Payload() interface{} {
	return r.payload
}

func (r *AttacheResponse) reset() {
	r.status = flux.StatusOK
	r.payload = nil
	r.headers = http.Header{}
}

func NewAttacheResponse() *AttacheResponse {
	dr := new(AttacheResponse)
	dr.reset()
	return dr
}
