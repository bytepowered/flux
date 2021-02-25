package context

import (
	"github.com/bytepowered/flux"
	"io"
	"net/http"
	"net/url"
)

var (
	_ flux.Request = new(DefaultRequest)
)

// DefaultRequest Request请求读取接口的实现
type DefaultRequest struct {
	flux.WebContext
}

func (r *DefaultRequest) Method() string {
	return r.WebContext.Method()
}

func (r *DefaultRequest) Host() string {
	return r.WebContext.Host()
}

func (r *DefaultRequest) UserAgent() string {
	return r.WebContext.UserAgent()
}

func (r *DefaultRequest) URI() string {
	return r.WebContext.URI()
}

func (r *DefaultRequest) URL() *url.URL {
	return r.WebContext.URL()
}

func (r *DefaultRequest) BodyReader() (io.ReadCloser, error) {
	return r.WebContext.BodyReader()
}

func (r *DefaultRequest) Rewrite(method string, path string) {
	r.WebContext.Rewrite(method, path)
}

func (r *DefaultRequest) HeaderVars() http.Header {
	return r.WebContext.HeaderVars()
}

func (r *DefaultRequest) QueryVars() url.Values {
	return r.WebContext.QueryVars()
}

func (r *DefaultRequest) PathVars() url.Values {
	return r.WebContext.PathVars()
}

func (r *DefaultRequest) FormVars() url.Values {
	return r.WebContext.FormVars()
}

func (r *DefaultRequest) CookieVars() []*http.Cookie {
	return r.WebContext.CookieVars()
}

func (r *DefaultRequest) CookieVar(name string) *http.Cookie {
	return r.WebContext.CookieVar(name)
}

func (r *DefaultRequest) QueryVar(name string) string {
	return r.WebContext.QueryVar(name)
}

func (r *DefaultRequest) PathVar(name string) string {
	return r.WebContext.PathVar(name)
}

func (r *DefaultRequest) FormVar(name string) string {
	return r.WebContext.FormVar(name)
}

func (r *DefaultRequest) HeaderVar(name string) string {
	return r.WebContext.HeaderVar(name)
}

func (r *DefaultRequest) attach(webex flux.WebContext) {
	r.WebContext = webex
}

func (r *DefaultRequest) release() {
	r.WebContext = nil
}

func NewDefaultRequest() *DefaultRequest {
	return &DefaultRequest{}
}

////

var _ flux.Response = new(DefaultResponse)

// DefaultResponse 定义响应数据
type DefaultResponse struct {
	status  int
	headers http.Header
	payload interface{}
}

func (r *DefaultResponse) StatusCode() int {
	return r.status
}

func (r *DefaultResponse) HeaderVars() http.Header {
	return r.headers
}

func (r *DefaultResponse) SetStatusCode(status int) {
	r.status = status
}

func (r *DefaultResponse) AddHeader(name, value string) {
	r.headers.Add(name, value)
}

func (r *DefaultResponse) SetHeader(name, value string) {
	r.headers.Set(name, value)
}

func (r *DefaultResponse) SetPayload(payload interface{}) {
	r.payload = payload
}

func (r *DefaultResponse) Payload() interface{} {
	return r.payload
}

func (r *DefaultResponse) reset() {
	r.status = flux.StatusOK
	r.payload = nil
	r.headers = http.Header{}
}

func NewDefaultResponse() *DefaultResponse {
	return &DefaultResponse{
		status:  flux.StatusOK,
		headers: http.Header{},
		payload: nil,
	}
}
