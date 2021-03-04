package backend

import (
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/pkg"
)

func DoExchangeTransport(ctx flux.Context, transport flux.BackendTransport) *flux.ServeError {
	response, err := transport.InvokeCodec(ctx, ctx.BackendService())
	if err != nil {
		return err
	}
	// response
	if response == nil {
		select {
		case <-ctx.Context().Done():
			return nil
		default:
			break
		}
	}
	pkg.AssertNotNil(response,
		"exchange: <response> must-not nil, request-id: "+ctx.RequestId())
	// attachments
	for k, v := range response.Attachments {
		ctx.SetAttribute(k, v)
	}
	writer := ctx.Response()
	writer.SetStatusCode(response.StatusCode)
	for k, vs := range response.Headers {
		for _, v := range vs {
			writer.AddHeader(k, v)
		}
	}
	writer.SetPayload(response.Body)
	return nil
}

// DoInvokeCodec 执行后端服务，获取响应结果；
func DoInvokeCodec(ctx flux.Context, service flux.BackendService) (*flux.BackendResponse, *flux.ServeError) {
	proto := service.AttrRpcProto()
	transport, ok := ext.GetBackendTransport(proto)
	if !ok {
		return nil, &flux.ServeError{
			StatusCode: flux.StatusServerError,
			ErrorCode:  flux.ErrorCodeGatewayInternal,
			Message:    flux.ErrorMessageProtocolUnknown,
			CauseError: fmt.Errorf("unknown rpc protocol:%s", proto),
		}
	}
	return transport.InvokeCodec(ctx, service)
}
