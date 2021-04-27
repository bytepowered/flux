package ext

import (
	"github.com/bytepowered/flux/flux-node"
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
