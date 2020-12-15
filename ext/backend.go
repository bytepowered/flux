package ext

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/pkg"
)

var (
	_protoNamedBackendTransports   = make(map[string]flux.BackendTransport, 4)
	_protoNamedBackendDecoderFuncs = make(map[string]flux.BackendTransportDecodeFunc, 4)
)

func StoreBackendTransport(protoName string, backend flux.BackendTransport) {
	protoName = pkg.RequireNotEmpty(protoName, "protoName is empty")
	_protoNamedBackendTransports[protoName] = pkg.RequireNotNil(backend, "BackendTransport is nil").(flux.BackendTransport)
}

func LoadBackendTransport(protoName string) (flux.BackendTransport, bool) {
	protoName = pkg.RequireNotEmpty(protoName, "protoName is empty")
	backend, ok := _protoNamedBackendTransports[protoName]
	return backend, ok
}

func StoreBackendTransportDecodeFunc(protoName string, decoder flux.BackendTransportDecodeFunc) {
	protoName = pkg.RequireNotEmpty(protoName, "protoName is empty")
	_protoNamedBackendDecoderFuncs[protoName] = pkg.RequireNotNil(decoder, "BackendTransportDecodeFunc is nil").(flux.BackendTransportDecodeFunc)
}

func LoadBackendTransportDecodeFunc(protoName string) (flux.BackendTransportDecodeFunc, bool) {
	protoName = pkg.RequireNotEmpty(protoName, "protoName is empty")
	decoder, ok := _protoNamedBackendDecoderFuncs[protoName]
	return decoder, ok
}

func LoadBackendTransports() map[string]flux.BackendTransport {
	m := make(map[string]flux.BackendTransport, len(_protoNamedBackendTransports))
	for p, e := range _protoNamedBackendTransports {
		m[p] = e
	}
	return m
}
