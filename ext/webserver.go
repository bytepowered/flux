package ext

import (
	"github.com/bytepowered/flux"
)

var (
	webServerFactory WebServerFactory
)

type WebServerFactory func(*flux.Configuration) flux.WebServer

func StoreWebServerFactory(f WebServerFactory) {
	webServerFactory = f
}

func LoadWebServerFactory() WebServerFactory {
	return webServerFactory
}
