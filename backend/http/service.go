package http

import (
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/backend"
	"github.com/bytepowered/flux/ext"
	"github.com/spf13/cast"
	"io"
	"net/http"
	"net/url"
	"time"
)

func init() {
	ext.StoreBackendTransport(flux.ProtoHttp, NewBackendTransportService())
}

var _ flux.BackendTransport = new(BackendTransportService)

type (
	// Option 配置函数
	Option func(service *BackendTransportService)
	// ArgumentsAssembleFunc Http调用参数封装函数，可外部化配置为其它协议的值对象
	ArgumentsAssembleFunc func(service *flux.BackendService, inURL *url.URL, bodyReader io.ReadCloser, ctx flux.Context) (*http.Request, error)
)

type BackendTransportService struct {
	httpClient       *http.Client
	resultDecodeFunc flux.BackendResultDecodeFunc
	argAssembleFunc  ArgumentsAssembleFunc
}

func NewBackendTransportService() *BackendTransportService {
	return &BackendTransportService{
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
		resultDecodeFunc: NewBackendResultDecodeFunc(),
	}
}

func NewBackendTransportServiceWith(opts ...Option) *BackendTransportService {
	bts := &BackendTransportService{
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
		resultDecodeFunc: NewBackendResultDecodeFunc(),
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

// WithResultDecodeFunc 用于配置响应数据解析实现函数
func WithResultDecodeFunc(fun flux.BackendResultDecodeFunc) Option {
	return func(service *BackendTransportService) {
		service.resultDecodeFunc = fun
	}
}

// WithArgumentAssembleFunc 用于配置转发Http请求参数封装实现函数
func WithArgumentAssembleFunc(fun ArgumentsAssembleFunc) Option {
	return func(service *BackendTransportService) {
		service.argAssembleFunc = fun
	}
}

func (ex *BackendTransportService) GetResultDecodeFunc() flux.BackendResultDecodeFunc {
	return ex.resultDecodeFunc
}

func (ex *BackendTransportService) Exchange(ctx flux.Context) *flux.ServeError {
	return backend.Exchange(ctx, ex)
}

func (ex *BackendTransportService) Invoke(service flux.BackendService, ctx flux.Context) (interface{}, *flux.ServeError) {
	inurl, _ := ctx.Request().RequestURL()
	body, _ := ctx.Request().RequestBodyReader()
	newRequest, err := ex.argAssembleFunc(&service, inurl, body, ctx)
	if nil != err {
		return nil, &flux.ServeError{
			StatusCode: flux.StatusServerError,
			ErrorCode:  flux.ErrorCodeGatewayInternal,
			Message:    flux.ErrorMessageHttpAssembleFailed,
			Internal:   err,
		}
	}
	return ex.ExecuteRequest(newRequest, service, ctx)
}

func (ex *BackendTransportService) ExecuteRequest(newRequest *http.Request, _ flux.BackendService, ctx flux.Context) (interface{}, *flux.ServeError) {
	// Header透传以及传递AttrValues
	if header, writable := ctx.Request().HeaderValues(); writable {
		newRequest.Header = header.Clone()
	} else {
		newRequest.Header = header
	}
	for k, v := range ctx.Attributes() {
		newRequest.Header.Set(k, cast.ToString(v))
	}
	resp, err := ex.httpClient.Do(newRequest)
	if nil != err {
		msg := flux.ErrorMessageHttpInvokeFailed
		if uErr, ok := err.(*url.Error); ok {
			msg = fmt.Sprintf("HTTPEX:REMOTE_ERROR:%s", uErr.Error())
		}
		return nil, &flux.ServeError{
			StatusCode: flux.StatusServerError,
			ErrorCode:  flux.ErrorCodeGatewayBackend,
			Message:    msg,
			Internal:   err,
		}
	}
	return resp, nil
}
