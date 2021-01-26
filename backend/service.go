package backend

import (
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
)

func DoExchangeTransport(ctx flux.Context, transport flux.BackendTransport) *flux.ServeError {
	result, err := transport.InvokeCodec(ctx, ctx.Service())
	if err != nil {
		return err
	}
	// attachments
	for k, v := range result.Attachments {
		ctx.SetAttribute(k, v)
	}
	writer := ctx.Response()
	writer.SetStatusCode(result.StatusCode)
	writer.SetHeaders(result.Headers)
	writer.SetBody(result.Body)
	return nil
}

// DoInvokeCodec 执行后端服务，获取响应结果；
func DoInvokeCodec(ctx flux.Context, service flux.BackendService) (*flux.BackendResponse, *flux.ServeError) {
	rpcProto := service.AttrRpcProto()
	transport, ok := ext.LoadBackendTransport(rpcProto)
	if !ok {
		return nil, &flux.ServeError{
			StatusCode: flux.StatusServerError,
			ErrorCode:  flux.ErrorCodeGatewayInternal,
			Message:    "GATEWAY:UNKNOWN_PROTOCOL",
			Internal:   fmt.Errorf("unknown protocol:%s", rpcProto),
		}
	}
	return transport.InvokeCodec(ctx, service)
}
