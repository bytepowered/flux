package flux

import (
	"net/http"
)

// BackendTransport 表示某种特定协议的后端服务，例如Dubbo, gRPC, Http等协议的后端服务。
// 默认实现了Dubbo(gRpc)和Http两种协议。
type BackendTransport interface {
	// Exchange 完成前端Http请求与后端服务的数据交互
	Exchange(Context) *ServeError
	// Invoke 真正执行指定目标EndpointService的通讯，返回响应结果
	Invoke(BackendService, Context) (interface{}, *ServeError)
	// GetResultDecodeFunc 获取响应结果解析函数
	GetResultDecodeFunc() BackendResultDecodeFunc
}

// BackendResult 后端服务返回统一响应数据结构
type BackendResult struct {
	// Http状态码
	StatusCode int
	// Header
	Headers http.Header
	// Attachment
	Attachments map[string]interface{}
	// 响应数据体
	Body interface{}
}

// BackendResultDecodeFunc 解析Backend返回的数据
type BackendResultDecodeFunc func(ctx Context, response interface{}) (*BackendResult, error)
