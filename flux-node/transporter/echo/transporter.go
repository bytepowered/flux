package echo

import (
	"github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-node/ext"
	"github.com/bytepowered/flux/flux-node/transporter"
	"io/ioutil"
	"net/http"
)

func init() {
	ext.RegisterTransporter(flux.ProtoEcho, NewTransporter())
}

var (
	_ flux.Transporter = new(RpcTransporter)
)

type RpcTransporter struct {
	codec  flux.TransportCodec
	writer flux.TransportWriter
}

func (b *RpcTransporter) Writer() flux.TransportWriter {
	return b.writer
}

func NewTransporter() flux.Transporter {
	return &RpcTransporter{
		codec:  NewTransportCodecFunc(),
		writer: new(transporter.DefaultTransportWriter),
	}
}

func (b *RpcTransporter) Transport(ctx *flux.Context) {
	transporter.DoTransport(ctx, b)
}

func (b *RpcTransporter) InvokeCodec(context *flux.Context, service flux.Service) (*flux.ResponseBody, *flux.ServeError) {
	resp, err := b.Invoke(context, service)
	if err != nil {
		return nil, err
	}
	codec, _ := b.codec(context, resp)
	return codec, nil
}

func (b *RpcTransporter) Invoke(ctx *flux.Context, service flux.Service) (interface{}, *flux.ServeError) {
	var data []byte
	if r, err := ctx.BodyReader(); nil == err {
		data, _ = ioutil.ReadAll(r)
		_ = r.Close()
	}
	header := ctx.HeaderVars()
	return map[string]interface{}{
		"service":  service,
		"request-id":           ctx.RequestId(),
		"request-uri":          ctx.URI(),
		"request-method":       ctx.Method(),
		"request-pathValues":   ctx.PathVars(),
		"request-queryValues":  ctx.QueryVars(),
		"request-formValues":   ctx.FormVars(),
		"request-headerValues": header,
		"request-body":         string(data),
	}, nil
}

func NewTransportCodecFunc() flux.TransportCodec {
	return func(ctx *flux.Context, value interface{}) (*flux.ResponseBody, error) {
		return &flux.ResponseBody{
			StatusCode: http.StatusOK,
			Headers:    make(http.Header, 0),
			Body:       value,
		}, nil
	}
}
