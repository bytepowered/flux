package backend

import (
	"fmt"
	"github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-node/ext"
	"github.com/bytepowered/flux/flux-pkg"
)

func DoTransport(ctx *flux.Context, transport flux.Transporter) {
	response, err := transport.InvokeCodec(ctx, ctx.Transporter())
	if err != nil {
		ctx.Logger().Errorw("BACKEND:TRANSPORT:INVOKE/ERROR", "error", err)
		transport.Writer().WriteError(ctx, err)
		return
	}
	// response
	if response == nil {
		select {
		case <-ctx.Context().Done():
			ctx.Logger().Warnw("BACKEND:TRANSPORT:CANCELED/BYCLIENT")
			return
		default:
			break
		}
	}
	fluxpkg.AssertNotNil(response, "exchange: <response> must-not nil, request-id: "+ctx.RequestId())
	// attachments
	for k, v := range response.Attachments {
		ctx.SetAttribute(k, v)
	}
	transport.Writer().Write(ctx, response)
}

// DoInvokeCodec 执行后端服务，获取响应结果；
func DoInvokeCodec(ctx *flux.Context, service flux.TransporterService) (*flux.ResponseBody, *flux.ServeError) {
	proto := service.RpcProto()
	transport, ok := ext.BackendTransporterBy(proto)
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

var _ flux.TransportWriter = new(RpcTransportWriter)

type RpcTransportWriter int

func (r *RpcTransportWriter) Write(ctx *flux.Context, response *flux.ResponseBody) {
	panic("implement me")
}

func (r *RpcTransportWriter) WriteError(ctx *flux.Context, err *flux.ServeError) {
	panic("implement me")
}
