package flux

import (
	"net/http"
)

// BackendTransporter 表示某种特定协议的后端服务，例如Dubbo, gRPC, Http等协议的后端服务。
// 默认实现了Dubbo(gRpc)和Http两种协议。
type (
	BackendTransporter interface {
		// Transport 完成前端Http请求与后端服务的数据交互
		Transport(*Context) *ServeError
		// Invoke 真正执行指定目标EndpointService的通讯，返回响应结果
		Invoke(*Context, TransporterService) (interface{}, *ServeError)
		// InvokeCodec 执行指定目标EndpointService的通讯，返回响应结果，并解析响应数据
		InvokeCodec(*Context, TransporterService) (*BackendResponse, *ServeError)
		// GetBackendCodecFunc 获取响应结果解析函数
		GetBackendCodecFunc() BackendCodecFunc
	}
	// BackendCodecFunc 解析Backend返回的原始数据
	BackendCodecFunc func(ctx *Context, raw interface{}) (*BackendResponse, error)
)

// BackendResponse 后端服务返回统一响应数据结构
type BackendResponse struct {
	// Http状态码
	StatusCode int
	// Header
	Headers http.Header
	// Attachment
	Attachments map[string]interface{}
	// 响应数据体
	Body interface{}
}

func (b *BackendResponse) GetStatusCode() int {
	return b.StatusCode
}

func (b *BackendResponse) GetHeaders() http.Header {
	return b.Headers
}

func (b *BackendResponse) GetAttachments() map[string]interface{} {
	return b.Attachments
}

func (b *BackendResponse) GetBody() interface{} {
	return b.Body
}
