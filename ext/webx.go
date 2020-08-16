package ext

import "github.com/bytepowered/flux/webx"

type WebServerFactory func() webx.WebServer

var _webServerFactory WebServerFactory

func SetWebServerFactory(f WebServerFactory) {
	_webServerFactory = f
}

func GetWebServerFactory() WebServerFactory {
	return _webServerFactory
}
