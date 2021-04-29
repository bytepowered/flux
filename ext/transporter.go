package ext

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/fluxkit"
)

var (
	transporters = make(map[string]flux.Transporter, 4)
)

func RegisterTransporter(proto string, transporter flux.Transporter) {
	proto = fluxkit.MustNotEmpty(proto, "protoName is empty")
	transporters[proto] = fluxkit.MustNotNil(transporter, "Transporter is nil").(flux.Transporter)
}

func TransporterByProto(proto string) (flux.Transporter, bool) {
	proto = fluxkit.MustNotEmpty(proto, "protoName is empty")
	transporter, ok := transporters[proto]
	return transporter, ok
}

func Transporters() map[string]flux.Transporter {
	m := make(map[string]flux.Transporter, len(transporters))
	for p, e := range transporters {
		m[p] = e
	}
	return m
}
