package flux

import (
	"net/http"
)

// Backend 表示某种特定协议的后端服务，例如Dubbo, gRPC, Http等协议的后端服务。
// 默认实现了Dubbo(gRpc)和Http两种协议。
type Backend interface {
	// Exchange 完成前端Http请求与后端服务的数据交互
	Exchange(Context) *StateError
	// Invoke 真正执行指定目标Endpoint的通讯，返回响应结果
	Invoke(*Endpoint, Context) (interface{}, *StateError)
}

// BackendResponseDecoder 解析Backend返回的数据
type BackendResponseDecoder func(ctx Context, response interface{}) (statusCode int, headers http.Header, body interface{}, err error)
