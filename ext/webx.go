package ext

import (
	"github.com/bytepowered/flux"
)

type WebServerFactory func() flux.WebServer

var _webServerFactory WebServerFactory

func SetWebServerFactory(f WebServerFactory) {
	_webServerFactory = f
}

func GetWebServerFactory() WebServerFactory {
	return _webServerFactory
}
