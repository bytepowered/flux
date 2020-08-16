package server

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/webx"
	"github.com/labstack/gommon/random"
	https "net/http"
)

func (s *HttpServer) debugFeatures(config *flux.Configuration) {
	config.SetDefaults(map[string]interface{}{
		"debug-auth-username": "fluxgo",
		"debug-auth-password": random.String(8),
	})
	username := config.GetString("debug-auth-username")
	password := config.GetString("debug-auth-password")
	logger.Infow("Http debug feature: [ENABLED], Auth: BasicAuth", "username", username, "password", password)
	auth := webx.NewBasicAuthMiddleware(func(user string, pass string, webc webx.WebContext) (bool, error) {
		return user == username && pass == password, nil
	})
	// Debug查询接口
	debugHandler := webx.AdaptHttpHandler(https.DefaultServeMux)
	s.webServer.AddWebRouteHandler("GET", DebugPathVars, debugHandler, auth)
	s.webServer.AddWebRouteHandler("GET", DebugPathPprof, debugHandler, auth)
	// Endpoint查询
	json := ext.GetSerializer(ext.TypeNameSerializerJson)
	s.webServer.AddWebRouteHandler("GET", DebugPathEndpoints, func(webc webx.WebContext) error {
		if data, err := json.Marshal(queryEndpoints(s.mvEndpointMap, webc)); nil != err {
			return err
		} else {
			rw := webc.Response()
			webc.ResponseHeader().Set(webx.HeaderContentType, webx.MIMEApplicationJSON)
			_, err := rw.Write(data)
			if nil != err {
				rw.WriteHeader(https.StatusOK)
			}
			return err
		}
	}, auth)
}
