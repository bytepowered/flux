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
	ext.RegisterTransporter(flux.ProtoHttp, NewRpcHttpTransporter())
}

var _ flux.Transporter = new(RpcHttpTransporter)

type (
	// Option 配置函数
	Option func(service *RpcHttpTransporter)
	// ArgumentResolver Http调用参数封装函数，可外部化配置为其它协议的值对象
	ArgumentResolver func(service *flux.TransporterService, inURL *url.URL, bodyReader io.ReadCloser, ctx *flux.Context) (*http.Request, error)
)

type RpcHttpTransporter struct {
	httpClient  *http.Client
	codec       flux.TransportCodec
	writer      flux.TransportWriter
	argResolver ArgumentResolver
}

func (b *RpcHttpTransporter) Writer() flux.TransportWriter {
	return b.writer
}

func NewRpcHttpTransporter() *RpcHttpTransporter {
	return &RpcHttpTransporter{
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
		codec:  NewTransportCodecFunc(),
		writer: new(backend.RpcTransportWriter),
	}
}

func NewRpcHttpTransporterWith(opts ...Option) *RpcHttpTransporter {
	bts := &RpcHttpTransporter{
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
		codec: NewTransportCodecFunc(),
	}
	for _, opt := range opts {
		opt(bts)
	}
	return bts
}

// WithHttpClient 用于配置HttpClient客户端
func WithHttpClient(client *http.Client) Option {
	return func(s *RpcHttpTransporter) {
		s.httpClient = client
	}
}

// WithTransportCodec 用于配置响应数据解析实现函数
func WithTransportCodec(fun flux.TransportCodec) Option {
	return func(service *RpcHttpTransporter) {
		service.codec = fun
	}
}

// WithArgumentResolver 用于配置转发Http请求参数封装实现函数
func WithArgumentResolver(fun ArgumentResolver) Option {
	return func(service *RpcHttpTransporter) {
		service.argResolver = fun
	}
}

// WithTransportWriter 用于配置响应数据解析实现函数
func WithTransportWriter(fun flux.TransportWriter) Option {
	return func(service *RpcHttpTransporter) {
		service.writer = fun
	}
}

func (b *RpcHttpTransporter) Transport(ctx *flux.Context) {
	backend.DoTransport(ctx, b)
}

func (b *RpcHttpTransporter) InvokeCodec(ctx *flux.Context, service flux.TransporterService) (*flux.ResponseBody, *flux.ServeError) {
	raw, serr := b.Invoke(ctx, service)
	if nil != serr {
		return nil, serr
	}
	// decode response
	result, err := b.codec(ctx, raw)
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

func (b *RpcHttpTransporter) Invoke(ctx *flux.Context, service flux.TransporterService) (interface{}, *flux.ServeError) {
	body, _ := ctx.BodyReader()
	newRequest, err := b.argResolver(&service, ctx.URL(), body, ctx)
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

func (b *RpcHttpTransporter) ExecuteRequest(newRequest *http.Request, _ flux.TransporterService, ctx *flux.Context) (interface{}, *flux.ServeError) {
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
