package ext

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/pkg"
)

var (
	_protoNamedBackends                = make(map[string]flux.Backend, 4)
	_protoNamedBackendResponseDecoders = make(map[string]flux.BackendResponseDecoder, 4)
)

func SetBackend(protoName string, backend flux.Backend) {
	_protoNamedBackends[protoName] = pkg.RequireNotNil(backend, "Backend is nil").(flux.Backend)
}

func SetBackendResponseDecoder(protoName string, decoder flux.BackendResponseDecoder) {
	_protoNamedBackendResponseDecoders[protoName] = pkg.RequireNotNil(decoder, "BackendResponseDecoder is nil").(flux.BackendResponseDecoder)
}

func GetBackend(protoName string) (flux.Backend, bool) {
	backend, ok := _protoNamedBackends[protoName]
	return backend, ok
}

func GetBackendResponseDecoder(protoName string) (flux.BackendResponseDecoder, bool) {
	decoder, ok := _protoNamedBackendResponseDecoders[protoName]
	return decoder, ok
}

func Backends() map[string]flux.Backend {
	m := make(map[string]flux.Backend, len(_protoNamedBackends))
	for p, e := range _protoNamedBackends {
		m[p] = e
	}
	return m
}
