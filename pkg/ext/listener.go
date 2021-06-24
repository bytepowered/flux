package ext

import (
	"github.com/bytepowered/fluxgo/pkg/flux"
)

var (
	webListenerFactory flux.WebListenerFactory
)

func SetWebListenerFactory(f flux.WebListenerFactory) {
	flux.AssertNotNil(f, "<web-listener-factory> must no nil")
	webListenerFactory = f
}

func WebListenerFactory() flux.WebListenerFactory {
	return webListenerFactory
}
