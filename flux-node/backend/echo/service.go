package echo

import (
	"github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-node/backend"
	"github.com/bytepowered/flux/flux-node/ext"
	"io/ioutil"
	"net/http"
)

func init() {
	ext.RegisterBackendTransporter(flux.ProtoEcho, NewBackendTransportService())
}

var (
	_ flux.BackendTransporter = new(TransportService)
)

type TransportService struct {
	decodeFunc flux.BackendCodecFunc
}

func NewBackendTransportService() flux.BackendTransporter {
	return &TransportService{
		decodeFunc: NewResponseCodecFunc(),
	}
}

func (b *TransportService) Transport(ctx *flux.Context) *flux.ServeError {
	return backend.DoTransport(ctx, b)
}

func (b *TransportService) InvokeCodec(context *flux.Context, service flux.TransporterService) (*flux.BackendResponse, *flux.ServeError) {
	resp, err := b.Invoke(context, service)
	if err != nil {
		return nil, err
	}
	codec, _ := b.decodeFunc(context, resp)
	return codec, nil
}

func (b *TransportService) Invoke(ctx *flux.Context, service flux.TransporterService) (interface{}, *flux.ServeError) {
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

func NewResponseCodecFunc() flux.BackendCodecFunc {
	return func(ctx *flux.Context, value interface{}) (*flux.BackendResponse, error) {
		return &flux.BackendResponse{
			StatusCode: http.StatusOK,
			Headers:    make(http.Header, 0),
			Body:       value,
		}, nil
	}
}
