package ext

import (
	"github.com/bytepowered/flux"
)

type WebServerFactory func() flux.WebServer

var _webServerFactory WebServerFactory

func StoreWebServerFactory(f WebServerFactory) {
	_webServerFactory = f
}

func LoadWebServerFactory() WebServerFactory {
	return _webServerFactory
}
