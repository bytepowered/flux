package http

import (
	"fmt"
	flux2 "github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-node/backend"
	"github.com/bytepowered/flux/flux-node/ext"
	"github.com/spf13/cast"
	"io"
	"net/http"
	"net/url"
	"time"
)

func init() {
	ext.RegisterBackendTransport(flux2.ProtoHttp, NewBackendTransportService())
}

var _ flux2.BackendTransport = new(BackendTransportService)

type (
	// Option 配置函数
	Option func(service *BackendTransportService)
	// ArgumentsAssembleFunc Http调用参数封装函数，可外部化配置为其它协议的值对象
	ArgumentsAssembleFunc func(service *flux2.BackendService, inURL *url.URL, bodyReader io.ReadCloser, ctx flux2.Context) (*http.Request, error)
)

type BackendTransportService struct {
	httpClient        *http.Client
	responseCodecFunc flux2.BackendResponseCodecFunc
	argAssembleFunc   ArgumentsAssembleFunc
}

func NewBackendTransportService() *BackendTransportService {
	return &BackendTransportService{
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
		responseCodecFunc: NewBackendResponseCodecFunc(),
	}
}

func NewBackendTransportServiceWith(opts ...Option) *BackendTransportService {
	bts := &BackendTransportService{
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
		responseCodecFunc: NewBackendResponseCodecFunc(),
	}
	for _, opt := range opts {
		opt(bts)
	}
	return bts
}

// WithHttpClient 用于配置HttpClient客户端
func WithHttpClient(client *http.Client) Option {
	return func(s *BackendTransportService) {
		s.httpClient = client
	}
}

// WithResponseCodecFunc 用于配置响应数据解析实现函数
func WithResponseCodecFunc(fun flux2.BackendResponseCodecFunc) Option {
	return func(service *BackendTransportService) {
		service.responseCodecFunc = fun
	}
}

// WithArgumentAssembleFunc 用于配置转发Http请求参数封装实现函数
func WithArgumentAssembleFunc(fun ArgumentsAssembleFunc) Option {
	return func(service *BackendTransportService) {
		service.argAssembleFunc = fun
	}
}

func (b *BackendTransportService) GetResponseCodecFunc() flux2.BackendResponseCodecFunc {
	return b.responseCodecFunc
}

func (b *BackendTransportService) Exchange(ctx flux2.Context) *flux2.ServeError {
	return backend.DoExchangeTransport(ctx, b)
}

func (b *BackendTransportService) InvokeCodec(ctx flux2.Context, service flux2.BackendService) (*flux2.BackendResponse, *flux2.ServeError) {
	// panic("implement me")
	raw, serr := b.Invoke(ctx, service)
	if nil != serr {
		return nil, serr
	}
	// decode response
	result, err := b.GetResponseCodecFunc()(ctx, raw)
	if nil != err {
		return nil, &flux2.ServeError{
			StatusCode: flux2.StatusServerError,
			ErrorCode:  flux2.ErrorCodeGatewayInternal,
			Message:    flux2.ErrorMessageBackendDecodeResponse,
			CauseError: fmt.Errorf("decode http response, err: %w", err),
		}
	}
	return result, nil
}

func (b *BackendTransportService) Invoke(ctx flux2.Context, service flux2.BackendService) (interface{}, *flux2.ServeError) {
	body, _ := ctx.Request().BodyReader()
	newRequest, err := b.argAssembleFunc(&service, ctx.Request().URL(), body, ctx)
	if nil != err {
		return nil, &flux2.ServeError{
			StatusCode: flux2.StatusServerError,
			ErrorCode:  flux2.ErrorCodeGatewayInternal,
			Message:    flux2.ErrorMessageHttpAssembleFailed,
			CauseError: err,
		}
	}
	return b.ExecuteRequest(newRequest, service, ctx)
}

func (b *BackendTransportService) ExecuteRequest(newRequest *http.Request, _ flux2.BackendService, ctx flux2.Context) (interface{}, *flux2.ServeError) {
	// Header透传以及传递AttrValues
	newRequest.Header = ctx.Request().HeaderVars()
	for k, v := range ctx.Attributes() {
		newRequest.Header.Set(k, cast.ToString(v))
	}
	resp, err := b.httpClient.Do(newRequest)
	if nil != err {
		msg := flux2.ErrorMessageHttpInvokeFailed
		if uErr, ok := err.(*url.Error); ok {
			msg = fmt.Sprintf("HTTPEX:REMOTE_ERROR:%s", uErr.Error())
		}
		return nil, &flux2.ServeError{
			StatusCode: flux2.StatusServerError,
			ErrorCode:  flux2.ErrorCodeGatewayBackend,
			Message:    msg,
			CauseError: err,
		}
	}
	return resp, nil
}
