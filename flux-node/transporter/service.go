package transporter

import (
	"fmt"
	"github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-node/common"
	"github.com/bytepowered/flux/flux-node/ext"
	"github.com/bytepowered/flux/flux-pkg"
	"github.com/spf13/cast"
)

func DoTransport(ctx *flux.Context, transport flux.Transporter) {
	response, err := transport.InvokeCodec(ctx, ctx.Transporter())
	if err != nil {
		ctx.Logger().Errorw("TRANSPORTER:INVOKE/ERROR", "error", err)
		transport.Writer().WriteError(ctx, err)
		return
	}
	// response
	if response == nil {
		select {
		case <-ctx.Context().Done():
			ctx.Logger().Warnw("TRANSPORTER:CANCELED/BYCLIENT")
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
	transport, ok := ext.TransporterBy(proto)
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

var _ flux.TransportWriter = new(DefaultTransportWriter)

type DefaultTransportWriter int

func (r *DefaultTransportWriter) Write(ctx *flux.Context, response *flux.ResponseBody) {
	header := ctx.ResponseWriter().Header()
	for k, hv := range response.Headers {
		for _, v := range hv {
			header.Add(k, v)
		}
	}
	if bytes, err := common.SerializeObject(response.Body); nil != err {
		r.WriteError(ctx, &flux.ServeError{
			StatusCode: flux.StatusServerError,
			Message:    flux.ErrorMessageTransportDecodeResponse,
			CauseError: err,
		})
	} else {
		r.write(ctx, response.StatusCode, bytes)
	}
}

func (r *DefaultTransportWriter) WriteError(ctx *flux.Context, err *flux.ServeError) {
	bytes, _ := common.SerializeObject(map[string]interface{}{
		"status":  "error",
		"code":    err.ErrorCode,
		"message": err.Message,
		"error":   cast.ToString(err.CauseError),
	})
	r.write(ctx, err.StatusCode, bytes)
}

func (r *DefaultTransportWriter) write(ctx *flux.Context, status int, body []byte) {
	ctx.Request().Header.Add("X-Writer-Id", "Fx-TWriter")
	err := ctx.Write(status, flux.MIMEApplicationJSONCharsetUTF8, body)
	if nil != err {
		ctx.Logger().Errorw("TRANSPORT:WRITE:ERROR", "error", err)
	} else {
		ctx.Logger().Infow("TRANSPORT:WRITE:COMPLETED", "body", string(body))
	}
}
