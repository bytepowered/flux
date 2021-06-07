package inapp

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"net/http"
)

func init() {
	ext.RegisterTransporter(flux.ProtoInApp, NewTransporter())
}

var (
	_ flux.Transporter = new(Transporter)
)

type Transporter struct {
	codec flux.TransportCodecFunc
}

type InvokeFunc func(context *flux.Context, service flux.Service) (interface{}, *flux.ServeError)

func NewTransporter() flux.Transporter {
	return &Transporter{
		codec: NewTransportCodecFunc(),
	}
}

func (b *Transporter) DoInvoke(context *flux.Context, service flux.Service) (*flux.ServeResponse, *flux.ServeError) {
	fun, ok := LoadInvokeFunc(service.ServiceID())
	if !ok {
		return nil, &flux.ServeError{
			StatusCode: flux.StatusServerError,
			ErrorCode:  flux.ErrorCodeGatewayInternal,
			Message:    "TRANSPORTER:INAPP:INVOKER/notfound",
		}
	}
	resp, err := fun(context, service)
	if err != nil {
		return nil, err
	}
	codec, _ := b.codec(context, resp)
	return codec, nil
}

func NewTransportCodecFunc() flux.TransportCodecFunc {
	return func(ctx *flux.Context, value interface{}) (*flux.ServeResponse, error) {
		return &flux.ServeResponse{
			StatusCode: http.StatusOK,
			Headers:    make(http.Header, 0),
			Body:       value,
		}, nil
	}
}
