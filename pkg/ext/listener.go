package ext

import (
	"github.com/bytepowered/fluxgo/pkg/flux"
)

var (
	webListenerFactory flux.WebListenerFactory
)

func SetWebListenerFactory(f flux.WebListenerFactory) {
	webListenerFactory = f
}

func WebListenerFactory() flux.WebListenerFactory {
	return webListenerFactory
}
