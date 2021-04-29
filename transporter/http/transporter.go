package http

import (
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/transporter"
	"github.com/spf13/cast"
	"io"
	"net/http"
	"net/url"
	"time"
)

func init() {
	ext.RegisterTransporter(flux.ProtoHttp, NewRpcHttpTransporter())
}

var _ flux.Transporter = new(RpcTransporter)

type (
	// Option 配置函数
	Option func(service *RpcTransporter)
	// ArgumentResolver Http调用参数封装函数，可外部化配置为其它协议的值对象
	ArgumentResolver func(service *flux.Service, inURL *url.URL, bodyReader io.ReadCloser, ctx *flux.Context) (*http.Request, error)
)

type RpcTransporter struct {
	httpClient  *http.Client
	codec       flux.TransportCodec
	writer      flux.TransportWriter
	argResolver ArgumentResolver
}

func (b *RpcTransporter) Writer() flux.TransportWriter {
	return b.writer
}

func NewRpcHttpTransporter() *RpcTransporter {
	return &RpcTransporter{
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
		codec:  NewTransportCodecFunc(),
		writer: new(transporter.DefaultTransportWriter),
	}
}

func NewRpcHttpTransporterWith(opts ...Option) *RpcTransporter {
	bts := &RpcTransporter{
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
	return func(s *RpcTransporter) {
		s.httpClient = client
	}
}

// WithTransportCodec 用于配置响应数据解析实现函数
func WithTransportCodec(fun flux.TransportCodec) Option {
	return func(service *RpcTransporter) {
		service.codec = fun
	}
}

// WithArgumentResolver 用于配置转发Http请求参数封装实现函数
func WithArgumentResolver(fun ArgumentResolver) Option {
	return func(service *RpcTransporter) {
		service.argResolver = fun
	}
}

// WithTransportWriter 用于配置响应数据解析实现函数
func WithTransportWriter(fun flux.TransportWriter) Option {
	return func(service *RpcTransporter) {
		service.writer = fun
	}
}

func (b *RpcTransporter) Transport(ctx *flux.Context) {
	transporter.DoTransport(ctx, b)
}

func (b *RpcTransporter) InvokeCodec(ctx *flux.Context, service flux.Service) (*flux.ResponseBody, *flux.ServeError) {
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
			Message:    flux.ErrorMessageTransportDecodeResponse,
			CauseError: fmt.Errorf("decode http response, err: %w", err),
		}
	}
	return result, nil
}

func (b *RpcTransporter) Invoke(ctx *flux.Context, service flux.Service) (interface{}, *flux.ServeError) {
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

func (b *RpcTransporter) ExecuteRequest(newRequest *http.Request, _ flux.Service, ctx *flux.Context) (interface{}, *flux.ServeError) {
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
			ErrorCode:  flux.ErrorCodeGatewayTransporter,
			Message:    msg,
			CauseError: err,
		}
	}
	return resp, nil
}
