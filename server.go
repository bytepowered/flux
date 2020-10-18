package flux

import "net/http"

type (
	// ServerContextExchangeHook
	ServerContextExchangeHook func(WebContext, Context)
	// 写入Error错误响应数据到WebServer
	ServerErrorsWriter func(webc WebContext, requestId string, header http.Header, error *StateError) error
	// 写入Body正常响应数据到WebServer
	ServerResponseWriter func(webc WebContext, requestId string, header http.Header, status int, body interface{}) error
)
