package http

import (
	"fmt"
	"github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-node/backend"
	"github.com/bytepowered/flux/flux-node/ext"
	"github.com/spf13/cast"
	"io"
	"net/http"
	"net/url"
	"time"
)

func init() {
	ext.RegisterBackendTransporter(flux.ProtoHttp, NewBackendTransporterService())
}

var _ flux.BackendTransporter = new(TransporterService)

type (
	// Option 配置函数
	Option func(service *TransporterService)
	// ArgumentsAssembleFunc Http调用参数封装函数，可外部化配置为其它协议的值对象
	ArgumentsAssembleFunc func(service *flux.TransporterService, inURL *url.URL, bodyReader io.ReadCloser, ctx *flux.Context) (*http.Request, error)
)

type TransporterService struct {
	httpClient        *http.Client
	responseCodecFunc flux.BackendCodecFunc
	argAssembleFunc   ArgumentsAssembleFunc
}

func NewBackendTransporterService() *TransporterService {
	return &TransporterService{
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
		responseCodecFunc: NewResponseCodecFunc(),
	}
}

func NewBackendTransporterServiceWith(opts ...Option) *TransporterService {
	bts := &TransporterService{
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
		responseCodecFunc: NewResponseCodecFunc(),
	}
	for _, opt := range opts {
		opt(bts)
	}
	return bts
}

// WithHttpClient 用于配置HttpClient客户端
func WithHttpClient(client *http.Client) Option {
	return func(s *TransporterService) {
		s.httpClient = client
	}
}

// WithResponseCodecFunc 用于配置响应数据解析实现函数
func WithResponseCodecFunc(fun flux.BackendCodecFunc) Option {
	return func(service *TransporterService) {
		service.responseCodecFunc = fun
	}
}

// WithArgumentAssembleFunc 用于配置转发Http请求参数封装实现函数
func WithArgumentAssembleFunc(fun ArgumentsAssembleFunc) Option {
	return func(service *TransporterService) {
		service.argAssembleFunc = fun
	}
}

func (b *TransporterService) GetBackendCodecFunc() flux.BackendCodecFunc {
	return b.responseCodecFunc
}

func (b *TransporterService) Transport(ctx *flux.Context) *flux.ServeError {
	return backend.DoTransport(ctx, b)
}

func (b *TransporterService) InvokeCodec(ctx *flux.Context, service flux.TransporterService) (*flux.BackendResponse, *flux.ServeError) {
	// panic("implement me")
	raw, serr := b.Invoke(ctx, service)
	if nil != serr {
		return nil, serr
	}
	// decode response
	result, err := b.GetBackendCodecFunc()(ctx, raw)
	if nil != err {
		return nil, &flux.ServeError{
			StatusCode: flux.StatusServerError,
			ErrorCode:  flux.ErrorCodeGatewayInternal,
			Message:    flux.ErrorMessageBackendDecodeResponse,
			CauseError: fmt.Errorf("decode http response, err: %w", err),
		}
	}
	return result, nil
}

func (b *TransporterService) Invoke(ctx *flux.Context, service flux.TransporterService) (interface{}, *flux.ServeError) {
	body, _ := ctx.BodyReader()
	newRequest, err := b.argAssembleFunc(&service, ctx.URL(), body, ctx)
	if nil != err {
		return nil, &flux.ServeError{
			StatusCode: flux.StatusServerError,
			ErrorCode:  flux.ErrorCodeGatewayInternal,
			Message:    flux.ErrorMessageHttpAssembleFailed,
			CauseError: err,
		}
	}
	return b.ExecuteRequest(newRequest, service, ctx)
}

func (b *TransporterService) ExecuteRequest(newRequest *http.Request, _ flux.TransporterService, ctx *flux.Context) (interface{}, *flux.ServeError) {
	// Header透传以及传递AttrValues
	newRequest.Header = ctx.HeaderVars()
	for k, v := range ctx.Attributes() {
		newRequest.Header.Set(k, cast.ToString(v))
	}
	resp, err := b.httpClient.Do(newRequest)
	if nil != err {
		msg := flux.ErrorMessageHttpInvokeFailed
		if uErr, ok := err.(*url.Error); ok {
			msg = fmt.Sprintf("HTTPEX:REMOTE_ERROR:%s", uErr.Error())
		}
		return nil, &flux.ServeError{
			StatusCode: flux.StatusServerError,
			ErrorCode:  flux.ErrorCodeGatewayBackend,
			Message:    msg,
			CauseError: err,
		}
	}
	return resp, nil
}
