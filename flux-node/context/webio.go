package context

import (
	"github.com/bytepowered/flux/flux-node"
	"io"
	"net/http"
	"net/url"
)

var (
	_ flux.Request = new(WebRequest)
)

// WebRequest Request请求读取接口的实现
type WebRequest struct {
	flux.WebExchange
}

func (r *WebRequest) Method() string {
	return r.WebExchange.Method()
}

func (r *WebRequest) Host() string {
	return r.WebExchange.Host()
}

func (r *WebRequest) UserAgent() string {
	return r.WebExchange.UserAgent()
}

func (r *WebRequest) URI() string {
	return r.WebExchange.URI()
}

func (r *WebRequest) URL() *url.URL {
	return r.WebExchange.URL()
}

func (r *WebRequest) BodyReader() (io.ReadCloser, error) {
	return r.WebExchange.BodyReader()
}

func (r *WebRequest) Rewrite(method string, path string) {
	r.WebExchange.Rewrite(method, path)
}

func (r *WebRequest) HeaderVars() http.Header {
	return r.WebExchange.HeaderVars()
}

func (r *WebRequest) QueryVars() url.Values {
	return r.WebExchange.QueryVars()
}

func (r *WebRequest) PathVars() url.Values {
	return r.WebExchange.PathVars()
}

func (r *WebRequest) FormVars() url.Values {
	return r.WebExchange.FormVars()
}

func (r *WebRequest) CookieVars() []*http.Cookie {
	return r.WebExchange.CookieVars()
}

func (r *WebRequest) CookieVar(name string) *http.Cookie {
	return r.WebExchange.CookieVar(name)
}

func (r *WebRequest) QueryVar(name string) string {
	return r.WebExchange.QueryVar(name)
}

func (r *WebRequest) PathVar(name string) string {
	return r.WebExchange.PathVar(name)
}

func (r *WebRequest) FormVar(name string) string {
	return r.WebExchange.FormVar(name)
}

func (r *WebRequest) HeaderVar(name string) string {
	return r.WebExchange.HeaderVar(name)
}

func (r *WebRequest) reset(webex flux.WebExchange) {
	r.WebExchange = webex
}

func NewWebRequest() *WebRequest {
	return &WebRequest{}
}

////

var _ flux.Response = new(WebResponse)

// WebResponse 定义响应数据
type WebResponse struct {
	status  int
	headers http.Header
	payload interface{}
}

func (r *WebResponse) StatusCode() int {
	return r.status
}

func (r *WebResponse) HeaderVars() http.Header {
	return r.headers
}

func (r *WebResponse) SetStatusCode(status int) {
	r.status = status
}

func (r *WebResponse) AddHeader(name, value string) {
	r.headers.Add(name, value)
}

func (r *WebResponse) SetHeader(name, value string) {
	r.headers.Set(name, value)
}

func (r *WebResponse) SetPayload(payload interface{}) {
	r.payload = payload
}

func (r *WebResponse) Payload() interface{} {
	return r.payload
}

func (r *WebResponse) reset() {
	r.status = flux.StatusOK
	r.payload = nil
	r.headers = http.Header{}
}

func NewWebResponse() *WebResponse {
	return &WebResponse{
		status:  flux.StatusOK,
		payload: nil,
		headers: http.Header{},
	}
}
