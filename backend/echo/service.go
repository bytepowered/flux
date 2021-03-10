package echo

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/backend"
	"github.com/bytepowered/flux/ext"
	"io/ioutil"
	"net/http"
)

func init() {
	ext.RegisterBackendTransport(flux.ProtoEcho, NewBackendTransportService())
}

var (
	_ flux.BackendTransport = new(BackendTransportService)
)

type BackendTransportService struct {
	decodeFunc flux.BackendResponseCodecFunc
}

func (b *BackendTransportService) GetResponseCodecFunc() flux.BackendResponseCodecFunc {
	return b.decodeFunc
}

func NewBackendTransportService() flux.BackendTransport {
	return &BackendTransportService{
		decodeFunc: NewBackendResponseCodecFunc(),
	}
}

func (b *BackendTransportService) Exchange(ctx flux.Context) *flux.ServeError {
	return backend.DoExchangeTransport(ctx, b)
}

func (b *BackendTransportService) InvokeCodec(context flux.Context, service flux.BackendService) (*flux.BackendResponse, *flux.ServeError) {
	resp, err := b.Invoke(context, service)
	if err != nil {
		return nil, err
	}
	codec, _ := b.decodeFunc(context, resp)
	return codec, nil
}

func (b *BackendTransportService) Invoke(ctx flux.Context, service flux.BackendService) (interface{}, *flux.ServeError) {
	var data []byte
	if r, err := ctx.Request().BodyReader(); nil == err {
		data, _ = ioutil.ReadAll(r)
		_ = r.Close()
	}
	header := ctx.Request().HeaderVars()
	return map[string]interface{}{
		"backend-service":      service,
		"request-id":           ctx.RequestId(),
		"request-uri":          ctx.URI(),
		"request-method":       ctx.Method(),
		"request-pathValues":   ctx.Request().PathVars(),
		"request-queryValues":  ctx.Request().QueryVars(),
		"request-formValues":   ctx.Request().FormVars(),
		"request-headerValues": header,
		"request-body":         string(data),
	}, nil
}

func NewBackendResponseCodecFunc() flux.BackendResponseCodecFunc {
	return func(ctx flux.Context, value interface{}) (*flux.BackendResponse, error) {
		return &flux.BackendResponse{
			StatusCode: http.StatusOK,
			Headers:    make(http.Header, 0),
			Body:       value,
		}, nil
	}
}
