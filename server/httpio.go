package server

import (
	"github.com/bytepowered/flux"
	"io"
	"net/http"
	"net/url"
)

var _ flux.RequestReader = new(WrappedRequestReader)

// WrappedRequestReader Request请求读取接口的实现
type WrappedRequestReader struct {
	flux.WebContext
}

func (r *WrappedRequestReader) Method() string {
	return r.WebContext.Method()
}

func (r *WrappedRequestReader) Host() string {
	return r.WebContext.Host()
}

func (r *WrappedRequestReader) UserAgent() string {
	return r.WebContext.UserAgent()
}

func (r *WrappedRequestReader) RequestURI() string {
	return r.WebContext.RequestURI()
}

func (r *WrappedRequestReader) RequestURL() (url *url.URL, writable bool) {
	return r.WebContext.RequestURL()
}

func (r *WrappedRequestReader) RequestBodyReader() (io.ReadCloser, error) {
	return r.WebContext.RequestBodyReader()
}

func (r *WrappedRequestReader) RequestRewrite(method string, path string) {
	r.WebContext.RequestRewrite(method, path)
}

func (r *WrappedRequestReader) HeaderValues() (header http.Header, writable bool) {
	return r.WebContext.HeaderValues()
}

func (r *WrappedRequestReader) QueryValues() url.Values {
	return r.WebContext.QueryValues()
}

func (r *WrappedRequestReader) PathValues() url.Values {
	return r.WebContext.PathValues()
}

func (r *WrappedRequestReader) FormValues() url.Values {
	return r.WebContext.FormValues()
}

func (r *WrappedRequestReader) CookieValues() []*http.Cookie {
	return r.WebContext.CookieValues()
}

func (r *WrappedRequestReader) CookieValue(name string) (cookie *http.Cookie, ok bool) {
	return r.WebContext.CookieValue(name)
}

func (r *WrappedRequestReader) QueryValue(name string) string {
	return r.WebContext.QueryValue(name)
}

func (r *WrappedRequestReader) PathValue(name string) string {
	return r.WebContext.PathValue(name)
}

func (r *WrappedRequestReader) FormValue(name string) string {
	return r.WebContext.FormValue(name)
}

func (r *WrappedRequestReader) HeaderValue(name string) string {
	return r.WebContext.HeaderValue(name)
}

func (r *WrappedRequestReader) reattach(webex flux.WebContext) {
	r.WebContext = webex
}

func (r *WrappedRequestReader) reset() {
	r.WebContext = nil
}

func newRequestWrappedReader() *WrappedRequestReader {
	return &WrappedRequestReader{}
}

////

var _ flux.ResponseWriter = new(WrappedResponseWriter)

// WrappedResponseWriter 定义响应数据
type WrappedResponseWriter struct {
	status  int
	headers http.Header
	body    interface{}
}

func (r *WrappedResponseWriter) StatusCode() int {
	return r.status
}

func (r *WrappedResponseWriter) HeaderValues() http.Header {
	return r.headers
}

func (r *WrappedResponseWriter) Body() interface{} {
	return r.body
}

func (r *WrappedResponseWriter) SetStatusCode(status int) {
	r.status = status
}

func (r *WrappedResponseWriter) AddHeader(name, value string) {
	r.headers.Add(name, value)
}

func (r *WrappedResponseWriter) SetHeader(name, value string) {
	r.headers.Set(name, value)
}

func (r *WrappedResponseWriter) SetHeaders(headers http.Header) {
	r.headers = headers
}

func (r *WrappedResponseWriter) SetBody(body interface{}) {
	r.body = body
}

func (r *WrappedResponseWriter) reset() {
	r.status = flux.StatusOK
	r.body = nil
	r.headers = http.Header{}
}

func newResponseWrappedWriter() *WrappedResponseWriter {
	return &WrappedResponseWriter{
		status:  flux.StatusOK,
		headers: http.Header{},
		body:    nil,
	}
}
