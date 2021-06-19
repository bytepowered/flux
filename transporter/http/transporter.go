package http

import (
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/spf13/cast"
	"net/http"
	"net/url"
	"time"
)

func init() {
	ext.RegisterTransporter(flux.ProtoHttp, NewTransporter())
}

var _ flux.Transporter = new(RpcTransporter)

type (
	// Option 配置函数
	Option func(service *RpcTransporter)
	// AssembleRequestFunc Http调用参数封装函数，可外部化配置为其它协议的值对象
	AssembleRequestFunc func(ctx *flux.Context, service *flux.Service) (*http.Request, error)
)

type RpcTransporter struct {
	httpClient      *http.Client
	codec           flux.TransportCodecFunc
	assembleRequest AssembleRequestFunc
}

func NewTransporter() *RpcTransporter {
	return &RpcTransporter{
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
		codec:           NewTransportCodecFunc(),
		assembleRequest: DefaultAssembleRequest,
	}
}

func NewTransporterWith(opts ...Option) *RpcTransporter {
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
func WithTransportCodec(fun flux.TransportCodecFunc) Option {
	return func(service *RpcTransporter) {
		service.codec = fun
	}
}

// WithAssembleRequest 用于配置转发Http请求参数封装实现函数
func WithAssembleRequest(fun AssembleRequestFunc) Option {
	return func(service *RpcTransporter) {
		service.assembleRequest = fun
	}
}

func (b *RpcTransporter) DoInvoke(ctx *flux.Context, service flux.Service) (*flux.ServeResponse, *flux.ServeError) {
	raw, serr := b.invoke0(ctx, service)
	if nil != serr {
		return nil, serr
	}
	// decode response
	result, err := b.codec(ctx, raw, make(map[string]interface{}, 0))
	if nil != err {
		return nil, &flux.ServeError{
			StatusCode: flux.StatusServerError,
			ErrorCode:  flux.ErrorCodeGatewayInternal,
			Message:    flux.ErrorMessageTransportCodecError,
			CauseError: fmt.Errorf("decode http response, err: %w", err),
		}
	}
	return result, nil
}

func (b *RpcTransporter) invoke0(ctx *flux.Context, service flux.Service) (interface{}, *flux.ServeError) {
	flux.AssertNotEmpty(service.Url, "<service.url> MUST NOT empty in http transporter")
	newRequest, err := b.assembleRequest(ctx, &service)
	if nil != err {
		return nil, &flux.ServeError{
			StatusCode: flux.StatusServerError,
			ErrorCode:  flux.ErrorCodeGatewayInternal,
			Message:    flux.ErrorMessageTransportHttpAssembleFailed,
			CauseError: err,
		}
	}
	return b.request(newRequest, service, ctx)
}

func (b *RpcTransporter) request(request *http.Request, _ flux.Service, ctx *flux.Context) (interface{}, *flux.ServeError) {
	// Header透传以及传递AttrValues
	request.Header = ctx.HeaderVars()
	for k, v := range ctx.Attributes() {
		request.Header.Set(k, cast.ToString(v))
	}
	resp, err := b.httpClient.Do(request)
	if nil != err {
		msg := flux.ErrorMessageTransportHttpInvokeFailed
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
