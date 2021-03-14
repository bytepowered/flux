package backend

import (
	"errors"
	"fmt"
	"github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-node/common"
	"github.com/bytepowered/flux/flux-node/ext"
	"github.com/bytepowered/flux/flux-pkg"
	"strconv"
)

func DoTransport(ctx *flux.Context, transport flux.BackendTransporter) *flux.ServeError {
	response, err := transport.InvokeCodec(ctx, ctx.Transporter())
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
	fluxpkg.AssertNotNil(response, "exchange: <response> must-not nil, request-id: "+ctx.RequestId())
	// attachments
	for k, v := range response.Attachments {
		ctx.SetAttribute(k, v)
	}
	headers := ctx.ResponseWriter().Header()
	for k, vs := range response.Headers {
		for _, v := range vs {
			headers.Add(k, v)
		}
	}
	bytes, serr := common.SerializeObject(response.Body)
	if nil != serr {
		return &flux.ServeError{
			StatusCode: flux.StatusServerError,
			ErrorCode:  flux.ErrorCodeGatewayInternal,
			Message:    flux.ErrorMessageBackendDecodeResponse,
			CauseError: errors.New("failed to decode response body"),
		}
	}
	// Write bytes
	headers.Set(flux.HeaderContentLength, strconv.Itoa(len(bytes)))
	serr = ctx.Write(response.StatusCode, flux.MIMEApplicationJSONCharsetUTF8, bytes)
	if nil != serr {
		return &flux.ServeError{
			StatusCode: flux.StatusServerError,
			ErrorCode:  flux.ErrorCodeGatewayInternal,
			Message:    flux.ErrorMessageBackendWriteResponse,
			CauseError: errors.New("failed to write response data"),
		}
	}
	return nil
}

// DoInvokeCodec 执行后端服务，获取响应结果；
func DoInvokeCodec(ctx *flux.Context, service flux.TransporterService) (*flux.BackendResponse, *flux.ServeError) {
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
