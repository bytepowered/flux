package transporter

import (
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/fluxkit"
	"github.com/spf13/cast"
)

func DoTransport(ctx *flux.Context, transport flux.Transporter) {
	response, serr := transport.InvokeCodec(ctx, ctx.Service())
	select {
	case <-ctx.Context().Done():
		ctx.Logger().Warnw("TRANSPORTER:CANCELED/BYCLIENT")
		return
	default:
		break
	}
	if serr != nil {
		ctx.Logger().Errorw("TRANSPORTER:INVOKE/ERROR", "error", serr)
		transport.Writer().WriteError(ctx, serr)
	} else {
		fluxkit.AssertNotNil(response, "exchange: <response> must-not nil, request-id: "+ctx.RequestId())
		for k, v := range response.Attachments {
			ctx.SetAttribute(k, v)
		}
		transport.Writer().Write(ctx, response)
	}
}

// DoInvokeCodec 执行后端服务，获取响应结果；
func DoInvokeCodec(ctx *flux.Context, service flux.Service) (*flux.ResponseBody, *flux.ServeError) {
	proto := service.RpcProto()
	transport, ok := ext.TransporterByProto(proto)
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

// DefaultTransportWriter

var _ flux.TransportWriter = new(DefaultTransportWriter)

type DefaultTransportWriter int

func (r *DefaultTransportWriter) Write(ctx *flux.Context, response *flux.ResponseBody) {
	header := ctx.ResponseWriter().Header()
	for k, hv := range response.Headers {
		for _, v := range hv {
			header.Add(k, v)
		}
	}
	if bytes, err := ext.JSONMarshalObject(response.Body); nil != err {
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
	bytes, _ := ext.JSONMarshalObject(map[string]interface{}{
		"status":  "error",
		"code":    err.ErrorCode,
		"message": err.Message,
		"error":   cast.ToString(err.CauseError),
	})
	r.write(ctx, err.StatusCode, bytes)
}

func (r *DefaultTransportWriter) write(ctx *flux.Context, status int, body []byte) {
	ctx.ResponseWriter().Header().Add("X-Writer-Id", "Fx-TWriter")
	err := ctx.Write(status, flux.MIMEApplicationJSONCharsetUTF8, body)
	if nil != err {
		ctx.Logger().Errorw("TRANSPORT:WRITE:ERROR", "error", err)
	} else {
		ctx.Logger().Infow("TRANSPORT:WRITE:COMPLETED", "body", string(body))
	}
}
