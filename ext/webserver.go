package ext

import (
	"github.com/bytepowered/flux"
)

var (
	webServerFactory WebServerFactory
)

type WebServerFactory func(*flux.Configuration) flux.ListenServer

func SetWebServerFactory(f WebServerFactory) {
	webServerFactory = f
}

func GetWebServerFactory() WebServerFactory {
	return webServerFactory
}
