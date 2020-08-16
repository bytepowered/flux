package server

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/webex"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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
	auth := middleware.BasicAuth(func(u string, p string, c webex.WebContext) (bool, error) {
		return u == username && p == password, nil
	})
	debugHandler := echo.WrapHandler(https.DefaultServeMux)
	s.webserver.AddRouteHandler("GET", DebugPathVars, debugHandler, auth)
	s.webserver.AddRouteHandler("GET", DebugPathPprof, debugHandler, auth)
	s.webserver.AddRouteHandler("GET", DebugPathEndpoints, func(c echo.Context) error {
		if data, err := ext.GetSerializer(ext.TypeNameSerializerJson).Marshal(queryEndpoints(s.mvEndpointMap, c)); nil != err {
			return err
		} else {
			return c.JSONBlob(flux.StatusOK, data)
		}
	}, auth)
}
