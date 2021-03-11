package echo

import (
	flux2 "github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-node/backend"
	"github.com/bytepowered/flux/flux-node/ext"
	"io/ioutil"
	"net/http"
)

func init() {
	ext.RegisterBackendTransport(flux2.ProtoEcho, NewBackendTransportService())
}

var (
	_ flux2.BackendTransport = new(BackendTransportService)
)

type BackendTransportService struct {
	decodeFunc flux2.BackendResponseCodecFunc
}

func (b *BackendTransportService) GetResponseCodecFunc() flux2.BackendResponseCodecFunc {
	return b.decodeFunc
}

func NewBackendTransportService() flux2.BackendTransport {
	return &BackendTransportService{
		decodeFunc: NewBackendResponseCodecFunc(),
	}
}

func (b *BackendTransportService) Exchange(ctx flux2.Context) *flux2.ServeError {
	return backend.DoExchangeTransport(ctx, b)
}

func (b *BackendTransportService) InvokeCodec(context flux2.Context, service flux2.BackendService) (*flux2.BackendResponse, *flux2.ServeError) {
	resp, err := b.Invoke(context, service)
	if err != nil {
		return nil, err
	}
	codec, _ := b.decodeFunc(context, resp)
	return codec, nil
}

func (b *BackendTransportService) Invoke(ctx flux2.Context, service flux2.BackendService) (interface{}, *flux2.ServeError) {
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

func NewBackendResponseCodecFunc() flux2.BackendResponseCodecFunc {
	return func(ctx flux2.Context, value interface{}) (*flux2.BackendResponse, error) {
		return &flux2.BackendResponse{
			StatusCode: http.StatusOK,
			Headers:    make(http.Header, 0),
			Body:       value,
		}, nil
	}
}
