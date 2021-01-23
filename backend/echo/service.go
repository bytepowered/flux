package echo

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/backend"
	"github.com/bytepowered/flux/ext"
	"io/ioutil"
	"net/http"
)

func init() {
	ext.StoreBackendTransport(flux.ProtoEcho, NewBackendTransportService())
}

var (
	_ flux.BackendTransport = new(BackendTransportService)
)

type BackendTransportService struct {
	decodeFunc flux.BackendResultDecodeFunc
}

func (b *BackendTransportService) GetResultDecodeFunc() flux.BackendResultDecodeFunc {
	return b.decodeFunc
}

func NewBackendTransportService() flux.BackendTransport {
	return &BackendTransportService{
		decodeFunc: NewEchoBackendResultDecodeFunc(),
	}
}

func (b *BackendTransportService) Exchange(ctx flux.Context) *flux.ServeError {
	return backend.Exchange(ctx, b)
}

func (b *BackendTransportService) Invoke(service flux.BackendService, ctx flux.Context) (interface{}, *flux.ServeError) {
	var data []byte
	if r, err := ctx.Request().RequestBodyReader(); nil == err {
		data, _ = ioutil.ReadAll(r)
		_ = r.Close()
	}
	header, _ := ctx.Request().HeaderValues()
	return map[string]interface{}{
		"backend-service":      service,
		"request-id":           ctx.RequestId(),
		"request-uri":          ctx.RequestURI(),
		"request-method":       ctx.Method(),
		"request-pathValues":   ctx.Request().PathValues(),
		"request-queryValues":  ctx.Request().QueryValues(),
		"request-formValues":   ctx.Request().FormValues(),
		"request-headerValues": header,
		"request-body":         string(data),
	}, nil
}

func NewEchoBackendResultDecodeFunc() flux.BackendResultDecodeFunc {
	return func(ctx flux.Context, value interface{}) (*flux.BackendResult, error) {
		return &flux.BackendResult{
			StatusCode: http.StatusOK,
			Headers:    make(http.Header, 0),
			Body:       value,
		}, nil
	}
}
