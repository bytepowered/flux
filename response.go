package flux

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

// HttpResponseWriter 实现将错误消息和响应数据写入Response实例
type HttpResponseWriter interface {
	WriteError(ctx echo.Context, requestId string, header http.Header, error *InvokeError) error
	WriteBody(ctx echo.Context, requestId string, header http.Header, status int, body interface{}) error
}
