package ext

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/pkg"
)

var (
	protoBackendTransports = make(map[string]flux.BackendTransport, 4)
)

func SetBackendTransport(protoName string, backend flux.BackendTransport) {
	protoName = pkg.RequireNotEmpty(protoName, "protoName is empty")
	protoBackendTransports[protoName] = pkg.RequireNotNil(backend, "BackendTransport is nil").(flux.BackendTransport)
}

func GetBackendTransport(protoName string) (flux.BackendTransport, bool) {
	protoName = pkg.RequireNotEmpty(protoName, "protoName is empty")
	backend, ok := protoBackendTransports[protoName]
	return backend, ok
}

func GetBackendTransports() map[string]flux.BackendTransport {
	m := make(map[string]flux.BackendTransport, len(protoBackendTransports))
	for p, e := range protoBackendTransports {
		m[p] = e
	}
	return m
}
