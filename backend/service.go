package backend

import (
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
)

func Exchange(ctx flux.Context, backend flux.BackendTransport) *flux.ServeError {
	endpoint := ctx.Endpoint()
	resp, ierr := backend.Invoke(endpoint.Service, ctx)
	if ierr != nil {
		return ierr
	}
	// decode response
	result, err := backend.GetResultDecodeFunc()(ctx, resp)
	if nil != err {
		return &flux.ServeError{
			StatusCode: flux.StatusServerError,
			ErrorCode:  flux.ErrorCodeGatewayInternal,
			Message:    flux.ErrorMessageBackendDecodeResponse,
			Internal:   err,
		}
	}
	writer := ctx.Response()
	writer.SetStatusCode(result.StatusCode)
	writer.SetHeaders(result.Headers)
	writer.SetBody(result.Body)
	// attachments
	for k, v := range result.Attachments {
		ctx.SetAttribute(k, v)
	}
	return nil
}

// Invoke 执行后端服务，获取响应结果；
func Invoke(service flux.BackendService, ctx flux.Context) (interface{}, *flux.ServeError) {
	rpcProto := service.AttrRpcProto()
	backend, ok := ext.LoadBackendTransport(rpcProto)
	if !ok {
		return nil, &flux.ServeError{
			StatusCode: flux.StatusServerError,
			ErrorCode:  flux.ErrorCodeGatewayInternal,
			Message:    "GATEWAY:UNKNOWN_PROTOCOL",
			Internal:   fmt.Errorf("unknown protocol:%s", rpcProto),
		}
	}
	return backend.Invoke(service, ctx)
}
