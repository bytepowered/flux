package flux

import "net/http"

const (
	StatusOK                 = 200
	StatusBadRequest         = 400
	StatusNotFound           = 404
	StatusUnauthorized       = 401
	StatusAccessDenied       = 403
	StatusServerError        = 500
	StatusBadGateway         = 502
	StatusServiceUnavailable = 503
)

// ResponseWriter 是写入响应数据的接口
type ResponseWriter interface {
	SetStatusCode(status int) ResponseWriter       // SetStatusCode 设置Http响应状态码
	StatusCode() int                               // StatusCode 获取Http响应状态码
	AddHeader(name, value string) ResponseWriter   // AddHeader 添加Header
	SetHeaders(headers http.Header) ResponseWriter // SetHeaders 设置全部Headers
	Headers() http.Header                          // Header 获取设置的Headers
	SetBody(body interface{}) ResponseWriter       // SetBody 设置数据响应体
	Body() interface{}                             // Body 设置响应数据体
}
