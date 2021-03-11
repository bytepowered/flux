package ext

import (
	flux2 "github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-pkg"
)

var (
	protoBackendTransports = make(map[string]flux2.BackendTransport, 4)
)

func RegisterBackendTransport(protoName string, backend flux2.BackendTransport) {
	protoName = fluxpkg.MustNotEmpty(protoName, "protoName is empty")
	protoBackendTransports[protoName] = fluxpkg.MustNotNil(backend, "BackendTransport is nil").(flux2.BackendTransport)
}

func BackendTransportByProto(protoName string) (flux2.BackendTransport, bool) {
	protoName = fluxpkg.MustNotEmpty(protoName, "protoName is empty")
	backend, ok := protoBackendTransports[protoName]
	return backend, ok
}

func BackendTransports() map[string]flux2.BackendTransport {
	m := make(map[string]flux2.BackendTransport, len(protoBackendTransports))
	for p, e := range protoBackendTransports {
		m[p] = e
	}
	return m
}
