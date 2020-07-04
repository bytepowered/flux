package flux

import (
	"io"
	"net/http"
)

const (
	StatusOK           = http.StatusOK
	StatusBadRequest   = http.StatusBadRequest
	StatusNotFound     = http.StatusNotFound
	StatusUnauthorized = http.StatusUnauthorized
	StatusAccessDenied = http.StatusForbidden
	StatusServerError  = http.StatusInternalServerError
	StatusBadGateway   = http.StatusBadGateway
)

// RequestReader 定义请求参数读取接口
type RequestReader interface {
	Headers() http.Header             // 获取Http请求的全部Header
	QueryValue(name string) string    // 获取Http请求的Query参数
	PathValue(name string) string     // 获取Http请求的Path路径参数
	FormValue(name string) string     // 获取Http请求的Form表单参数
	HeaderValue(name string) string   // 获取Http请求的Header参数
	CookieValue(name string) string   // 获取Http请求的Cookie参数
	RemoteAddress() string            // 获取Http请求的远程地址
	HttpBody() (io.ReadCloser, error) // 返回Http请求的Body可重复读取的接口
	HttpRequest() *http.Request       // Http原始Request
}

// ResponseWriter 是写入响应数据的接口
type ResponseWriter interface {
	SetStatusCode(status int) ResponseWriter       // SetStatusCode 设置Http响应状态码
	StatusCode() int                               // StatusCode 获取Http响应状态码
	AddHeader(name, value string) ResponseWriter   // AddHeader 添加Header
	SetHeaders(headers http.Header) ResponseWriter // SetHeaders 设置全部Headers
	Headers() http.Header                          // Header 获取设置的Headers
	SetBody(body interface{}) ResponseWriter       // SetBody 设置数据响应体
	Body() interface{}                             // Body 响应数据体
}
