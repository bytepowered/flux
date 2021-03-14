package ext

import (
	"github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-pkg"
)

var (
	protoBackendTransporters = make(map[string]flux.BackendTransporter, 4)
)

func RegisterBackendTransporter(protoName string, backend flux.BackendTransporter) {
	protoName = fluxpkg.MustNotEmpty(protoName, "protoName is empty")
	protoBackendTransporters[protoName] = fluxpkg.MustNotNil(backend, "BackendTransporter is nil").(flux.BackendTransporter)
}

func BackendTransporterBy(protoName string) (flux.BackendTransporter, bool) {
	protoName = fluxpkg.MustNotEmpty(protoName, "protoName is empty")
	backend, ok := protoBackendTransporters[protoName]
	return backend, ok
}

func BackendTransporters() map[string]flux.BackendTransporter {
	m := make(map[string]flux.BackendTransporter, len(protoBackendTransporters))
	for p, e := range protoBackendTransporters {
		m[p] = e
	}
	return m
}
