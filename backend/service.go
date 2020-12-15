package backend

import (
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
)

var (
	ErrBackendTransportDecodeFuncNotFound = &flux.ServeError{
		StatusCode: flux.StatusServerError,
		ErrorCode:  flux.ErrorCodeGatewayInternal,
		Message:    flux.ErrorMessageBackendDecoderNotFound,
	}
)

func DoExchange(ctx flux.Context, exchange flux.BackendTransport) *flux.ServeError {
	endpoint := ctx.Endpoint()
	resp, err := exchange.Invoke(endpoint.Service, ctx)
	if err != nil {
		return err
	}
	// decode responseWriter
	decoder, ok := ext.LoadBackendTransportDecodeFunc(endpoint.Service.RpcProto)
	if !ok {
		return ErrBackendTransportDecodeFuncNotFound
	}
	if code, headers, body, err := decoder(ctx, resp); nil == err {
		ctx.Response().SetStatusCode(code)
		ctx.Response().SetHeaders(headers)
		ctx.Response().SetBody(body)
		return nil
	} else {
		return &flux.ServeError{
			StatusCode: flux.StatusServerError,
			ErrorCode:  flux.ErrorCodeGatewayInternal,
			Message:    flux.ErrorMessageBackendDecodeResponse,
			Internal:   err,
		}
	}
}

// DoInvoke 执行后端服务，获取响应结果；
func DoInvoke(service flux.BackendService, ctx flux.Context) (interface{}, *flux.ServeError) {
	backend, ok := ext.LoadBackend(service.RpcProto)
	if !ok {
		return nil, &flux.ServeError{
			StatusCode: flux.StatusServerError,
			ErrorCode:  flux.ErrorCodeGatewayInternal,
			Message:    "GATEWAY:UNKNOWN_PROTOCOL",
			Internal:   fmt.Errorf("unknown protocol:%s", service.RpcProto),
		}
	}
	return backend.Invoke(service, ctx)
}
