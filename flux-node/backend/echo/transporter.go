package echo

import (
	"github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-node/backend"
	"github.com/bytepowered/flux/flux-node/ext"
	"io/ioutil"
	"net/http"
)

func init() {
	ext.RegisterTransporter(flux.ProtoEcho, NewTransporter())
}

var (
	_ flux.Transporter = new(RpcEchoTransporter)
)

type RpcEchoTransporter struct {
	codec  flux.TransportCodec
	writer flux.TransportWriter
}

func (b *RpcEchoTransporter) Writer() flux.TransportWriter {
	return b.writer
}

func NewTransporter() flux.Transporter {
	return &RpcEchoTransporter{
		codec:  NewTransportCodecFunc(),
		writer: new(backend.RpcTransportWriter),
	}
}

func (b *RpcEchoTransporter) Transport(ctx *flux.Context) {
	backend.DoTransport(ctx, b)
}

func (b *RpcEchoTransporter) InvokeCodec(context *flux.Context, service flux.TransporterService) (*flux.ResponseBody, *flux.ServeError) {
	resp, err := b.Invoke(context, service)
	if err != nil {
		return nil, err
	}
	codec, _ := b.codec(context, resp)
	return codec, nil
}

func (b *RpcEchoTransporter) Invoke(ctx *flux.Context, service flux.TransporterService) (interface{}, *flux.ServeError) {
	var data []byte
	if r, err := ctx.BodyReader(); nil == err {
		data, _ = ioutil.ReadAll(r)
		_ = r.Close()
	}
	header := ctx.HeaderVars()
	return map[string]interface{}{
		"backend-service":      service,
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
