package ext

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/pkg"
)

var (
	_protoNamedBackends                = make(map[string]flux.Backend, 4)
	_protoNamedBackendResponseDecoders = make(map[string]flux.BackendResponseDecoder, 4)
)

func StoreBackend(protoName string, backend flux.Backend) {
	protoName = pkg.RequireNotEmpty(protoName, "protoName is empty")
	_protoNamedBackends[protoName] = pkg.RequireNotNil(backend, "Backend is nil").(flux.Backend)
}

func StoreBackendResponseDecoder(protoName string, decoder flux.BackendResponseDecoder) {
	protoName = pkg.RequireNotEmpty(protoName, "protoName is empty")
	_protoNamedBackendResponseDecoders[protoName] = pkg.RequireNotNil(decoder, "BackendResponseDecoder is nil").(flux.BackendResponseDecoder)
}

func LoadBackend(protoName string) (flux.Backend, bool) {
	protoName = pkg.RequireNotEmpty(protoName, "protoName is empty")
	backend, ok := _protoNamedBackends[protoName]
	return backend, ok
}

func LoadBackendResponseDecoder(protoName string) (flux.BackendResponseDecoder, bool) {
	protoName = pkg.RequireNotEmpty(protoName, "protoName is empty")
	decoder, ok := _protoNamedBackendResponseDecoders[protoName]
	return decoder, ok
}

func LoadBackends() map[string]flux.Backend {
	m := make(map[string]flux.Backend, len(_protoNamedBackends))
	for p, e := range _protoNamedBackends {
		m[p] = e
	}
	return m
}
