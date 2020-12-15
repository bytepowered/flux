package http

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
)

func init() {
	// Backends
	ext.StoreBackendTransport(flux.ProtoHttp, NewHttpBackendTransport())
	ext.StoreBackendTransportDecodeFunc(flux.ProtoHttp, NewHttpBackendTransportDecodeFunc())
}
