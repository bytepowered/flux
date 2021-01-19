package server

import (
	"github.com/bytepowered/flux"
	"io"
	"net/http"
	"net/url"
)

var _ flux.RequestReader = new(DefaultRequestReader)

// DefaultRequestReader Request请求读取接口的实现
type DefaultRequestReader struct {
	flux.WebContext
}

func (r *DefaultRequestReader) Method() string {
	return r.WebContext.Method()
}

func (r *DefaultRequestReader) Host() string {
	return r.WebContext.Host()
}

func (r *DefaultRequestReader) UserAgent() string {
	return r.WebContext.UserAgent()
}

func (r *DefaultRequestReader) RequestURI() string {
	return r.WebContext.RequestURI()
}

func (r *DefaultRequestReader) RequestURL() (url *url.URL, writable bool) {
	return r.WebContext.RequestURL()
}

func (r *DefaultRequestReader) RequestBodyReader() (io.ReadCloser, error) {
	return r.WebContext.RequestBodyReader()
}

func (r *DefaultRequestReader) RequestRewrite(method string, path string) {
	r.WebContext.RequestRewrite(method, path)
}

func (r *DefaultRequestReader) HeaderValues() (header http.Header, writable bool) {
	return r.WebContext.HeaderValues()
}

func (r *DefaultRequestReader) QueryValues() url.Values {
	return r.WebContext.QueryValues()
}

func (r *DefaultRequestReader) PathValues() url.Values {
	return r.WebContext.PathValues()
}

func (r *DefaultRequestReader) FormValues() url.Values {
	return r.WebContext.FormValues()
}

func (r *DefaultRequestReader) CookieValues() []*http.Cookie {
	return r.WebContext.CookieValues()
}

func (r *DefaultRequestReader) CookieValue(name string) (cookie *http.Cookie, ok bool) {
	return r.WebContext.CookieValue(name)
}

func (r *DefaultRequestReader) QueryValue(name string) string {
	return r.WebContext.QueryValue(name)
}

func (r *DefaultRequestReader) PathValue(name string) string {
	return r.WebContext.PathValue(name)
}

func (r *DefaultRequestReader) FormValue(name string) string {
	return r.WebContext.FormValue(name)
}

func (r *DefaultRequestReader) HeaderValue(name string) string {
	return r.WebContext.HeaderValue(name)
}

func (r *DefaultRequestReader) reattach(webex flux.WebContext) {
	r.WebContext = webex
}

func (r *DefaultRequestReader) reset() {
	r.WebContext = nil
}

func NewDefaultRequestReader() *DefaultRequestReader {
	return &DefaultRequestReader{}
}

////

var _ flux.ResponseWriter = new(DefaultResponseWriter)

// DefaultResponseWriter 定义响应数据
type DefaultResponseWriter struct {
	status  int
	headers http.Header
	body    interface{}
}

func (r *DefaultResponseWriter) StatusCode() int {
	return r.status
}

func (r *DefaultResponseWriter) HeaderValues() http.Header {
	return r.headers
}

func (r *DefaultResponseWriter) Body() interface{} {
	return r.body
}

func (r *DefaultResponseWriter) SetStatusCode(status int) {
	r.status = status
}

func (r *DefaultResponseWriter) AddHeader(name, value string) {
	r.headers.Add(name, value)
}

func (r *DefaultResponseWriter) SetHeader(name, value string) {
	r.headers.Set(name, value)
}

func (r *DefaultResponseWriter) SetHeaders(headers http.Header) {
	r.headers = headers
}

func (r *DefaultResponseWriter) SetBody(body interface{}) {
	r.body = body
}

func (r *DefaultResponseWriter) reset() {
	r.status = flux.StatusOK
	r.body = nil
	r.headers = http.Header{}
}

func NewDefaultResponseWriter() *DefaultResponseWriter {
	return &DefaultResponseWriter{
		status:  flux.StatusOK,
		headers: http.Header{},
		body:    nil,
	}
}
