package ext

import (
	"github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-pkg"
)

var (
	protoBackendTransporters = make(map[string]flux.Transporter, 4)
)

func RegisterTransporter(protoName string, backend flux.Transporter) {
	protoName = fluxpkg.MustNotEmpty(protoName, "protoName is empty")
	protoBackendTransporters[protoName] = fluxpkg.MustNotNil(backend, "Transporter is nil").(flux.Transporter)
}

func BackendTransporterBy(protoName string) (flux.Transporter, bool) {
	protoName = fluxpkg.MustNotEmpty(protoName, "protoName is empty")
	backend, ok := protoBackendTransporters[protoName]
	return backend, ok
}

func BackendTransporters() map[string]flux.Transporter {
	m := make(map[string]flux.Transporter, len(protoBackendTransporters))
	for p, e := range protoBackendTransporters {
		m[p] = e
	}
	return m
}
