package backend

import (
	"fmt"
	flux2 "github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-node/ext"
	"github.com/bytepowered/flux/flux-pkg"
)

func DoExchangeTransport(ctx flux2.Context, transport flux2.BackendTransport) *flux2.ServeError {
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
	fluxpkg.AssertNotNil(response,
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
func DoInvokeCodec(ctx flux2.Context, service flux2.BackendService) (*flux2.BackendResponse, *flux2.ServeError) {
	proto := service.AttrRpcProto()
	transport, ok := ext.BackendTransportByProto(proto)
	if !ok {
		return nil, &flux2.ServeError{
			StatusCode: flux2.StatusServerError,
			ErrorCode:  flux2.ErrorCodeGatewayInternal,
			Message:    flux2.ErrorMessageProtocolUnknown,
			CauseError: fmt.Errorf("unknown rpc protocol:%s", proto),
		}
	}
	return transport.InvokeCodec(ctx, service)
}
