package flux

import "net/http"

type (
	// ServerContextHookFunc 用于WebContext与Context的交互勾子；
	// 在每个请求被路由执行时，在创建Context后被调用。
	ServerContextHookFunc func(WebContext, Context)
	// ServerErrorsWriter 用于写入Error错误响应数据到WebServer
	ServerErrorsWriter func(webc WebContext, requestId string, header http.Header, error *ServeError) error
	// ServerResponseWriter 用于写入Body正常响应数据到WebServer
	ServerResponseWriter func(webc WebContext, requestId string, header http.Header, status int, body interface{}) error
)
