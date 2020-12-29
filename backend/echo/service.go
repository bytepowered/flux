package echo

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/backend"
	"io/ioutil"
	"net/http"
)

var (
	_ flux.BackendTransport = new(BackendTransportService)
)

type BackendTransportService struct {
}

func NewEchoBackendTransport() flux.BackendTransport {
	return &BackendTransportService{}
}

func (b *BackendTransportService) Exchange(ctx flux.Context) *flux.ServeError {
	return backend.DoExchange(ctx, b)
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

func NewEchoBackendTransportDecodeFunc() flux.BackendTransportDecodeFunc {
	return func(ctx flux.Context, value interface{}) (statusCode int, headers http.Header, body interface{}, err error) {
		return http.StatusOK, http.Header{}, value, nil
	}
}
