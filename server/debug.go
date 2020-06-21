package server

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/random"
	https "net/http"
)

func (fs *FluxServer) debugFeatures(config *flux.Configuration) {
	config.SetDefaults(map[string]interface{}{
		"debug-auth-username": "fluxgo",
		"debug-auth-password": random.String(8),
	})
	username := config.GetString("debug-auth-username")
	password := config.GetString("debug-auth-password")
	logger.Infow("Http debug feature: [ENABLED], Auth: BasicAuth", "username", username, "password", password)
	auth := middleware.BasicAuth(func(u string, p string, c echo.Context) (bool, error) {
		return u == username && p == password, nil
	})
	debugHandler := echo.WrapHandler(https.DefaultServeMux)
	fs.httpServer.GET(DebugPathVars, debugHandler, auth)
	fs.httpServer.GET(DebugPathPprof, debugHandler, auth)
	fs.httpServer.GET(DebugPathEndpoints, func(c echo.Context) error {
		decoder := ext.GetSerializer(ext.TypeNameSerializerJson)
		if data, err := decoder.Marshal(queryEndpoints(fs.endpointMvMap, c)); nil != err {
			return err
		} else {
			return c.JSONBlob(flux.StatusOK, data)
		}
	}, auth)
}
