package echo

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
)

func init() {
	ext.StoreBackendTransport(flux.ProtoEcho, NewEchoBackendTransport())
	ext.StoreBackendTransportDecodeFunc(flux.ProtoEcho, NewEchoBackendTransportDecodeFunc())
}
