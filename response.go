package flux

import (
	"github.com/bytepowered/flux/webex"
	h2tp "net/http"
)

// HttpResponseWriter 实现将错误消息和响应数据写入Response实例
type HttpResponseWriter interface {
	WriteError(webex webex.WebContext, requestId string, header h2tp.Header, error *StateError) error
	WriteBody(webex webex.WebContext, requestId string, header h2tp.Header, status int, body interface{}) error
}
