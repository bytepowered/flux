package ext

import (
	"github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-pkg"
)

var (
	transporters = make(map[string]flux.Transporter, 4)
)

func RegisterTransporter(protoName string, transporter flux.Transporter) {
	protoName = fluxpkg.MustNotEmpty(protoName, "protoName is empty")
	transporters[protoName] = fluxpkg.MustNotNil(transporter, "Transporter is nil").(flux.Transporter)
}

func TransporterBy(protoName string) (flux.Transporter, bool) {
	protoName = fluxpkg.MustNotEmpty(protoName, "protoName is empty")
	transporter, ok := transporters[protoName]
	return transporter, ok
}

func Transporters() map[string]flux.Transporter {
	m := make(map[string]flux.Transporter, len(transporters))
	for p, e := range transporters {
		m[p] = e
	}
	return m
}
