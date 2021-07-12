package ext

import (
	"github.com/bytepowered/fluxgo/pkg/flux"
)

var (
	transporters = make(map[string]flux.Transporter, 4)
)

func RegisterTransporter(proto string, transporter flux.Transporter) {
	proto = flux.MustNotEmpty(proto, "<proto-name> must not empty")
	transporters[proto] = flux.MustNotNil(transporter, "<transporter> must not nil").(flux.Transporter)
}

func TransporterByProto(proto string) (flux.Transporter, bool) {
	proto = flux.MustNotEmpty(proto, "<proto-name> must not empty")
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
