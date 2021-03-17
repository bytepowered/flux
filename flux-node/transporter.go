package flux

import (
	"net/http"
)

type (
	// Transporter 表示某种特定协议的后端服务，例如Dubbo, gRPC, Http等协议的后端服务。
	// 默认实现了Dubbo(gRpc)和Http两种协议。
	Transporter interface {
		// Invoke 真正执行指定目标EndpointService的通讯，返回响应结果
		Invoke(*Context, TransporterService) (interface{}, *ServeError)
		// InvokeCodec 执行指定目标EndpointService的通讯，返回响应结果，并解析响应数据
		InvokeCodec(*Context, TransporterService) (*ResponseBody, *ServeError)
		// Transport 完成前端Http请求与后端服务的数据交互
		Transport(*Context)
		// Writer
		Writer() TransportWriter
	}
	// TransportCodec 解析 Transporter 返回的原始数据，生成响应对象
	TransportCodec func(ctx *Context, packet interface{}) (*ResponseBody, error)
	// TransportWriter
	TransportWriter interface {
		Write(ctx *Context, response *ResponseBody)
		WriteError(ctx *Context, err *ServeError)
	}
	// ResponseBody 后端服务返回统一响应数据结构
	ResponseBody struct {
		StatusCode  int                    // Http状态码
		Headers     http.Header            // Header
		Attachments map[string]interface{} // Attachment
		Body        interface{}            // 响应数据体
	}
)
